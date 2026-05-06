package logger

import (
	"log/slog"

	"github.com/natefinch/lumberjack"
)

var Log *slog.Logger

func InitLogger() {
	handler := slog.NewJSONHandler(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	Log = slog.New(handler.WithAttrs([]slog.Attr{
		// Add global characteristics to all logs from this logger
		slog.String("service", "go-gin-testing-todos"),
	}))

	slog.SetDefault(Log)
}
