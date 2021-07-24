// Package log provides common functions for logging.
package log

import "log"

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	log.Printf("[debug] "+format, args...)
}

// Debug logs an info message.
func Info(format string, args ...interface{}) {
	log.Printf("[info] "+format, args...)
}

// Debug logs a warning message.
func Warn(format string, args ...interface{}) {
	log.Printf("[warning] "+format, args...)
}

// Debug logs an error message.
func Error(format string, args ...interface{}) {
	log.Printf("[error] "+format, args...)
}
