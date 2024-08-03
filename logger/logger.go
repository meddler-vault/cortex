package logger

import (
	"fmt"
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
var _logger *ColorLogger

func init() {

	TaskId = new(taskId)

	InitNewTask("cortx.nucleus")

	_logger = NewColorLogger()

	err := Println("cortex", "hello-world")

	_logger.infoLogger.Print("logger-fluentd", logger.FluentHost, "  ", err)
	// use package init to make sure path is always instantiated
}

func InitNewTask(tag string) {

	TaskId.mu.Lock()
	defer TaskId.mu.Unlock()
	TaskId.taskid = tag

}

func Logln(v ...interface{}) {

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

	// log.Println(v)
	// _logger.warningLogger.Println(v)
	_logger.Warning(v...)
	// err := logger.Post("system", loggingData)
	// log.Println(err)
	// log.Println(logger.FluentHost)
	// log.Println(logger.FluentPort)
}

func Println(v ...interface{}) error {

	// loggingData["message"] = scanner.Text()
	// logger.Post(tag, loggingData
	_logger.Info(v...)

	message := ""
	for _, msg := range v {
		pMsg := fmt.Sprintf("%v ", msg)
		message += pMsg
	}

	loggingData := map[string]string{
		"metadata": "golang-fluentd",
		"message":  message,
	}

	// _logger.Info("logging:", TaskId.taskid, loggingData)

	return logger.Post(TaskId.taskid, loggingData)
	// log.Println(loggingData)

	// err := logger.Post("system", loggingData)
	// log.Println(err)
	// log.Println(logger.FluentHost)
	// log.Println(logger.FluentPort)
}

func Fatalln(v ...interface{}) {

	Println(v)
	os.Exit(1)

}
