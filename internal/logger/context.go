package logger

import "github.com/gin-gonic/gin"

const businessContextKey = "logger_business_context"

// AddBusinessContext adds a key-value pair to the request's business context.
// This context will be logged as part of the wide event at the end of the request.
func AddBusinessContext(c *gin.Context, key string, value any) {
	var ctxMap map[string]any

	if existing, exists := c.Get(businessContextKey); exists {
		ctxMap = existing.(map[string]any)
	} else {
		ctxMap = make(map[string]any)
	}

	ctxMap[key] = value
	c.Set(businessContextKey, ctxMap)
}

// GetBusinessContext retrieves the accumulated business context from the request.
func GetBusinessContext(c *gin.Context) map[string]any {
	if existing, exists := c.Get(businessContextKey); exists {
		return existing.(map[string]any)
	}
	return nil
}
