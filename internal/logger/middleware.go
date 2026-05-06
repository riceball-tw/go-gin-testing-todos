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

		fields := []any{
			slog.String("method", c.Request.Method),
			slog.Int("status_code", statusCode),
			slog.String("path", c.Request.URL.Path),
			slog.String("query", c.Request.URL.RawQuery),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
			slog.String("requestId", requestIDStr),
		}

		// Gather business context and determine log message
		var msg string = "http_completed"
		if bizCtx := GetBusinessContext(c); bizCtx != nil {
			for k, v := range bizCtx {
				fields = append(fields, slog.Any(k, v))
			}

			// Build message from resource + action if both present
			if resource, ok := bizCtx["resource"].(string); ok {
				if action, ok := bizCtx["action"].(string); ok {
					msg = resource + "_" + action
				}
			}
		}

		// Add Error to log
		if len(c.Errors) > 0 {
				fields = append(fields, slog.String("error_message", c.Errors.Last().Error()))
		}

		switch {
		case statusCode >= 500:
				Log.Error(msg, fields...)
		case statusCode >= 400:
				Log.Warn(msg, fields...)
		default:
				Log.Info(msg, fields...)
		}
	}
}
