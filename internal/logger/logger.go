package logger

import (
	"log/slog"
	"os"
)

// Config holds logger configuration options.
type Config struct {
	Level string
}

// New creates a new JSON slog.Logger that writes to stdout with the given log level.
// Unknown levels fall back to "info".
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
