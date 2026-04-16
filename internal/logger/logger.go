package logger

import (
	"log/slog"
	"os"
)

func New(level string) *slog.Logger {
	var lvl slog.Level

	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	})

	return slog.New(handler)
}
