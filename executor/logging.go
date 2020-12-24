package executor

import (
	"bufio"
	"io"
	"log"
	"os"
	"sync"

	"github.com/fluent/fluent-logger-golang/fluent"
)

const maxCapacity = 512 * 1024

// bindLoggingPipe spawns a goroutine for passing through logging of the given output pipe.
func bindLoggingPipe(name string, tag string, pipe io.Reader, output io.Writer) *sync.WaitGroup {
	var wg sync.WaitGroup

	log.Printf("Started logging %s from function. TAG: %s", name, tag)

	scanner := bufio.NewScanner(pipe)
	// logger := log.New(output, log.Prefix(), log.Flags())

	logger, _ := fluent.New(fluent.Config{

		FluentPort: 24224,
		FluentHost: "localhost",
		TagPrefix:  "myapp.",
	})

	wg.Add(1)
	go func() {

		loggingData := map[string]string{
			"metadata": "golang-fluentd",
			"message":  "",
		}

		for scanner.Scan() {
			loggingData["message"] = scanner.Text()
			logger.Post(tag, loggingData)

		}
		if err := scanner.Err(); err != nil {
			loggingData["message"] = err.Error()

			logger.Post(tag, loggingData)
			log.Println("Scanner Error", err)
			// log.Printf("Error scanning %s: %s: %s", name, tag, err.Error())
		}

		wg.Done()
	}()

	return &wg
}

// bindLoggingPipe spawns a goroutine for passing through logging of the given output pipe.
func bindFluentLoggingPipe(logger *fluent.Fluent, name string, tag string, pipe io.Reader, wg *sync.WaitGroup) {

	log.Printf(os.Getenv("fluent_host"))
	log.Printf(os.Getenv("fluent_port"))
	log.Printf("Started logging %s from function. TAG: %s", name, tag)

	scanner := bufio.NewScanner(pipe)

	//adjust the capacity to your need (max characters in line)
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	// logger := log.New(output, log.Prefix(), log.Flags())

	wg.Add(1)
	go func() {

		loggingData := map[string]string{
			"message": "",
			"pipe":    name,
		}

		for scanner.Scan() {
			loggingData["message"] = scanner.Text()
			logger.Post(tag, loggingData)
			log.Println("message", tag, loggingData)

		}

		loggingData["pipe"] = "stdend"

		if err := scanner.Err(); err != nil {
			loggingData["message"] = err.Error()
			logger.Post(tag, loggingData)
			log.Println("err", err)

		}

		wg.Done()
	}()

}
