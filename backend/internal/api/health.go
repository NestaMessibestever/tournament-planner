// internal/api/health.go
// Health check endpoint for monitoring

package api

import (
	"net/http"

	"tournament-planner/internal/config"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns a health check handler
func HealthCheck(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "healthy",
			"environment": cfg.Environment,
			"version":     "1.0.0",
			"services": gin.H{
				"api":       "operational",
				"websocket": cfg.Features.EnableWebSocket,
				"payments":  cfg.Features.EnablePayments,
			},
		})
	}
}
