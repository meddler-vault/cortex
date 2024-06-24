package logger

import (
	"log"
	"os"

	"github.com/fatih/color"
)

type ColorLogger struct {
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
}

func NewColorLogger() *ColorLogger {
	infoColor := color.New(color.FgGreen).SprintFunc()
	warningColor := color.New(color.FgYellow).SprintFunc()
	errorColor := color.New(color.FgRed).SprintFunc()

	return &ColorLogger{
		infoLogger:    log.New(os.Stdout, infoColor("[INFO] "), log.LstdFlags),
		warningLogger: log.New(os.Stdout, warningColor("[WARNING] "), log.LstdFlags),
		errorLogger:   log.New(os.Stderr, errorColor("[ERROR] "), log.LstdFlags),
	}
}

func (cl *ColorLogger) Info(v ...interface{}) {
	cl.infoLogger.Println(v...)
}

func (cl *ColorLogger) Warning(v ...interface{}) {
	cl.warningLogger.Println(v...)
}

func (cl *ColorLogger) Error(v ...interface{}) {
	cl.errorLogger.Println(v...)
}

func main() {
	logger := NewColorLogger()

	logger.Info("This is an info message")
	logger.Warning("This is a warning message")
	logger.Error("This is an error message")
}
