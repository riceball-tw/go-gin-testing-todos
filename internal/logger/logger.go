package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
	"github.com/natefinch/lumberjack"
)

var Log *slog.Logger

func InitLogger() {
	var handler slog.Handler

	if gin.Mode() == gin.DebugMode {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	} else {
		handler = slog.NewJSONHandler(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		}, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	Log = slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("service", "go-gin-testing-todos"),
	}))

	slog.SetDefault(Log)
}
