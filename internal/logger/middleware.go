package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Base on Structured Logging from gin doc
// https://gin-gonic.com/en/docs/logging/structured-logging/
// WideEventMiddleware intercepts requests and emits a single wide event upon completion.
func WideEventMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		statusCode := c.Writer.Status()
		requestID, _ := c.Get("requestId")
		requestIDStr, _ := requestID.(string)

		fields := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.Int("status_code", statusCode),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("requestId", requestIDStr),
		}

		// Add Error to log
		if len(c.Errors) > 0 {
			fields = append(fields, slog.String("error_message", c.Errors.Last().Error()))
		}

		msg := "request_completed"
		switch {
		case statusCode >= 500:
			Log.LogAttrs(c.Request.Context(), slog.LevelError, msg, fields...)
		case statusCode >= 400:
			Log.LogAttrs(c.Request.Context(), slog.LevelWarn, msg, fields...)
		default:
			Log.LogAttrs(c.Request.Context(), slog.LevelInfo, msg, fields...)
		}
	}
}