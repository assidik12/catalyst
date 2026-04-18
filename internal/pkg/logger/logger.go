package logger

import (
	"log/slog"
	"os"
)

// New initializes a new structured logger (slog).
// It uses JSON format for "production" environment and Text format for others (dev/test).
func New(env string) *slog.Logger {
	var handler slog.Handler

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
