package watchdog

import (
	"log"
	"os"

	"github.com/meddler-io/watchdog/config"
	"github.com/meddler-io/watchdog/executor"
)

func Start(cmd string, args []string, env map[string]string) error {

	// environment := make(map[string]string)
	environment := []string{
		"exec_timeout=2",
		"fprocess=sleep",
	}

	environment = []string{}

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
	log.Println("Running", cmd, args)
	req := executor.FunctionRequest{
		Process:                 cmd,
		ProcessArgs:             args,
		InputReader:             os.Stdin,
		OutputWriter:            os.Stdout,
		Environment:             environment,
		TractID:                 env["TraceId"],
		CurrentWorkingDirectory: env["CWD"],
	}

	err := functionInvoker.Run(req)
	if err != nil {
		log.Println(err)
	}
	return err
}
