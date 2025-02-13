package executor

import (
	"fmt"
	"io"

	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	_logger "github.com/meddler-vault/cortex/logger"

	"github.com/fluent/fluent-logger-golang/fluent"
)

// FunctionRunner runs a function
type FunctionRunner interface {
	Run(f FunctionRequest) error
}

// FunctionRequest stores request for function execution
type FunctionRequest struct {
	Process                 string
	ProcessArgs             []string
	Environment             []string
	GracefullProcessKiller  <-chan string
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
func (f *ForkFunctionRunner) Run(req FunctionRequest) (map[string]interface{}, error) {

	var meta_data = make(map[string]interface{})

	_logger.Println("Running ", req.Process, req.ProcessArgs, req.Environment)
	_logger.Println("fluentd debug", getenvInt("fluent_port", 24224), getenvStr("fluent_host", "localhost"))
	start := time.Now()
	cmd := exec.Command(req.Process, req.ProcessArgs...)
	// TODO: Review Killing all process goups
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	//
	cmd.Env = req.Environment
	cmd.Env = append(os.Environ(), cmd.Env...) //Load & Curren Env From Docker File via current process
	// _logger.Println("EnvVaribles", cmd.Env)
	cmd.Dir = req.CurrentWorkingDirectory

	var timer *time.Timer
	if f.ExecTimeout > time.Millisecond*0 {
		timer = time.NewTimer(f.ExecTimeout)
		meta_data["exec_stop_timeout_attempt"] = false

		_logger.Println("PM: Starting Process Killer Timeout", f.ExecTimeout)

		go func() {
			_logger.Println("PM: Started Process Killer Timeout", f.ExecTimeout)

			<-timer.C

			meta_data["exec_stop_timeout_attempt"] = true
			_logger.Println("Function will be killed by ExecTimeout:", f.ExecTimeout.String())

			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err != nil {
				meta_data["exec_stop_timeout_status"] = false

				_logger.Println("Kill Signal Failed: coudnb;t get process group_id")
				_logger.Println("Error", err)
				return
			}

			// killErr := cmd.Process.Kill()
			killErr := syscall.Kill(-pgid, syscall.SIGKILL)
			if killErr != nil {
				meta_data["exec_stop_timeout_status"] = false
				fmt.Println("Error killing function due to ExecTimeout", killErr)
			}

			_logger.Println("Kill Signal Sent")

			killErr = cmd.Wait()
			if killErr != nil {
				meta_data["exec_stop_timeout_status"] = false
				fmt.Println("Error waiting function due to ExecTimeout", killErr)
			}

			meta_data["exec_stop_timeout_status"] = true

			_logger.Println("Successully Killed")

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
	// _logger.Printf("TractId", req.TractID)

	var wg sync.WaitGroup
	bindFluentLoggingPipe(logger, "stderr", req.TractID, errPipe, &wg)
	bindFluentLoggingPipe(logger, "stdout", req.TractID, stdoutPipe, &wg)

	startErr := cmd.Start()
	wg.Wait()

	if startErr != nil {
		meta_data["startup_status"] = false
		_logger.Println("Starting error", startErr, "path", cmd.Path, "lookPathErr")

		logger.Post(req.TractID, map[string]string{
			"pipe":    "stdend",
			"message": "End: " + startErr.Error(),
		})
		return meta_data, startErr
	}

	go func() {
		meta_data["exec_stop_gracefull_attempt"] = false

		_logger.Println("gracefull-process Killer Activated")
		<-req.GracefullProcessKiller
		_logger.Println("gracefull-process Killer Executing")

		meta_data["exec_stop_gracefull_attempt"] = true
		_logger.Println("Function will be killed by ExecTimeout:", f.ExecTimeout.String())

		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err != nil {
			meta_data["exec_stop_gracefull_status"] = false

			_logger.Println("Kill Signal Failed: coudnb;t get process group_id")
			_logger.Println("Error", err)
			return
		}

		// killErr := cmd.Process.Kill()
		killErr := syscall.Kill(-pgid, syscall.SIGKILL)
		if killErr != nil {
			meta_data["exec_stop_gracefull_status"] = false
			fmt.Println("Error killing function due to ExecTimeout", killErr)
		}

		_logger.Println("Kill Signal Sent")

		killErr = cmd.Wait()
		if killErr != nil {
			meta_data["exec_stop_gracefull_status"] = false
			fmt.Println("Error waiting function due to ExecTimeout", killErr)
		}

		meta_data["exec_stop_gracefull_status"] = true

		_logger.Println("Successully Killed")

	}()

	logger.Post(req.TractID, map[string]string{
		"pipe":    "stdend",
		"message": "End: " + "Process completed successfully",
	})

	waitErr := cmd.Wait()

	meta_data["exit_code"] = cmd.ProcessState.ExitCode()
	meta_data["exit_message"] = cmd.ProcessState.String()

	done := time.Since(start)
	_logger.Println("Took ", done.Seconds(), "seconds")
	if timer != nil {
		timer.Stop()
	}

	req.InputReader.Close()

	req.OutputWriter.Write([]byte("Trace-ID: " + req.TractID + strconv.Itoa(getenvInt("fluent_port", 24224)) + getenvStr("fluent_host", "localhost")))

	if waitErr != nil {
		return meta_data, waitErr
	}

	return meta_data, nil
}
