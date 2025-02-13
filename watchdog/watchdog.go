package watchdog

import (
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/meddler-vault/cortex/logger"

	"github.com/meddler-vault/cortex/config"
	"github.com/meddler-vault/cortex/executor"
)

func Start(id string, cmd []string, args []string, env map[string]string) (error, map[string]interface{}) {

	// environment := make(map[string]string)
	environment := []string{}

	// environment = []string{}

	for k, v := range env {
		environment = append(environment, k+"="+v)
	}

	// environment["exec_timeout"] = "1000"
	// environment["fprocess"] = "ls"

	watchdogConfig := config.New(env)
	if watchdogConfig.InjectCGIHeaders {

	}

	// commandName, arguments := watchdogConfig.Process()
	functionInvoker := executor.ForkFunctionRunner{
		ExecTimeout: watchdogConfig.ExecTimeout,
	}

	// commandName = "echo"
	// arguments = []string{"10"}

	if len(cmd) > 1 {
		args = append(cmd[1:], args...)
	}

	logger.Logln("Running", cmd[0], args)

	req := executor.FunctionRequest{
		Process:                 cmd[0],
		ProcessArgs:             args,
		InputReader:             os.Stdin,
		OutputWriter:            os.Stdout,
		Environment:             environment,
		TractID:                 id,
		CurrentWorkingDirectory: env["CWD"],
	}

	logger.Println("Environment", req.Environment)

	start_time := time.Now().Unix()
	err := functionInvoker.Run(req)
	end_time := time.Now().Unix()

	exitCode := -1
	if err != nil {
		logger.Println(err)

		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}
	}

	return err, map[string]interface{}{
		"start_time":     start_time,
		"end_time":       end_time,
		"execution_time": end_time - start_time,
		"exit_code":      exitCode,
	}
}
