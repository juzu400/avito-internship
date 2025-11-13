package logger

import (
	"log/slog"
	"os"
)

type Config struct {
	Level string
}

func New(cfg Config) *slog.Logger {
	level := slog.LevelInfo

	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}
