// internal/services/container.go
// Service container provides dependency injection for all business logic services.
// This pattern makes testing easier and keeps services loosely coupled.

package services

import (
	"errors"
	"log"

	"tournament-planner/internal/config"
	"tournament-planner/internal/database"
	"tournament-planner/internal/repositories"
)

// Container holds all service instances and provides them to handlers
type Container struct {
	Auth         *AuthService
	User         *UserService
	Tournament   *TournamentService
	Match        *MatchService
	Payment      *PaymentService
	Notification *NotificationService
	Cache        *CacheService
	Analytics    *AnalyticsService
}

// NewContainer creates a new service container with all dependencies
func NewContainer(db *database.Connections, cfg *config.Config, logger *log.Logger) *Container {
	// Initialize repositories
	repos := repositories.NewContainer(db)

	// Initialize cache service
	cache := NewCacheService(db.Redis, logger)

	// Initialize notification service
	notification := NewNotificationService(db, cfg, logger)

	// Initialize services with their dependencies
	auth := NewAuthService(repos.User, cfg.Auth, cache, logger)
	user := NewUserService(repos.User, repos.UserPreferences, logger)
	tournament := NewTournamentService(repos, cache, notification, logger)
	match := NewMatchService(repos, cache, notification, logger)
	payment := NewPaymentService(repos, cfg.External, logger)
	analytics := NewAnalyticsService(db.MongoDB, cache, logger)

	return &Container{
		Auth:         auth,
		User:         user,
		Tournament:   tournament,
		Match:        match,
		Payment:      payment,
		Notification: notification,
		Cache:        cache,
		Analytics:    analytics,
	}
}

// Common errors used across services
var (
	ErrNotFound                 = errors.New("resource not found")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrForbidden                = errors.New("forbidden")
	ErrInvalidInput             = errors.New("invalid input")
	ErrEmailAlreadyExists       = errors.New("email already exists")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrInvalidToken             = errors.New("invalid token")
	ErrInsufficientParticipants = errors.New("insufficient participants")
	ErrCapacityExceeded         = errors.New("capacity exceeded")
	ErrNoVenues                 = errors.New("no venues available")
	ErrSchedulingImpossible     = errors.New("scheduling impossible with current constraints")
	ErrTournamentFull           = errors.New("tournament is full")
	ErrRegistrationClosed       = errors.New("registration is closed")
	ErrAlreadyRegistered        = errors.New("already registered for this tournament")
	ErrPaymentRequired          = errors.New("payment required")
	ErrInvalidFormat            = errors.New("invalid tournament format")
)
