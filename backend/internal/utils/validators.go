// internal/utils/validators.go
// Validation utility functions

package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"time"
)

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// ValidatePhone validates a phone number (basic validation)
func ValidatePhone(phone string) error {
	// Basic phone validation - in production, use a proper library
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return fmt.Errorf("invalid phone format")
	}
	return nil
}

// ValidateDateRange validates that start date is before end date
func ValidateDateRange(start, end time.Time) error {
	if start.After(end) {
		return fmt.Errorf("start date must be before end date")
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one number
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	return nil
}

// ValidateTournamentName validates tournament name
func ValidateTournamentName(name string) error {
	if len(name) < 3 {
		return fmt.Errorf("tournament name must be at least 3 characters long")
	}
	if len(name) > 255 {
		return fmt.Errorf("tournament name must not exceed 255 characters")
	}
	return nil
}

// ValidateTimezone validates timezone string
func ValidateTimezone(tz string) error {
	_, err := time.LoadLocation(tz)
	if err != nil {
		return fmt.Errorf("invalid timezone")
	}
	return nil
}
