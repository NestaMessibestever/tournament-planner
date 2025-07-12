// internal/api/match_handlers.go
// Match management HTTP handlers

package api

import (
	"net/http"
	"time"

	"tournament-planner/internal/models"
	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleGetMatch retrieves a single match
func HandleGetMatch(matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		match, err := matchService.GetByID(c.Request.Context(), matchID)
		if err != nil {
			if err == services.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Match not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve match"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"match": match,
		})
	}
}

// HandleUpdateMatch updates match information
func HandleUpdateMatch(matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		var req struct {
			ScheduledTime string `json:"scheduled_time"`
			VenueID       string `json:"venue_id"`
			RefereeID     string `json:"referee_id"`
			Notes         string `json:"notes"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Parse scheduled time if provided
		if req.ScheduledTime != "" {
			scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledTime)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid datetime format"})
				return
			}

			if err := matchService.UpdateSchedule(c.Request.Context(), matchID, scheduledTime, req.VenueID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update match"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Match updated successfully"})
	}
}

// HandleStartMatch marks a match as in progress
func HandleStartMatch(matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		if err := matchService.StartMatch(c.Request.Context(), matchID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start match"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Match started successfully"})
	}
}

// HandleReportScore reports match score
func HandleReportScore(matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		var req struct {
			Score1       int                  `json:"score1" binding:"min=0"`
			Score2       int                  `json:"score2" binding:"min=0"`
			ScoreDetails *models.ScoreDetails `json:"score_details"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := matchService.ReportScore(c.Request.Context(), matchID, req.Score1, req.Score2, req.ScoreDetails); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to report score", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Score reported successfully"})
	}
}

// HandleCancelMatch cancels a match
func HandleCancelMatch(matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		matchID := c.Param("id")

		var req struct {
			Reason string `json:"reason" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := matchService.CancelMatch(c.Request.Context(), matchID, req.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel match"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Match cancelled successfully"})
	}
}
