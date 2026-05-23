package logger

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const businessContextKey = "logger_context"

// AddContext appends typed slog attributes to the request's business context.
func AddContext(c *gin.Context, attrs ...slog.Attr) {
	if len(attrs) == 0 {
		return
	}

	ctxAttrs := make([]slog.Attr, 0, len(attrs))
	if existing, exists := c.Get(businessContextKey); exists {
		if existingAttrs, ok := existing.([]slog.Attr); ok {
			ctxAttrs = append(ctxAttrs, existingAttrs...)
		}
	}

	ctxAttrs = append(ctxAttrs, attrs...)
	c.Set(businessContextKey, ctxAttrs)
}

// GetContext retrieves the accumulated business context from the request.
// Returns a copy so callers cannot mutate the context state.
func GetContext(c *gin.Context) []slog.Attr {
	if existing, exists := c.Get(businessContextKey); exists {
		if attrs, ok := existing.([]slog.Attr); ok {
			return append([]slog.Attr(nil), attrs...)
		}
	}
	return nil
}
