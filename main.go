package main

import (
	"fmt"

	consumernats "github.com/meddler-vault/cortex/consumer-nats"
	"github.com/meddler-vault/cortex/logger"

	"os"
	"syscall"

	reaper "github.com/ramr/go-reaper"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

func __main() {
	reap, hasReaper := os.LookupEnv("REAPER")
	logger.Println("LookupEnv REAPER", reap, hasReaper)
	// Use an environment variable REAPER to indicate whether or not
	// we are the child/parent.
	if _, hasReaper = os.LookupEnv("REAPER"); !hasReaper {
		logger.Println("Started REAPER")

		//  Start background reaping of orphaned child processes.
		go reaper.Reap()

		// Note: Optionally add an argument to the end to more
		//       easily distinguish the parent and child in
		//       something like `ps` etc.
		args := os.Args
		// args := append(os.Args, "#kiddo")

		pwd, err := os.Getwd()
		if err != nil {
			// Note: Better to use a default dir ala "/tmp".
			panic(err)
		}

		kidEnv := []string{fmt.Sprintf("REAPER=%d", os.Getpid())}

		var wstatus syscall.WaitStatus
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

		pid, _ := syscall.ForkExec(args[0], args, pattrs)

		// fmt.Printf("kiddo-pid = %d\n", pid)
		_, err = syscall.Wait4(pid, &wstatus, 0, nil)
		for syscall.EINTR == err {
			_, err = syscall.Wait4(pid, &wstatus, 0, nil)
		}

		// If you put this code into a function, then exit here.
		os.Exit(0)
		return
	}
	// _main()

	//  Rest of your code goes here ...

} /*  End of func  main.  */
func _main() {
	logger.Println("[[Watchdog]]", consumernats.WatchdogVersion)
	consumernats.Start()
}

// const version = "1.0.0" // Current version of your application
func main() {

	var updater = &selfupdate.Updater{
		CurrentVersion: consumernats.WatchdogVersion, // the current version of your app used to determine if an update is necessary
		// these endpoints can be the same if everything is hosted in the same place
		ApiURL:  "http://updates.yourdomain.com/", // endpoint to get update manifest
		BinURL:  "http://updates.yourdomain.com/", // endpoint to get full binaries
		DiffURL: "http://updates.yourdomain.com/", // endpoint to get binary diff/patches
		Dir:     "update/",                        // directory relative to your app to store temporary state files related to go-selfupdate
		CmdName: "cortex",                         // your app's name (must correspond to app name hosting the updates)
		// app name allows you to serve updates for multiple apps on the same server/endpoint
	}

	// go look for an update when your app starts up
	updater.BackgroundRun()
	// your app continues to run...
}
