// internal/models/user.go
// User and authentication related models

package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID            string    `json:"id" db:"id"`
	Email         string    `json:"email" db:"email"`
	PasswordHash  string    `json:"-" db:"password_hash"` // Never expose in JSON
	FullName      string    `json:"full_name" db:"full_name"`
	Phone         *string   `json:"phone,omitempty" db:"phone"`
	Role          UserRole  `json:"role" db:"role"`
	EmailVerified bool      `json:"email_verified" db:"email_verified"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole defines user access levels
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleOrganizer UserRole = "organizer"
	RoleAdmin     UserRole = "admin"
)

// TokenPair represents JWT access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// LoginRequest represents authentication credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents new user registration data
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required,min=2,max=100"`
	Phone    string `json:"phone,omitempty" binding:"omitempty,e164"`
}
