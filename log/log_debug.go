//go:build debug

package log

import "log"

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	log.Printf("[debug] "+format, args...)
}
