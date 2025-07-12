// internal/api/routes.go
// Central route registration for all API endpoints

package api

import (
	"tournament-planner/internal/config"
	"tournament-planner/internal/middleware"
	"tournament-planner/internal/services"

	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(router *gin.RouterGroup, services *services.Container) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", HandleRegister(services.Auth))
		auth.POST("/login", HandleLogin(services.Auth))
		auth.POST("/logout", middleware.RequireAuth(services.Auth), HandleLogout(services.Auth))
		auth.POST("/refresh", HandleRefreshToken(services.Auth))
		auth.POST("/forgot-password", HandleForgotPassword(services.Auth))
		auth.POST("/reset-password", HandleResetPassword(services.Auth))
		auth.POST("/verify-email", HandleVerifyEmail(services.Auth))
	}
}

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(router *gin.RouterGroup, services *services.Container) {
	users := router.Group("/users")
	users.Use(middleware.RequireAuth(services.Auth))
	{
		users.GET("/me", HandleGetCurrentUser(services.User))
		users.PUT("/me", HandleUpdateProfile(services.User))
		users.PUT("/me/password", HandleChangePassword(services.Auth))
		users.GET("/me/preferences", HandleGetPreferences(services.User))
		users.PUT("/me/preferences", HandleUpdatePreferences(services.User))
		users.GET("/me/tournaments", HandleGetUserTournaments(services.User))
		users.GET("/me/statistics", HandleGetUserStatistics(services.User))
	}
}

// RegisterTournamentRoutes registers tournament-related routes
func RegisterTournamentRoutes(router *gin.RouterGroup, services *services.Container) {
	tournaments := router.Group("/tournaments")
	{
		// Public routes
		tournaments.GET("", HandleListTournaments(services.Tournament))
		tournaments.GET("/:id", HandleGetTournament(services.Tournament))
		tournaments.GET("/:id/bracket", HandleGetBracket(services.Tournament, services.Match))
		tournaments.GET("/:id/schedule", HandleGetSchedule(services.Tournament, services.Match))
		tournaments.GET("/:id/participants", HandleGetParticipants(services.Tournament))
		tournaments.POST("/:id/register", middleware.OptionalAuth(services.Auth), HandleRegisterParticipant(services.Tournament))
		tournaments.POST("/:id/waitlist", middleware.OptionalAuth(services.Auth), HandleJoinWaitlist(services.Tournament))

		// Protected routes
		tournaments.Use(middleware.RequireAuth(services.Auth))
		tournaments.POST("", HandleCreateTournament(services.Tournament))
		tournaments.PUT("/:id", middleware.RequireTournamentOwner(services), HandleUpdateTournament(services.Tournament))
		tournaments.DELETE("/:id", middleware.RequireTournamentOwner(services), HandleDeleteTournament(services.Tournament))
		tournaments.POST("/:id/publish", middleware.RequireTournamentOwner(services), HandlePublishTournament(services.Tournament))
		tournaments.POST("/:id/start", middleware.RequireTournamentOwner(services), HandleStartTournament(services.Tournament))
		tournaments.POST("/:id/complete", middleware.RequireTournamentOwner(services), HandleCompleteTournament(services.Tournament))

		// Fixture generation
		tournaments.POST("/:id/fixtures/generate", middleware.RequireTournamentOwner(services), HandleGenerateFixtures(services.Tournament))
		tournaments.POST("/:id/schedule/auto", middleware.RequireTournamentOwner(services), HandleAutoSchedule(services.Tournament))

		// Venue management
		tournaments.GET("/:id/venues", HandleGetVenues(services.Tournament))
		tournaments.POST("/:id/venues", middleware.RequireTournamentOwner(services), HandleAddVenue(services.Tournament))
		tournaments.PUT("/:id/venues/:venueId", middleware.RequireTournamentOwner(services), HandleUpdateVenue(services.Tournament))
		tournaments.DELETE("/:id/venues/:venueId", middleware.RequireTournamentOwner(services), HandleDeleteVenue(services.Tournament))

		// Participant management
		tournaments.PUT("/:id/participants/:participantId", middleware.RequireTournamentOwner(services), HandleUpdateParticipant(services.Tournament))
		tournaments.DELETE("/:id/participants/:participantId", middleware.RequireTournamentOwner(services), HandleRemoveParticipant(services.Tournament))
		tournaments.POST("/:id/participants/:participantId/checkin", middleware.RequireTournamentOwner(services), HandleCheckInParticipant(services.Tournament))
	}
}

// RegisterMatchRoutes registers match-related routes
func RegisterMatchRoutes(router *gin.RouterGroup, services *services.Container) {
	matches := router.Group("/matches")
	matches.Use(middleware.RequireAuth(services.Auth))
	{
		matches.GET("/:id", HandleGetMatch(services.Match))
		matches.PUT("/:id", middleware.RequireMatchAccess(services), HandleUpdateMatch(services.Match))
		matches.POST("/:id/start", middleware.RequireMatchAccess(services), HandleStartMatch(services.Match))
		matches.POST("/:id/score", middleware.RequireMatchAccess(services), HandleReportScore(services.Match))
		matches.POST("/:id/cancel", middleware.RequireMatchAccess(services), HandleCancelMatch(services.Match))
	}
}

// RegisterPaymentRoutes registers payment-related routes
func RegisterPaymentRoutes(router *gin.RouterGroup, services *services.Container, cfg *config.Config) {
	if !cfg.Features.EnablePayments {
		return
	}

	payments := router.Group("/payments")
	payments.Use(middleware.RequireAuth(services.Auth))
	{
		payments.POST("/process", HandleProcessPayment(services.Payment))
		payments.POST("/refund", HandleRefundPayment(services.Payment))
		payments.POST("/webhook", HandleStripeWebhook(services.Payment, cfg))
	}
}

// RegisterAdminRoutes registers admin-only routes
func RegisterAdminRoutes(router *gin.RouterGroup, services *services.Container) {
	admin := router.Group("/admin")
	admin.Use(middleware.RequireAuth(services.Auth))
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/stats", HandleGetPlatformStats(services.Analytics))
		admin.GET("/users", HandleListUsers(services.User))
		admin.PUT("/users/:id/role", HandleUpdateUserRole(services.User))
		admin.GET("/tournaments", HandleListAllTournaments(services.Tournament))
		admin.DELETE("/tournaments/:id", HandleForceDeleteTournament(services.Tournament))
	}
}
