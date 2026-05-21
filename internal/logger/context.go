package logger

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

const businessContextKey = "logger_business_context"

// AddBusinessContext adds a key-value pair to the request's business context.
func AddBusinessContext(c *gin.Context, attr slog.Attr) {
	ctxMap := make(map[string]slog.Attr)

	if existing, exists := c.Get(businessContextKey); exists {
		ctxMap = existing.(map[string]slog.Attr)
	}

	ctxMap[attr.Key] = attr
	c.Set(businessContextKey, ctxMap)
}

func AddResourceAction(c *gin.Context, resource string, action string) {
	AddBusinessContext(c, slog.String("resource", resource))
	AddBusinessContext(c, slog.String("action", action))
}

// GetBusinessContext retrieves the accumulated business context from the request.
// Returns nil if no context exists.
func GetBusinessContext(c *gin.Context) map[string]slog.Attr {
	if existing, exists := c.Get(businessContextKey); exists {
		return existing.(map[string]slog.Attr)
	}
	return nil
}
