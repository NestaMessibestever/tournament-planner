// internal/api/admin_handlers.go
// Admin-only HTTP handlers

package api

import (
	"net/http"

	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleGetPlatformStats retrieves platform-wide statistics
func HandleGetPlatformStats(analyticsService *services.AnalyticsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats, err := analyticsService.GetPlatformStats(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve statistics"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"statistics": stats,
		})
	}
}

// HandleListUsers lists all users (admin only)
func HandleListUsers(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement user listing with pagination
		c.JSON(http.StatusNotImplemented, gin.H{"error": "User listing not implemented yet"})
	}
}

// HandleUpdateUserRole updates a user's role
func HandleUpdateUserRole(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")

		var req struct {
			Role string `json:"role" binding:"required,oneof=user organizer admin"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// TODO: Implement role update
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Role update not implemented yet"})
	}
}

// HandleListAllTournaments lists all tournaments (admin only)
func HandleListAllTournaments(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would be similar to HandleListTournaments but without user filtering
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Admin tournament listing not implemented yet"})
	}
}

// HandleForceDeleteTournament force deletes a tournament (admin only)
func HandleForceDeleteTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		// TODO: Implement hard delete with cascade
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Force delete not implemented yet"})
	}
}
