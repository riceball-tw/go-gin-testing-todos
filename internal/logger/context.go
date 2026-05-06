package logger

import "github.com/gin-gonic/gin"

const businessContextKey = "logger_business_context"

// AddBusinessContext adds a key-value pair to the request's business context.
// User provides a custom context type T to enforce compile-time type safety.
// Example: type MyContext map[string]string
func AddBusinessContext[T any](c *gin.Context, key string, value T) {
	ctxMap := make(map[string]interface{})

	if existing, exists := c.Get(businessContextKey); exists {
		ctxMap = existing.(map[string]interface{})
	}

	ctxMap[key] = value
	c.Set(businessContextKey, ctxMap)
}

// GetBusinessContext retrieves the accumulated business context from the request.
// Returns nil if no context exists.
func GetBusinessContext[T any](c *gin.Context) map[string]T {
	if existing, exists := c.Get(businessContextKey); exists {
		rawMap := existing.(map[string]interface{})
		result := make(map[string]T)
		for k, v := range rawMap {
			result[k] = v.(T)
		}
		return result
	}
	return nil
}
