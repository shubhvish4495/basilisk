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

func GetLogger() Logger {
	if l == nil {
		InitLogger(nil)
	}
	return l
}

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
