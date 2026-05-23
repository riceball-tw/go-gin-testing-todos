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
	slogmulti "github.com/samber/slog-multi"
	slogsyslog "github.com/samber/slog-syslog/v2"
)

var Log = slog.Default()

func InitLogger() {
	initLogger(buildHandler())
}

func initLogger(handler slog.Handler) {
	Log = slog.New(handler.WithAttrs([]slog.Attr{
		slog.String("service", "go-gin-testing-todos"),
	}))

	slog.SetDefault(Log)
}

func buildHandler() slog.Handler {
	if gin.Mode() == gin.DebugMode {
		return buildDebugHandler()
	}

	if syslogHandler := newSyslogHandler(os.Getenv("SYSLOG_URL")); syslogHandler != nil {
		return slogmulti.Fanout(syslogHandler)
	}

	handlers := []slog.Handler{newFileHandler()}
	if mongoLogHandler := newMongoLogHandler(); mongoLogHandler != nil {
		handlers = append(handlers, mongoLogHandler)
	}

	return slogmulti.Fanout(handlers...)
}

func buildDebugHandler() slog.Handler {
	const pinkColor uint8 = 5

	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      slog.LevelDebug,
		TimeFormat: time.Kitchen,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "error_message" && len(groups) == 0 {
				return tint.Attr(pinkColor, a)
			}
			return a
		},
	})
}

func newFileHandler() slog.Handler {
	return slog.NewJSONHandler(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
}

func newSyslogHandler(syslogURL string) slog.Handler {
	if syslogURL == "" {
		return nil
	}

	u, err := url.Parse(syslogURL)
	if err != nil {
		return nil
	}

	syslogWriter, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		return nil
	}

	return slogsyslog.Option{
		Level:  slog.LevelInfo,
		Writer: syslogWriter,
	}.NewSyslogHandler()
}
