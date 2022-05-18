//go:build debug

package log

import "log"

// DebugEnabled is a flag indicating the state of debug level.
var DebugEnabled = true

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	log.Printf("[debug] "+format, args...)
}
