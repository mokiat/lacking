// Package log provides common functions for logging.
package log

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger

	allowedDebugPaths []string

	rootLogger Logger
)

func init() {
	flags := log.LstdFlags | log.Lmsgprefix
	debugLogger = log.New(os.Stdout, "[debug] ", flags)
	infoLogger = log.New(os.Stdout, "[info]  ", flags)
	warningLogger = log.New(os.Stderr, "[warn]  ", flags)
	errorLogger = log.New(os.Stderr, "[error] ", flags)

	if debugEnv, ok := os.LookupEnv("DEBUG"); ok {
		allowedDebugPaths = strings.Split(debugEnv, ",")
	}

	rootLogger = Path("/")
}

func isDebugPathEnabled(path string) bool {
	for _, allowedPath := range allowedDebugPaths {
		if strings.HasPrefix(path, allowedPath) {
			return true
		}
	}
	return false
}

// Path creates a new Logger at the specified path.
func Path(path string) Logger {
	return Logger{
		path:         path,
		debugEnabled: isDebugPathEnabled(path),

		debugPrefix:   fmt.Sprintf("[debug] [%s] ", path),
		infoPrefix:    fmt.Sprintf("[info]  [%s] ", path),
		warningPrefix: fmt.Sprintf("[warn]  [%s] ", path),
		errorPrefix:   fmt.Sprintf("[error] [%s] ", path),
	}
}

// Logger provides a mechanism by which log messages can be output.
type Logger struct {
	path         string
	debugEnabled bool

	debugPrefix   string
	infoPrefix    string
	warningPrefix string
	errorPrefix   string
}

// DebugEnabled returns whether this Logger will print debug messages.
func (l Logger) DebugEnabled() bool {
	return l.debugEnabled
}

// Path creates a new Logger by appending the specified path to the
// current logger's path.
func (l Logger) Path(path string) Logger {
	return Path(l.path + path)
}

// Debug logs a debug message.
func (l Logger) Debug(format string, args ...any) {
	if l.debugEnabled {
		debugLogger.SetPrefix(l.debugPrefix)
		debugLogger.Printf(format, args...)
	}
}

// Info logs an info message.
func (l Logger) Info(format string, args ...any) {
	infoLogger.SetPrefix(l.infoPrefix)
	infoLogger.Printf(format, args...)
}

// Warn logs a warning message.
func (l Logger) Warn(format string, args ...any) {
	warningLogger.SetPrefix(l.warningPrefix)
	warningLogger.Printf(format, args...)
}

// Error logs an error message.
func (l Logger) Error(format string, args ...any) {
	errorLogger.SetPrefix(l.errorPrefix)
	errorLogger.Printf(format, args...)
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
