// internal/services/auth_service.go
// Authentication and authorization service

package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"tournament-planner/internal/config"
	"tournament-planner/internal/models"
	"tournament-planner/internal/repositories"
	"tournament-planner/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication and authorization
type AuthService struct {
	userRepo *repositories.UserRepository
	config   config.AuthConfig
	cache    *CacheService
	logger   *log.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo *repositories.UserRepository,
	config config.AuthConfig,
	cache *CacheService,
	logger *log.Logger,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   config,
		cache:    cache,
		logger:   logger,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req models.RegisterRequest) (*models.User, *models.TokenPair, error) {
	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BCryptCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:           utils.GenerateUUID(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		Phone:        &req.Phone,
		Role:         models.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Send verification email (async)
	go s.sendVerificationEmail(user)

	// Clear password hash from response
	user.PasswordHash = ""

	return user, tokenPair, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, *models.TokenPair, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokenPair, err := s.generateTokenPair(user)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login
	go s.userRepo.UpdateLastLogin(context.Background(), user.ID)

	// Clear password hash from response
	user.PasswordHash = ""

	return user, tokenPair, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	// Check if refresh token exists in cache
	cacheKey := fmt.Sprintf("refresh_token_%s", refreshToken)
	var userID string
	if err := s.cache.Get(cacheKey, &userID); err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Delete old refresh token
	s.cache.Delete(cacheKey)

	// Generate new token pair
	return s.generateTokenPair(user)
}

// generateTokenPair creates access and refresh tokens
func (s *AuthService) generateTokenPair(user *models.User) (*models.TokenPair, error) {
	// Generate access token
	accessToken, err := utils.GenerateJWT(user.ID, string(user.Role), s.config.JWTSecret, s.config.JWTExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in cache
	cacheKey := fmt.Sprintf("refresh_token_%s", refreshToken)
	if err := s.cache.Set(cacheKey, user.ID, s.config.RefreshTokenExpiry); err != nil {
		return nil, fmt.Errorf("failed to cache refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.config.JWTExpiration),
	}, nil
}

// ValidateToken validates a JWT token and returns the user ID and role
func (s *AuthService) ValidateToken(token string) (string, string, error) {
	userID, role, err := utils.ValidateJWT(token, s.config.JWTSecret)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	return userID, role, nil
}

// Logout invalidates a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken != "" {
		cacheKey := fmt.Sprintf("refresh_token_%s", refreshToken)
		s.cache.Delete(cacheKey)
	}
	return nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all refresh tokens for this user
	// In a production system, you'd track all refresh tokens per user

	return nil
}

// sendVerificationEmail sends an email verification link
func (s *AuthService) sendVerificationEmail(user *models.User) {
	// TODO: Implement email sending
	s.logger.Printf("Would send verification email to %s", user.Email)
}

// VerifyEmail marks a user's email as verified
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// TODO: Implement email verification token logic
	// For now, this is a placeholder
	return nil
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	// Check if user exists
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not
		return nil
	}

	// Generate reset token
	resetToken := utils.GenerateSecureToken()

	// Store reset token in cache with expiry
	cacheKey := fmt.Sprintf("password_reset_%s", resetToken)
	if err := s.cache.Set(cacheKey, user.ID, 1*time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send reset email (async)
	go s.sendPasswordResetEmail(user, resetToken)

	return nil
}

// sendPasswordResetEmail sends password reset email
func (s *AuthService) sendPasswordResetEmail(user *models.User, token string) {
	// TODO: Implement email sending
	s.logger.Printf("Would send password reset email to %s with token %s", user.Email, token)
}

// ResetPassword resets a user's password using a reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// Check if reset token is valid
	cacheKey := fmt.Sprintf("password_reset_%s", token)
	var userID string
	if err := s.cache.Get(cacheKey, &userID); err != nil {
		return ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Delete reset token
	s.cache.Delete(cacheKey)

	return nil
}
