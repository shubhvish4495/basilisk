package helper

import (
	"log/slog"
	"os"
)

type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

var l Logger

// GetLogger returns the singleton instance of Logger. If the logger has not been
// initialized yet, it initializes the logger by calling InitLogger with a nil argument.
func GetLogger() Logger {
	if l == nil {
		InitLogger(nil)
	}
	return l
}

// InitLogger initializes the logger with the provided Logger instance.
// If the provided Logger instance is nil, it creates a new JSON logger
// with default settings and assigns it to the global logger variable.
//
// Parameters:
//
//	lVar - Logger instance to initialize. If nil, a default JSON logger is created.
func InitLogger(lVar Logger) {
	// init logger
	if lVar == nil {
		lVar = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		}))
	}

	l = lVar
}
