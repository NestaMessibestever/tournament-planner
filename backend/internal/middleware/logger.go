// ========================================
// internal/middleware/logger.go
// Request logging middleware with structured logs

package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger creates a custom logging middleware
func Logger(logger *log.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		// Structured log format
		logger.Printf("[%s] %s %s %d %v %s %s",
			c.GetString("request_id"),
			clientIP,
			method,
			statusCode,
			latency,
			path,
			errorMessage,
		)
	}
}
