// ========================================
// internal/middleware/maintenance.go
// Maintenance mode middleware

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaintenanceMode returns 503 when maintenance mode is enabled
func MaintenanceMode() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow health check endpoint even in maintenance mode
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service temporarily unavailable for maintenance",
			"message": "We'll be back shortly!",
		})
		c.Abort()
	}
}
