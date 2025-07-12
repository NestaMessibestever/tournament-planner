// ========================================
// internal/middleware/request_id.go
// Generates unique request IDs for tracing

package middleware

import (
	"tournament-planner/internal/utils"

	"github.com/gin-gonic/gin"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = utils.GenerateRequestID()
		}

		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
