// Package log provides common functions for logging.
package log

import (
	"fmt"
	"log"
	"os"
)

var (
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

func init() {
	flags := log.LstdFlags | log.Lmsgprefix
	debugLogger = log.New(os.Stderr, "[debug] ", flags)
	infoLogger = log.New(os.Stderr, "[info] ", flags)
	warningLogger = log.New(os.Stderr, "[warning] ", flags)
	errorLogger = log.New(os.Stderr, "[error] ", flags)
}

// Info logs an info message.
func Info(format string, args ...interface{}) {
	infoLogger.Output(2, fmt.Sprintf(format, args...))
}

// Warn logs a warning message.
func Warn(format string, args ...interface{}) {
	warningLogger.Output(2, fmt.Sprintf(format, args...))
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	errorLogger.Output(2, fmt.Sprintf(format, args...))
}
