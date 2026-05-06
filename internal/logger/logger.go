package logger

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// InitLogger initializes a global structured JSON logger.
func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	Log = slog.New(handler.WithAttrs([]slog.Attr{
		// Add global characteristics to all logs from this logger
		slog.String("service", "go-gin-testing-todos"),
	}))

	slog.SetDefault(Log)
}
