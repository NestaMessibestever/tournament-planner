// internal/repositories/tournament_participant_repository.go
// Tournament participant junction table data access

package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	"tournament-planner/internal/models"
)

// TournamentParticipantRepository handles tournament-participant relationships
type TournamentParticipantRepository struct {
	db *sql.DB
}

// NewTournamentParticipantRepository creates a new repository
func NewTournamentParticipantRepository(db *sql.DB) *TournamentParticipantRepository {
	return &TournamentParticipantRepository{db: db}
}

// Create adds a participant to a tournament
func (r *TournamentParticipantRepository) Create(ctx context.Context, tournamentID, participantID string, data map[string]interface{}) error {
	registrationDataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO tournament_participants (
			tournament_id, participant_id, payment_status, registration_data, registered_at
		) VALUES (?, ?, 'pending', ?, NOW())
	`

	_, err = r.db.ExecContext(ctx, query, tournamentID, participantID, registrationDataJSON)
	return err
}

// GetByTournamentID retrieves all participants for a tournament
func (r *TournamentParticipantRepository) GetByTournamentID(ctx context.Context, tournamentID string) ([]*models.Participant, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.name, p.type, p.contact_email, p.contact_phone,
			p.total_matches_played, p.total_matches_won, p.created_at, p.updated_at,
			tp.seed, tp.division, tp.group_name, tp.payment_status, tp.checked_in,
			tp.registration_data
		FROM participants p
		JOIN tournament_participants tp ON p.id = tp.participant_id
		WHERE tp.tournament_id = ?
		ORDER BY tp.seed, p.name
	`

	rows, err := r.db.QueryContext(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	participants := make([]*models.Participant, 0)
	for rows.Next() {
		var p models.Participant
		var registrationDataJSON []byte

		err := rows.Scan(
			&p.ID, &p.UserID, &p.Name, &p.Type, &p.ContactEmail,
			&p.ContactPhone, &p.TotalMatchesPlayed, &p.TotalMatchesWon,
			&p.CreatedAt, &p.UpdatedAt, &p.Seed, &p.Division,
			&p.GroupName, &p.PaymentStatus, &p.CheckedIn,
			&registrationDataJSON,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal registration data
		if len(registrationDataJSON) > 0 {
			if err := json.Unmarshal(registrationDataJSON, &p.RegistrationData); err != nil {
				return nil, err
			}
		}

		participants = append(participants, &p)
	}

	return participants, nil
}

// UpdateSeed updates participant seeding
func (r *TournamentParticipantRepository) UpdateSeed(ctx context.Context, tournamentID, participantID string, seed int) error {
	query := `
		UPDATE tournament_participants 
		SET seed = ? 
		WHERE tournament_id = ? AND participant_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, seed, tournamentID, participantID)
	return err
}

// UpdatePaymentStatus updates payment status
func (r *TournamentParticipantRepository) UpdatePaymentStatus(ctx context.Context, tournamentID, participantID string, status models.PaymentStatus) error {
	query := `
		UPDATE tournament_participants 
		SET payment_status = ? 
		WHERE tournament_id = ? AND participant_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, status, tournamentID, participantID)
	return err
}

// Delete removes a participant from a tournament
func (r *TournamentParticipantRepository) Delete(ctx context.Context, tournamentID, participantID string) error {
	query := `DELETE FROM tournament_participants WHERE tournament_id = ? AND participant_id = ?`
	_, err := r.db.ExecContext(ctx, query, tournamentID, participantID)
	return err
}

// CheckIn marks a participant as checked in
func (r *TournamentParticipantRepository) CheckIn(ctx context.Context, tournamentID, participantID string) error {
	query := `
		UPDATE tournament_participants 
		SET checked_in = TRUE 
		WHERE tournament_id = ? AND participant_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, tournamentID, participantID)
	return err
}
