//go:build debug

package log

import "fmt"

// DebugEnabled is a flag indicating the state of debug level.
var DebugEnabled = true

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	debugLogger.Output(2, fmt.Sprintf(format, args...))
}
