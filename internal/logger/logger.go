package logger

import (
	"log/slog"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
	"github.com/natefinch/lumberjack"
	slogsyslog "github.com/samber/slog-syslog/v2"
)

var Log *slog.Logger

func InitLogger() {
	var handler slog.Handler

	const pinkColor uint8 = 5

	if gin.Mode() == gin.DebugMode {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == "error_message" && len(groups) == 0 {
					return tint.Attr(pinkColor, a)
				}
				return a
			},
		})
	} else {
		fileHandler := slog.NewJSONHandler(&lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     30,
			Compress:   true,
		}, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

		handler = fileHandler
		if syslogURL := os.Getenv("SYSLOG_URL"); syslogURL != "" {
			if u, err := url.Parse(syslogURL); err == nil {
				if syslogWriter, err := net.Dial(u.Scheme, u.Host); err == nil {
					handler = slogsyslog.Option{
						Level:  slog.LevelInfo,
						Writer: syslogWriter,
					}.NewSyslogHandler()
				}
			}
		}
	}

	Log = slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("service", "go-gin-testing-todos"),
	}))

	slog.SetDefault(Log)
}
