// internal/repositories/tournament_repository.go
// Tournament data access layer

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"tournament-planner/internal/models"
)

// TournamentRepository handles tournament data access
type TournamentRepository struct {
	db *sql.DB
}

// NewTournamentRepository creates a new tournament repository
func NewTournamentRepository(db *sql.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

// Create inserts a new tournament
func (r *TournamentRepository) Create(ctx context.Context, tournament *models.Tournament) error {
	query := `
		INSERT INTO tournaments (
			id, organizer_id, name, description, sport_id, format_type,
			format_config, start_date, end_date, timezone, max_matches_per_day,
			operational_hours, avg_match_duration, buffer_time, registration_deadline,
			entry_fee, allow_onsite_payment, capacity_limit, current_participants,
			status, is_public, custom_fields, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	// Convert custom fields to JSON
	customFieldsJSON, err := json.Marshal(tournament.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tournament.ID,
		tournament.OrganizerID,
		tournament.Name,
		tournament.Description,
		tournament.SportID,
		tournament.FormatType,
		tournament.FormatConfig,
		tournament.StartDate,
		tournament.EndDate,
		tournament.Timezone,
		tournament.MaxMatchesPerDay,
		tournament.OperationalHours,
		tournament.AvgMatchDuration,
		tournament.BufferTime,
		tournament.RegistrationDeadline,
		tournament.EntryFee,
		tournament.AllowOnsitePayment,
		tournament.CapacityLimit,
		tournament.CurrentParticipants,
		tournament.Status,
		tournament.IsPublic,
		customFieldsJSON,
		tournament.CreatedAt,
		tournament.UpdatedAt,
	)

	return err
}

// CreateWithTx creates a tournament within a transaction
func (r *TournamentRepository) CreateWithTx(tx *sql.Tx, tournament *models.Tournament) error {
	query := `
		INSERT INTO tournaments (
			id, organizer_id, name, description, sport_id, format_type,
			format_config, start_date, end_date, timezone, max_matches_per_day,
			operational_hours, avg_match_duration, buffer_time, registration_deadline,
			entry_fee, allow_onsite_payment, capacity_limit, current_participants,
			status, is_public, custom_fields, created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	customFieldsJSON, err := json.Marshal(tournament.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	_, err = tx.ExecContext(context.Background(), query,
		tournament.ID,
		tournament.OrganizerID,
		tournament.Name,
		tournament.Description,
		tournament.SportID,
		tournament.FormatType,
		tournament.FormatConfig,
		tournament.StartDate,
		tournament.EndDate,
		tournament.Timezone,
		tournament.MaxMatchesPerDay,
		tournament.OperationalHours,
		tournament.AvgMatchDuration,
		tournament.BufferTime,
		tournament.RegistrationDeadline,
		tournament.EntryFee,
		tournament.AllowOnsitePayment,
		tournament.CapacityLimit,
		tournament.CurrentParticipants,
		tournament.Status,
		tournament.IsPublic,
		customFieldsJSON,
		tournament.CreatedAt,
		tournament.UpdatedAt,
	)

	return err
}

// GetByID retrieves a tournament by ID
func (r *TournamentRepository) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	query := `
		SELECT 
			id, organizer_id, name, description, sport_id, format_type,
			format_config, start_date, end_date, timezone, max_matches_per_day,
			operational_hours, avg_match_duration, buffer_time, registration_deadline,
			entry_fee, allow_onsite_payment, capacity_limit, current_participants,
			status, is_public, custom_fields, created_at, updated_at
		FROM tournaments
		WHERE id = ?
	`

	var tournament models.Tournament
	var customFieldsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tournament.ID,
		&tournament.OrganizerID,
		&tournament.Name,
		&tournament.Description,
		&tournament.SportID,
		&tournament.FormatType,
		&tournament.FormatConfig,
		&tournament.StartDate,
		&tournament.EndDate,
		&tournament.Timezone,
		&tournament.MaxMatchesPerDay,
		&tournament.OperationalHours,
		&tournament.AvgMatchDuration,
		&tournament.BufferTime,
		&tournament.RegistrationDeadline,
		&tournament.EntryFee,
		&tournament.AllowOnsitePayment,
		&tournament.CapacityLimit,
		&tournament.CurrentParticipants,
		&tournament.Status,
		&tournament.IsPublic,
		&customFieldsJSON,
		&tournament.CreatedAt,
		&tournament.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tournament not found")
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal custom fields
	if len(customFieldsJSON) > 0 {
		if err := json.Unmarshal(customFieldsJSON, &tournament.CustomFields); err != nil {
			return nil, fmt.Errorf("failed to unmarshal custom fields: %w", err)
		}
	}

	return &tournament, nil
}

// GetByIDWithDetails retrieves a tournament with all related data
func (r *TournamentRepository) GetByIDWithDetails(ctx context.Context, id string) (*models.Tournament, error) {
	// First get the tournament
	tournament, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: Load related data (venues, participants count, etc.)
	// This would involve additional queries or a more complex JOIN query

	return tournament, nil
}

// Update updates a tournament
func (r *TournamentRepository) Update(ctx context.Context, tournament *models.Tournament) error {
	query := `
		UPDATE tournaments SET
			name = ?, description = ?, sport_id = ?, format_type = ?,
			format_config = ?, start_date = ?, end_date = ?, timezone = ?,
			max_matches_per_day = ?, operational_hours = ?, avg_match_duration = ?,
			buffer_time = ?, registration_deadline = ?, entry_fee = ?,
			allow_onsite_payment = ?, capacity_limit = ?, status = ?,
			is_public = ?, custom_fields = ?, updated_at = NOW()
		WHERE id = ?
	`

	customFieldsJSON, err := json.Marshal(tournament.CustomFields)
	if err != nil {
		return fmt.Errorf("failed to marshal custom fields: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tournament.Name,
		tournament.Description,
		tournament.SportID,
		tournament.FormatType,
		tournament.FormatConfig,
		tournament.StartDate,
		tournament.EndDate,
		tournament.Timezone,
		tournament.MaxMatchesPerDay,
		tournament.OperationalHours,
		tournament.AvgMatchDuration,
		tournament.BufferTime,
		tournament.RegistrationDeadline,
		tournament.EntryFee,
		tournament.AllowOnsitePayment,
		tournament.CapacityLimit,
		tournament.Status,
		tournament.IsPublic,
		customFieldsJSON,
		tournament.ID,
	)

	return err
}

// List retrieves tournaments with pagination and filters
func (r *TournamentRepository) List(ctx context.Context, filter ListFilter) ([]*models.Tournament, int, error) {
	// Build dynamic query based on filters
	var conditions []string
	var args []interface{}

	// Base query
	baseQuery := "FROM tournaments WHERE 1=1"

	// Apply filters
	if filter.OrganizerID != "" {
		conditions = append(conditions, "organizer_id = ?")
		args = append(args, filter.OrganizerID)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Public {
		conditions = append(conditions, "is_public = TRUE")
	}
	if filter.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Add conditions to base query
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build select query with pagination
	selectQuery := `
		SELECT 
			id, organizer_id, name, description, sport_id, format_type,
			format_config, start_date, end_date, timezone, max_matches_per_day,
			operational_hours, avg_match_duration, buffer_time, registration_deadline,
			entry_fee, allow_onsite_payment, capacity_limit, current_participants,
			status, is_public, custom_fields, created_at, updated_at
		` + baseQuery + " ORDER BY created_at DESC LIMIT ? OFFSET ?"

	// Add pagination args
	args = append(args, filter.Limit, (filter.Page-1)*filter.Limit)

	// Execute query
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Parse results
	tournaments := make([]*models.Tournament, 0)
	for rows.Next() {
		var t models.Tournament
		var customFieldsJSON []byte

		err := rows.Scan(
			&t.ID, &t.OrganizerID, &t.Name, &t.Description, &t.SportID,
			&t.FormatType, &t.FormatConfig, &t.StartDate, &t.EndDate,
			&t.Timezone, &t.MaxMatchesPerDay, &t.OperationalHours,
			&t.AvgMatchDuration, &t.BufferTime, &t.RegistrationDeadline,
			&t.EntryFee, &t.AllowOnsitePayment, &t.CapacityLimit,
			&t.CurrentParticipants, &t.Status, &t.IsPublic,
			&customFieldsJSON, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Unmarshal custom fields
		if len(customFieldsJSON) > 0 {
			if err := json.Unmarshal(customFieldsJSON, &t.CustomFields); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal custom fields: %w", err)
			}
		}

		tournaments = append(tournaments, &t)
	}

	return tournaments, total, nil
}

// UpdateStatusWithTx updates tournament status within a transaction
func (r *TournamentRepository) UpdateStatusWithTx(tx *sql.Tx, id string, status models.TournamentStatus) error {
	query := `UPDATE tournaments SET status = ?, updated_at = NOW() WHERE id = ?`
	_, err := tx.ExecContext(context.Background(), query, status, id)
	return err
}

// IncrementParticipants increments the participant count
func (r *TournamentRepository) IncrementParticipants(ctx context.Context, id string) error {
	query := `UPDATE tournaments SET current_participants = current_participants + 1 WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// DecrementParticipants decrements the participant count
func (r *TournamentRepository) DecrementParticipants(ctx context.Context, id string) error {
	query := `UPDATE tournaments SET current_participants = current_participants - 1 WHERE id = ? AND current_participants > 0`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListFilter defines filtering options for tournament queries
type ListFilter struct {
	Page        int
	Limit       int
	OrganizerID string
	Status      string
	Public      bool
	Search      string
}
