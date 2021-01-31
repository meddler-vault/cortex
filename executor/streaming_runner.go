package executor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

// FunctionRunner runs a function
type FunctionRunner interface {
	Run(f FunctionRequest) error
}

// FunctionRequest stores request for function execution
type FunctionRequest struct {
	Process     string
	ProcessArgs []string
	Environment []string

	InputReader             io.ReadCloser
	OutputWriter            io.Writer
	ContentLength           *int64
	TractID                 string
	CurrentWorkingDirectory string
}

// ForkFunctionRunner forks a process for each invocation
type ForkFunctionRunner struct {
	ExecTimeout time.Duration
}

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func getenvInt(key string, defaultValue int) int {
	s := getenvStr(key, strconv.Itoa(defaultValue))

	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

var logger, _ = fluent.New(fluent.Config{
	FluentPort:   getenvInt("fluent_port", 24224),
	FluentHost:   getenvStr("fluent_host", "localhost"),
	TagPrefix:    "watchdog",
	MaxRetryWait: 4,
	// Async:        false,
	MaxRetry: 0,
})

// Run run a fork for each invocation
func (f *ForkFunctionRunner) Run(req FunctionRequest) error {
	log.Println("Running ", req.Process, req.ProcessArgs, req.Environment)
	log.Println("fluentd debug", getenvInt("fluent_port", 24224), getenvStr("fluent_host", "localhost"))
	start := time.Now()
	cmd := exec.Command(req.Process, req.ProcessArgs...)
	cmd.Env = req.Environment
	// cmd.Env = append(cmd.Env, os.Environ()...) //Load & Curren Env From Docker File via current process
	// log.Println("EnvVaribles", cmd.Env)
	cmd.Dir = req.CurrentWorkingDirectory

	var timer *time.Timer
	if f.ExecTimeout > time.Millisecond*0 {
		timer = time.NewTimer(f.ExecTimeout)

		go func() {
			<-timer.C

			log.Printf("Function was killed by ExecTimeout: %s\n", f.ExecTimeout.String())
			killErr := cmd.Process.Kill()
			if killErr != nil {
				fmt.Println("Error killing function due to ExecTimeout", killErr)
			}
		}()
	}

	if timer != nil {
		defer timer.Stop()
	}

	if req.InputReader != nil {
		defer req.InputReader.Close()
		cmd.Stdin = req.InputReader
	}
	// cmd.Stdout = req.OutputWriter

	// Prints stderr to console and is picked up by container logging driver.
	errPipe, _ := cmd.StderrPipe()
	stdoutPipe, _ := cmd.StdoutPipe()
	// log.Printf("TractId", req.TractID)

	var wg sync.WaitGroup
	bindFluentLoggingPipe(logger, "stderr", req.TractID, errPipe, &wg)
	bindFluentLoggingPipe(logger, "stdout", req.TractID, stdoutPipe, &wg)

	startErr := cmd.Start()
	wg.Wait()

	if startErr != nil {
		log.Println("Starting error", startErr)

		logger.Post(req.TractID, map[string]string{
			"pipe":    "stdend",
			"message": "End: " + startErr.Error(),
		})
		return startErr
	}

	logger.Post(req.TractID, map[string]string{
		"pipe":    "stdend",
		"message": "End: " + "Process completed successfully",
	})

	waitErr := cmd.Wait()
	done := time.Since(start)
	log.Printf("Took %f secs", done.Seconds())
	if timer != nil {
		timer.Stop()
	}

	req.InputReader.Close()

	req.OutputWriter.Write([]byte("Trace-ID: " + req.TractID + strconv.Itoa(getenvInt("fluent_port", 24224)) + getenvStr("fluent_host", "localhost")))

	if waitErr != nil {
		return waitErr
	}

	return nil
}
