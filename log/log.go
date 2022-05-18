// Package log provides common functions for logging.
package log

import "log"

// Info logs an info message.
func Info(format string, args ...interface{}) {
	log.Printf("[info] "+format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...interface{}) {
	log.Printf("[warning] "+format, args...)
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	log.Printf("[error] "+format, args...)
}
