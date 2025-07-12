// internal/api/user_handlers.go
// User profile and preferences HTTP handlers

package api

import (
	"net/http"

	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleGetCurrentUser retrieves the current user's profile
func HandleGetCurrentUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		user, err := userService.GetByID(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
}

// HandleUpdateProfile updates user profile
func HandleUpdateProfile(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		user, err := userService.UpdateProfile(c.Request.Context(), userID, updates)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
}

// HandleGetPreferences retrieves user preferences
func HandleGetPreferences(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		preferences, err := userService.GetPreferences(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve preferences"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"preferences": preferences,
		})
	}
}

// HandleUpdatePreferences updates user preferences
func HandleUpdatePreferences(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var preferences map[string]interface{}
		if err := c.ShouldBindJSON(&preferences); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := userService.UpdatePreferences(c.Request.Context(), userID, preferences); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
	}
}

// HandleGetUserTournaments retrieves user's tournament history
func HandleGetUserTournaments(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		tournaments, err := userService.GetTournamentHistory(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tournaments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tournaments": tournaments,
		})
	}
}

// HandleGetUserStatistics retrieves user statistics
func HandleGetUserStatistics(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		stats, err := userService.GetStatistics(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"statistics": stats,
		})
	}
}
