// internal/repositories/venue_repository.go
// Venue data access layer

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"tournament-planner/internal/models"
)

// VenueRepository handles venue data access
type VenueRepository struct {
	db *sql.DB
}

// NewVenueRepository creates a new venue repository
func NewVenueRepository(db *sql.DB) *VenueRepository {
	return &VenueRepository{db: db}
}

// Create inserts a new venue
func (r *VenueRepository) Create(ctx context.Context, venue *models.Venue) error {
	query := `
		INSERT INTO venues (
			id, tournament_id, name, type, availability_rules, is_active, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		venue.ID,
		venue.TournamentID,
		venue.Name,
		venue.Type,
		venue.AvailabilityRules,
		venue.IsActive,
		venue.CreatedAt,
	)

	return err
}

// CreateWithTx creates a venue within a transaction
func (r *VenueRepository) CreateWithTx(tx *sql.Tx, venue *models.Venue) error {
	query := `
		INSERT INTO venues (
			id, tournament_id, name, type, availability_rules, is_active, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.ExecContext(context.Background(), query,
		venue.ID,
		venue.TournamentID,
		venue.Name,
		venue.Type,
		venue.AvailabilityRules,
		venue.IsActive,
		venue.CreatedAt,
	)

	return err
}

// GetByID retrieves a venue by ID
func (r *VenueRepository) GetByID(ctx context.Context, id string) (*models.Venue, error) {
	query := `
		SELECT id, tournament_id, name, type, availability_rules, is_active, created_at
		FROM venues
		WHERE id = ?
	`

	var venue models.Venue
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&venue.ID,
		&venue.TournamentID,
		&venue.Name,
		&venue.Type,
		&venue.AvailabilityRules,
		&venue.IsActive,
		&venue.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("venue not found")
	}

	return &venue, err
}

// GetByTournamentID retrieves all venues for a tournament
func (r *VenueRepository) GetByTournamentID(ctx context.Context, tournamentID string) ([]*models.Venue, error) {
	query := `
		SELECT id, tournament_id, name, type, availability_rules, is_active, created_at
		FROM venues
		WHERE tournament_id = ? AND is_active = TRUE
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	venues := make([]*models.Venue, 0)
	for rows.Next() {
		var v models.Venue
		err := rows.Scan(
			&v.ID,
			&v.TournamentID,
			&v.Name,
			&v.Type,
			&v.AvailabilityRules,
			&v.IsActive,
			&v.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		venues = append(venues, &v)
	}

	return venues, nil
}

// Update updates venue information
func (r *VenueRepository) Update(ctx context.Context, venue *models.Venue) error {
	query := `
		UPDATE venues SET
			name = ?, type = ?, availability_rules = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		venue.Name,
		venue.Type,
		venue.AvailabilityRules,
		venue.ID,
	)

	return err
}

// Delete soft deletes a venue
func (r *VenueRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE venues SET is_active = FALSE WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// CountByTournamentID counts active venues for a tournament
func (r *VenueRepository) CountByTournamentID(ctx context.Context, tournamentID string) (int, error) {
	query := `SELECT COUNT(*) FROM venues WHERE tournament_id = ? AND is_active = TRUE`

	var count int
	err := r.db.QueryRowContext(ctx, query, tournamentID).Scan(&count)
	return count, err
}
