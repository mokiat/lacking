//go:build !debug

package log

// DebugEnabled is a flag indicating the state of debug level.
var DebugEnabled = false

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
}
