// internal/api/tournament_handlers.go
// Tournament management HTTP handlers

package api

import (
	"net/http"
	"strconv"

	"tournament-planner/internal/repositories"
	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// HandleCreateTournament handles tournament creation
func HandleCreateTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		organizerID := c.GetString("user_id")

		var req services.CreateTournamentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
			return
		}

		tournament, err := tournamentService.Create(c.Request.Context(), organizerID, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tournament", "details": err.Error()})
			return
		}

		// Include capacity information in response
		c.JSON(http.StatusCreated, gin.H{
			"tournament": tournament,
			"capacity_info": gin.H{
				"calculated_capacity": tournament.CapacityLimit,
				"max_matches_total":   tournament.MaxMatchesPerDay * tournamentService.CalculateTournamentDays(tournament.StartDate, tournament.EndDate),
				"message":             "Tournament created successfully with calculated capacity",
			},
		})
	}
}

// HandleGetTournament retrieves a single tournament
func HandleGetTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		tournament, err := tournamentService.GetByID(c.Request.Context(), tournamentID)
		if err != nil {
			if err == services.ErrNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tournament"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tournament": tournament,
		})
	}
}

// HandleListTournaments lists tournaments with filters
func HandleListTournaments(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

		filter := repositories.ListFilter{
			Page:        page,
			Limit:       limit,
			OrganizerID: c.Query("organizer_id"),
			Status:      c.Query("status"),
			Public:      c.Query("public") == "true",
			Search:      c.Query("search"),
		}

		tournaments, total, err := tournamentService.List(c.Request.Context(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tournaments"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"tournaments": tournaments,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + limit - 1) / limit,
			},
		})
	}
}

// HandleUpdateTournament updates tournament information
func HandleUpdateTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		if err := tournamentService.Update(c.Request.Context(), tournamentID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tournament"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tournament updated successfully"})
	}
}

// HandleDeleteTournament soft deletes a tournament
func HandleDeleteTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement soft delete
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete not implemented yet"})
	}
}

// HandlePublishTournament publishes a tournament and opens registration
func HandlePublishTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		if err := tournamentService.Publish(c.Request.Context(), tournamentID); err != nil {
			if err == services.ErrNoVenues {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot publish tournament without venues"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish tournament"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Tournament published successfully"})
	}
}

// HandleGenerateFixtures generates tournament fixtures
func HandleGenerateFixtures(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		var req struct {
			SeedingMethod string                 `json:"seeding_method" binding:"required,oneof=manual random skill"`
			SeedingData   []services.SeedingData `json:"seeding_data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		fixtures, err := tournamentService.GenerateFixtures(c.Request.Context(), tournamentID, req.SeedingMethod, req.SeedingData)
		if err != nil {
			if err == services.ErrInsufficientParticipants {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient participants to generate fixtures"})
				return
			}
			if err == services.ErrCapacityExceeded {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament format requires more matches than capacity allows"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate fixtures", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Fixtures generated successfully",
			"fixtures": fixtures,
			"count":    len(fixtures),
		})
	}
}

// HandleAutoSchedule automatically schedules all matches
func HandleAutoSchedule(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement auto-scheduling algorithm
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Auto-scheduling not implemented yet"})
	}
}

// HandleGetBracket retrieves tournament bracket
func HandleGetBracket(tournamentService *services.TournamentService, matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		// Get tournament
		tournament, err := tournamentService.GetByID(c.Request.Context(), tournamentID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
			return
		}

		// Get all matches
		matches, err := matchService.GetByTournamentID(c.Request.Context(), tournamentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve matches"})
			return
		}

		// TODO: Structure matches into proper bracket format based on tournament type

		c.JSON(http.StatusOK, gin.H{
			"tournament": gin.H{
				"id":     tournament.ID,
				"name":   tournament.Name,
				"format": tournament.FormatType,
			},
			"matches": matches,
		})
	}
}

// HandleGetSchedule retrieves tournament schedule
func HandleGetSchedule(tournamentService *services.TournamentService, matchService *services.MatchService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")

		matches, err := matchService.GetByTournamentID(c.Request.Context(), tournamentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve schedule"})
			return
		}

		// Group matches by date and venue
		schedule := make(map[string]interface{})
		// TODO: Implement proper schedule grouping

		c.JSON(http.StatusOK, gin.H{
			"schedule": matches,
		})
	}
}

// HandleGetParticipants retrieves tournament participants
func HandleGetParticipants(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement participant retrieval
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Get participants not implemented yet"})
	}
}

// HandleRegisterParticipant handles participant registration
func HandleRegisterParticipant(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tournamentID := c.Param("id")
		userID, _ := c.Get("user_id")

		var req struct {
			Name             string                 `json:"name" binding:"required"`
			Type             string                 `json:"type" binding:"required,oneof=individual team"`
			ContactEmail     string                 `json:"contact_email" binding:"required,email"`
			ContactPhone     string                 `json:"contact_phone"`
			RegistrationData map[string]interface{} `json:"registration_data"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// TODO: Implement registration logic
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Registration not implemented yet"})
	}
}

// HandleJoinWaitlist handles waitlist registration
func HandleJoinWaitlist(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement waitlist logic
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Waitlist not implemented yet"})
	}
}

// Additional tournament handlers...
// HandleStartTournament, HandleCompleteTournament, HandleGetVenues, etc.
// These would follow similar patterns

// Venue management handlers
func HandleGetVenues(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleAddVenue(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleUpdateVenue(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleDeleteVenue(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

// Participant management handlers
func HandleUpdateParticipant(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleRemoveParticipant(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleCheckInParticipant(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleStartTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}

func HandleCompleteTournament(tournamentService *services.TournamentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
	}
}
