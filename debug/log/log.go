package log

import (
	"log/slog"
	"os"
)

func init() {
	logLevel := slog.LevelInfo
	if IsDebugEnabled {
		logLevel = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(logLevel)

	goLogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(goLogger)
}

// ForNamespace returns an slog.Logger for the specified namespace.
func ForNamespace(namespace string) *slog.Logger {
	return slog.Default().With("ns", namespace)
}
