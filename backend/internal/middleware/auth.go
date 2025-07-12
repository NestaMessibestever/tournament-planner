// internal/middleware/auth.go
// Authentication middleware validates JWT tokens and sets user context

package middleware

import (
	"net/http"
	"strings"

	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// RequireAuth validates that a request has a valid JWT token
func RequireAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		// Validate token
		userID, role, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", userID)
		c.Set("user_role", role)
		c.Set("authenticated", true)

		c.Next()
	}
}

// OptionalAuth checks for authentication but doesn't require it
func OptionalAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Set("authenticated", false)
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			if userID, role, err := authService.ValidateToken(parts[1]); err == nil {
				c.Set("user_id", userID)
				c.Set("user_role", role)
				c.Set("authenticated", true)
			}
		}

		c.Next()
	}
}

// RequireRole ensures the user has a specific role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		if role.(string) != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTournamentOwner ensures the user owns the tournament
func RequireTournamentOwner(services *services.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		tournamentID := c.Param("id")

		// Check if user owns the tournament
		isOwner, err := services.Tournament.IsOwner(c.Request.Context(), tournamentID, userID.(string))
		if err != nil || !isOwner {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMatchAccess ensures the user can access/modify a match
func RequireMatchAccess(services *services.Container) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		matchID := c.Param("id")

		// Check if user has access to the match
		hasAccess, err := services.Match.HasAccess(c.Request.Context(), matchID, userID.(string))
		if err != nil || !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		c.Next()
	}
}
