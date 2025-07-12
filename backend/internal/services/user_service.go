// internal/services/user_service.go
// User profile and preferences management

package services

import (
	"context"
	"fmt"
	"log"

	"tournament-planner/internal/models"
	"tournament-planner/internal/repositories"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo        *repositories.UserRepository
	preferencesRepo *repositories.UserPreferencesRepository
	logger          *log.Logger
}

// NewUserService creates a new user service
func NewUserService(
	userRepo *repositories.UserRepository,
	preferencesRepo *repositories.UserPreferencesRepository,
	logger *log.Logger,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		preferencesRepo: preferencesRepo,
		logger:          logger,
	}
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Don't expose password hash
	user.PasswordHash = ""

	return user, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(ctx context.Context, userID string, updates map[string]interface{}) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if fullName, ok := updates["full_name"].(string); ok && fullName != "" {
		user.FullName = fullName
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = &phone
	}

	// Save updates
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Don't expose password hash
	user.PasswordHash = ""

	return user, nil
}

// GetPreferences retrieves user preferences
func (s *UserService) GetPreferences(ctx context.Context, userID string) (map[string]interface{}, error) {
	prefs, err := s.preferencesRepo.Get(ctx, userID)
	if err != nil {
		// Return default preferences if none exist
		return s.getDefaultPreferences(), nil
	}

	return prefs, nil
}

// UpdatePreferences updates user preferences
func (s *UserService) UpdatePreferences(ctx context.Context, userID string, preferences map[string]interface{}) error {
	return s.preferencesRepo.Set(ctx, userID, preferences)
}

// getDefaultPreferences returns default user preferences
func (s *UserService) getDefaultPreferences() map[string]interface{} {
	return map[string]interface{}{
		"notifications": map[string]bool{
			"email": true,
			"push":  true,
			"sms":   false,
		},
		"theme":    "light",
		"language": "en",
		"timezone": "UTC",
	}
}

// GetTournamentHistory retrieves user's tournament participation history
func (s *UserService) GetTournamentHistory(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	// TODO: Implement tournament history retrieval
	// This would join participants, tournament_participants, and tournaments tables
	return []map[string]interface{}{}, nil
}

// GetStatistics retrieves user statistics
func (s *UserService) GetStatistics(ctx context.Context, userID string) (map[string]interface{}, error) {
	// TODO: Implement statistics retrieval
	// This would aggregate data from matches and tournaments
	return map[string]interface{}{
		"tournaments_played": 0,
		"matches_played":     0,
		"matches_won":        0,
		"win_rate":           0.0,
	}, nil
}

// UpgradeToOrganizer upgrades a user to organizer role
func (s *UserService) UpgradeToOrganizer(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Role != models.RoleUser {
		return fmt.Errorf("user is already an organizer or admin")
	}

	user.Role = models.RoleOrganizer

	return s.userRepo.Update(ctx, user)
}
