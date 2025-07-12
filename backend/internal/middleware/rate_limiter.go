// ========================================
// internal/middleware/rate_limiter.go
// Rate limiting to prevent abuse

package middleware

import (
	"fmt"
	"net/http"
	"time"

	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements rate limiting using Redis
func RateLimiter(cache *services.CacheService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID if authenticated)
		var key string
		if userID, exists := c.Get("user_id"); exists {
			key = fmt.Sprintf("rate_limit:user:%s", userID)
		} else {
			key = fmt.Sprintf("rate_limit:ip:%s", c.ClientIP())
		}

		// Check rate limit (100 requests per minute)
		limit := 100
		window := time.Minute

		count, err := cache.Increment(key, window)
		if err != nil {
			// Don't block on rate limit errors
			c.Next()
			return
		}

		if count > limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limit-count))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

		c.Next()
	}
}
