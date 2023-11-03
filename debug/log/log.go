package log

import (
	"os"
	"strings"

	"github.com/mokiat/gog/ds"
)

const defaultNamespace = "app"

var (
	enabledDebugNamespaces *ds.Set[string]
	rootLogger             Logger
)

func init() {
	if debugEnv, ok := os.LookupEnv("DEBUG"); ok {
		enabledDebugNamespaces = ds.SetFromSlice(strings.Split(debugEnv, ","))
	} else {
		enabledDebugNamespaces = ds.NewSet[string](0)
	}

	rootLogger = Namespace(defaultNamespace)
}

// IsNamespaceDebugEnabled returns whether debug level is enabled for the
// specified namespace.
func IsNamespaceDebugEnabled(namespace string) bool {
	return enabledDebugNamespaces.Contains(namespace)
}

// Namespace creates a new Logger with the specified namespace.
func Namespace(namespace string) Logger {
	return Logger{
		namespace:    namespace,
		debugEnabled: IsNamespaceDebugEnabled(namespace),
	}
}

// Logger provides a mechanism by which log messages can be output.
type Logger struct {
	namespace    string
	debugEnabled bool
}

// IsDebugEnabled returns whether this Logger will print debug messages.
func (l Logger) IsDebugEnabled() bool {
	return l.debugEnabled
}

// Debug logs a debug message.
func (l Logger) Debug(format string, args ...any) {
	if l.debugEnabled {
		output("DEBUG", "", l.namespace, format, args...)
	}
}

// Info logs an info message.
func (l Logger) Info(format string, args ...any) {
	output("INFO", " ", l.namespace, format, args...)
}

// Warn logs a warning message.
func (l Logger) Warn(format string, args ...any) {
	output("WARN", " ", l.namespace, format, args...)
}

// Error logs an error message.
func (l Logger) Error(format string, args ...any) {
	output("ERROR", "", l.namespace, format, args...)
}

// Debug logs a debug message at the root path.
func Debug(format string, args ...any) {
	rootLogger.Debug(format, args...)
}

// Info logs an info message at the root path.
func Info(format string, args ...any) {
	rootLogger.Info(format, args...)
}

// Warn logs a warning message at the root path.
func Warn(format string, args ...any) {
	rootLogger.Warn(format, args...)
}

// Error logs an error message at the root path.
func Error(format string, args ...any) {
	rootLogger.Error(format, args...)
}
