// internal/repositories/match_repository.go
// Match data access layer

package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"tournament-planner/internal/models"
)

// MatchRepository handles match data access
type MatchRepository struct {
	db *sql.DB
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db *sql.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

// Create inserts a new match
func (r *MatchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `
		INSERT INTO matches (
			id, tournament_id, round_number, match_number, stage, group_name,
			participant1_id, participant2_id, winner_id, score1, score2,
			score_details, status, scheduled_datetime, actual_start_time,
			actual_end_time, venue_id, referee_id, next_match_id, notes,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		match.ID,
		match.TournamentID,
		match.RoundNumber,
		match.MatchNumber,
		match.Stage,
		match.GroupName,
		match.Participant1ID,
		match.Participant2ID,
		match.WinnerID,
		match.Score1,
		match.Score2,
		match.ScoreDetails,
		match.Status,
		match.ScheduledDatetime,
		match.ActualStartTime,
		match.ActualEndTime,
		match.VenueID,
		match.RefereeID,
		match.NextMatchID,
		match.Notes,
		match.CreatedAt,
		match.UpdatedAt,
	)

	return err
}

// CreateWithTx creates a match within a transaction
func (r *MatchRepository) CreateWithTx(tx *sql.Tx, match *models.Match) error {
	query := `
		INSERT INTO matches (
			id, tournament_id, round_number, match_number, stage, group_name,
			participant1_id, participant2_id, winner_id, score1, score2,
			score_details, status, scheduled_datetime, actual_start_time,
			actual_end_time, venue_id, referee_id, next_match_id, notes,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
		)
	`

	_, err := tx.ExecContext(context.Background(), query,
		match.ID,
		match.TournamentID,
		match.RoundNumber,
		match.MatchNumber,
		match.Stage,
		match.GroupName,
		match.Participant1ID,
		match.Participant2ID,
		match.WinnerID,
		match.Score1,
		match.Score2,
		match.ScoreDetails,
		match.Status,
		match.ScheduledDatetime,
		match.ActualStartTime,
		match.ActualEndTime,
		match.VenueID,
		match.RefereeID,
		match.NextMatchID,
		match.Notes,
		match.CreatedAt,
		match.UpdatedAt,
	)

	return err
}

// GetByID retrieves a match by ID
func (r *MatchRepository) GetByID(ctx context.Context, id string) (*models.Match, error) {
	query := `
		SELECT 
			id, tournament_id, round_number, match_number, stage, group_name,
			participant1_id, participant2_id, winner_id, score1, score2,
			score_details, status, scheduled_datetime, actual_start_time,
			actual_end_time, venue_id, referee_id, next_match_id, notes,
			created_at, updated_at
		FROM matches
		WHERE id = ?
	`

	var match models.Match
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.TournamentID,
		&match.RoundNumber,
		&match.MatchNumber,
		&match.Stage,
		&match.GroupName,
		&match.Participant1ID,
		&match.Participant2ID,
		&match.WinnerID,
		&match.Score1,
		&match.Score2,
		&match.ScoreDetails,
		&match.Status,
		&match.ScheduledDatetime,
		&match.ActualStartTime,
		&match.ActualEndTime,
		&match.VenueID,
		&match.RefereeID,
		&match.NextMatchID,
		&match.Notes,
		&match.CreatedAt,
		&match.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("match not found")
	}

	return &match, err
}

// GetByTournamentID retrieves all matches for a tournament
func (r *MatchRepository) GetByTournamentID(ctx context.Context, tournamentID string) ([]*models.Match, error) {
	query := `
		SELECT 
			id, tournament_id, round_number, match_number, stage, group_name,
			participant1_id, participant2_id, winner_id, score1, score2,
			score_details, status, scheduled_datetime, actual_start_time,
			actual_end_time, venue_id, referee_id, next_match_id, notes,
			created_at, updated_at
		FROM matches
		WHERE tournament_id = ?
		ORDER BY round_number, match_number
	`

	rows, err := r.db.QueryContext(ctx, query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]*models.Match, 0)
	for rows.Next() {
		var m models.Match
		err := rows.Scan(
			&m.ID, &m.TournamentID, &m.RoundNumber, &m.MatchNumber,
			&m.Stage, &m.GroupName, &m.Participant1ID, &m.Participant2ID,
			&m.WinnerID, &m.Score1, &m.Score2, &m.ScoreDetails,
			&m.Status, &m.ScheduledDatetime, &m.ActualStartTime,
			&m.ActualEndTime, &m.VenueID, &m.RefereeID, &m.NextMatchID,
			&m.Notes, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, &m)
	}

	return matches, nil
}

// Update updates match information
func (r *MatchRepository) Update(ctx context.Context, match *models.Match) error {
	query := `
		UPDATE matches SET
			scheduled_datetime = ?, venue_id = ?, referee_id = ?,
			notes = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		match.ScheduledDatetime,
		match.VenueID,
		match.RefereeID,
		match.Notes,
		match.ID,
	)

	return err
}

