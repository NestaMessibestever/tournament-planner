// internal/repositories/participant_repository.go
// Participant data access layer

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"tournament-planner/internal/models"
)

// ParticipantRepository handles participant data access
type ParticipantRepository struct {
	db *sql.DB
}

// NewParticipantRepository creates a new participant repository
func NewParticipantRepository(db *sql.DB) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

// Create inserts a new participant
func (r *ParticipantRepository) Create(ctx context.Context, participant *models.Participant) error {
	query := `
		INSERT INTO participants (
			id, user_id, name, type, contact_email, contact_phone,
			total_matches_played, total_matches_won, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		participant.ID,
		participant.UserID,
		participant.Name,
		participant.Type,
		participant.ContactEmail,
		participant.ContactPhone,
		participant.TotalMatchesPlayed,
		participant.TotalMatchesWon,
		participant.CreatedAt,
		participant.UpdatedAt,
	)

	return err
}

// GetByID retrieves a participant by ID
func (r *ParticipantRepository) GetByID(ctx context.Context, id string) (*models.Participant, error) {
	query := `
		SELECT 
			id, user_id, name, type, contact_email, contact_phone,
			total_matches_played, total_matches_won, created_at, updated_at
		FROM participants
		WHERE id = ?
	`

	var participant models.Participant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&participant.ID,
		&participant.UserID,
		&participant.Name,
		&participant.Type,
		&participant.ContactEmail,
		&participant.ContactPhone,
		&participant.TotalMatchesPlayed,
		&participant.TotalMatchesWon,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("participant not found")
	}

	return &participant, err
}

// GetByUserID retrieves a participant by user ID
func (r *ParticipantRepository) GetByUserID(ctx context.Context, userID string) (*models.Participant, error) {
	query := `
		SELECT 
			id, user_id, name, type, contact_email, contact_phone,
			total_matches_played, total_matches_won, created_at, updated_at
		FROM participants
		WHERE user_id = ?
	`

	var participant models.Participant
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&participant.ID,
		&participant.UserID,
		&participant.Name,
		&participant.Type,
		&participant.ContactEmail,
		&participant.ContactPhone,
		&participant.TotalMatchesPlayed,
		&participant.TotalMatchesWon,
		&participant.CreatedAt,
		&participant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &participant, err
}

// UpdateStats updates participant statistics
func (r *ParticipantRepository) UpdateStats(ctx context.Context, id string, matchesPlayed, matchesWon int) error {
	query := `
		UPDATE participants SET
			total_matches_played = total_matches_played + ?,
			total_matches_won = total_matches_won + ?,
			updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query, matchesPlayed, matchesWon, id)
	return err
}
