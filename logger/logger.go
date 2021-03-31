package logger

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/fluent/fluent-logger-golang/fluent"
)

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

type taskId struct {
	mu     sync.Mutex
	taskid string
}

var TaskId *taskId

func init() {
	// use package init to make sure path is always instantiated
	TaskId = new(taskId)
}

func InitNewTask(tag string) {

	TaskId.mu.Lock()
	defer TaskId.mu.Unlock()
	TaskId.taskid = tag

}

func Println(v ...interface{}) {

	// loggingData["message"] = scanner.Text()
	// logger.Post(tag, loggingData

	message := ""
	for _, msg := range v {
		pMsg := fmt.Sprintf("%v ", msg)
		message += pMsg
	}

	loggingData := map[string]string{
		"metadata": "golang-fluentd",
		"message":  message,
	}

	logger.Post(TaskId.taskid, loggingData)

	log.Println(v)
	// err := logger.Post("system", loggingData)
	// log.Println(err)
	// log.Println(logger.FluentHost)
	// log.Println(logger.FluentPort)
}

func Fatalln(v ...interface{}) {

	Println(v)
	os.Exit(1)

}