// UpdateScore updates match score and status
func (r *MatchRepository) UpdateScore(ctx context.Context, id string, score1, score2 int, winnerID string, scoreDetails *models.ScoreDetails) error {
	query := `
		UPDATE matches SET
			score1 = ?, score2 = ?, winner_id = ?, score_details = ?,
			status = ?, actual_end_time = NOW(), updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		score1, score2, winnerID, scoreDetails,
		models.MatchCompleted, id,
	)

	return err
}

// UpdateStatus updates match status
func (r *MatchRepository) UpdateStatus(ctx context.Context, id string, status models.MatchStatus) error {
	query := `UPDATE matches SET status = ?, updated_at = NOW() WHERE id = ?`

	// Update actual start time if match is starting
	if status == models.MatchInProgress {
		query = `UPDATE matches SET status = ?, actual_start_time = NOW(), updated_at = NOW() WHERE id = ?`
	}

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// GetNextMatch retrieves the next match in the bracket
func (r *MatchRepository) GetNextMatch(ctx context.Context, matchID string) (*models.Match, error) {
	query := `
		SELECT m2.* FROM matches m1
		JOIN matches m2 ON m1.next_match_id = m2.id
		WHERE m1.id = ?
	`

	var match models.Match
	err := r.db.QueryRowContext(ctx, query, matchID).Scan(
		&match.ID, &match.TournamentID, &match.RoundNumber, &match.MatchNumber,
		&match.Stage, &match.GroupName, &match.Participant1ID, &match.Participant2ID,
		&match.WinnerID, &match.Score1, &match.Score2, &match.ScoreDetails,
		&match.Status, &match.ScheduledDatetime, &match.ActualStartTime,
		&match.ActualEndTime, &match.VenueID, &match.RefereeID, &match.NextMatchID,
		&match.Notes, &match.CreatedAt, &match.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &match, err
}

// ListByVenueAndDate retrieves matches for a specific venue and date
func (r *MatchRepository) ListByVenueAndDate(ctx context.Context, venueID string, date time.Time) ([]*models.Match, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT 
			id, tournament_id, round_number, match_number, stage, group_name,
			participant1_id, participant2_id, winner_id, score1, score2,
			score_details, status, scheduled_datetime, actual_start_time,
			actual_end_time, venue_id, referee_id, next_match_id, notes,
			created_at, updated_at
		FROM matches
		WHERE venue_id = ? AND scheduled_datetime >= ? AND scheduled_datetime < ?
		ORDER BY scheduled_datetime
	`

	rows, err := r.db.QueryContext(ctx, query, venueID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]*models.Match, 0)
	for rows.Next() {
		var m models.Match
		err := rows.Scan(
			&m.ID, &m.TournamentID, &m.RoundNumber, &m.MatchNumber,
			&m.Stage, &m.GroupName, &m.Participant1ID, &m.Participant2ID,
			&m.WinnerID, &m.Score1, &m.Score2, &m.ScoreDetails,
			&m.Status, &m.ScheduledDatetime, &m.ActualStartTime,
			&m.ActualEndTime, &m.VenueID, &m.RefereeID, &m.NextMatchID,
			&m.Notes, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		matches = append(matches, &m)
	}

	return matches, nil
}
