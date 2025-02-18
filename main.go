package main

import (
	"errors"
	"fmt"
	"log"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/meddler-vault/cortex/bootstrap"
	consumernats "github.com/meddler-vault/cortex/consumer-nats"
	"github.com/meddler-vault/cortex/healthchecker"
	"github.com/meddler-vault/cortex/logger"
	"github.com/meddler-vault/cortex/selfupdate"

	"os"
	"syscall"

	reaper "github.com/ramr/go-reaper"
)

// Do not change this logic
func doUpdateStartupCheck() error {

	log.Println("doUpdateStartupCheck")
	if bootstrap.DEBUG {
		return errors.New("debug mode enabled...skipping update and restaet!")
	}
	// selfupdate.ForceQuit()
	// return nil

	_, version, err := selfupdate.Update(consumernats.WatchdogVersion)
	if err != nil {
		// Handle error
		log.Println("+++++++ [[No Force Restarting Startup]] +++++++", err)

		return err
	} else {
		log.Println("+++++++ [[Force Restarting Startup Required]] +++++++", consumernats.WatchdogVersion, " -->", version)
		selfupdate.ForceQuit()

	}

	return nil

}
func handleSignals(pid int) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		fmt.Printf("Received signal %s, terminating child process %d\n", sig, pid)
		// Send signal to child process
		syscall.Kill(pid, sig.(syscall.Signal))
		os.Exit(0)
	}()
}
func main() {
	reap, hasReaper := os.LookupEnv("REAPER")
	logger.Println("LookupEnv REAPER", reap, hasReaper)
	// Use an environment variable REAPER to indicate whether or not
	// we are the child/parent.
	if _, hasReaper = os.LookupEnv("REAPER"); !hasReaper {
		log.Println("Started REAPER")

		var wstatus syscall.WaitStatus

		execPath, err := os.Executable()
		if err != nil {
			log.Fatalf("Error getting executable path: %v", err)
		}
		execPath, err = filepath.Abs(execPath) // Get absolute path
		if err != nil {
			log.Fatalf("Error getting absolute path of executable: %v", err)
		}

		log.Println("execPathDefined", execPath)
		//  Start background reaping of orphaned child processes.
		go reaper.Reap()

		for {
			// Define command arguments and environment
			// Note: Optionally add an argument to the end to more
			//       easily distinguish the parent and child in
			//       something like `ps` etc.
			// args := os.Args
			// Typically, you want to include the executable path in the args
			args := append([]string{execPath}, os.Args[1:]...)
			// args := append(os.Args, "#kiddo")

			pwd, err := os.Getwd()
			if err != nil {
				// Note: Better to use a default dir ala "/tmp".
				panic(err)
			}

			kidEnv := []string{fmt.Sprintf("REAPER=%d", os.Getpid())}

			pattrs := &syscall.ProcAttr{
				Dir: pwd,
				Env: append(os.Environ(), kidEnv...),
				Sys: &syscall.SysProcAttr{Setsid: true},
				Files: []uintptr{
					uintptr(syscall.Stdin),
					uintptr(syscall.Stdout),
					uintptr(syscall.Stderr),
				},
			}

			log.Println("ForkExec", execPath, args)

			pid, err := syscall.ForkExec(execPath, args, pattrs)
			handleSignals(pid)

			if err != nil {
				log.Fatalf("Error forking the process: %v", err)
			}
			// fmt.Printf("kiddo-pid = %d\n", pid)
			_, err = syscall.Wait4(pid, &wstatus, 0, nil)
			for syscall.EINTR == err {
				_, err = syscall.Wait4(pid, &wstatus, 0, nil)
			}

			if err != nil {
				log.Fatalf("Error waiting for child process: %v", err)
			}

			// Get the exit status code
			if wstatus.Exited() {
				exitCode := wstatus.ExitStatus()
				fmt.Printf("Child process exited with code %d\n", exitCode)
				if exitCode == 1 {
					fmt.Println("Exit code 1 detected. Restarting child process...")
					time.Sleep(1 * time.Second) // Optional: Add delay before restart
					continue
				}
			} else if wstatus.Signaled() {
				signal := wstatus.Signal()
				fmt.Printf("Child process was terminated by signal %d (%s)\n", signal, signal.String())
				if signal == syscall.SIGINT {
					fmt.Println("Signal SIGINT detected. Restarting child process...")
					time.Sleep(1 * time.Second) // Optional: Add delay before restart
					continue
				}
			} else {
				fmt.Println("Child process did not exit normally")
				// Add more detailed logging if needed
			}
			os.Exit(0)

			// If you put this code into a function, then exit here.
		}
		// return
	}
	cMain()

	//  Rest of your code goes here ...

} /*  End of func  main.  */
func cMain() {
	log.Println("cMain")
	// Get the path to the current executable

	logger.Println("+++++++ [[Watchdog Started]] +++++++", consumernats.WatchdogVersion)

	doUpdateStartupCheck()

	// Start the health checker

	endpoint := "https://api.meddler.io/secoflex/health-check"

	endpoint = bootstrap.CONSTANTS.Reserved.CORTEXPINGURL

	initialMessage := map[string]interface{}{
		"state": "resurrected", "details": "Ressurected. Probably my first times in this world, or reincarnation!"}
	pubsub := healthchecker.InitializeGlobalHealth(bootstrap.CONSTANTS.Reserved.CORTEXUUID, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE, endpoint, initialMessage)
	// stopCh := make(chan struct{})
	// interval := time.Duration(bootstrap.CONSTANTS.Reserved.CORTEXPINGINTERVAL) * time.Second
	// healthchecker.HealthCheckWorker(interval, stopCh)

	//

	consumernats.Start(pubsub)
}

// const version = "1.0.0" // Current version of your application
func _main() {

	log.Println("My version", consumernats.WatchdogVersion)
	err := doUpdateStartupCheck()
	log.Println("Error", err)

}
