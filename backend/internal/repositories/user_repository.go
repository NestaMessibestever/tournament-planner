// internal/repositories/user_repository.go
// User data access layer

package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tournament-planner/internal/models"
)

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, password_hash, full_name, phone, role,
			email_verified, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Phone,
		user.Role,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, full_name, phone, role,
			email_verified, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Phone,
		&user.Role,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	return &user, err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT 
			id, email, password_hash, full_name, phone, role,
			email_verified, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Phone,
		&user.Role,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}

	return &user, err
}

// Update updates user information
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users SET
			full_name = ?, phone = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		user.FullName,
		user.Phone,
		time.Now(),
		user.ID,
	)

	return err
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, id string, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), id)
	return err
}

// UpdateEmailVerified marks email as verified
func (r *UserRepository) UpdateEmailVerified(ctx context.Context, id string) error {
	query := `UPDATE users SET email_verified = TRUE, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string) error {
	query := `UPDATE users SET updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	return exists, err
}
