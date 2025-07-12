// internal/models/match.go
// Match and fixture related models

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Match represents a tournament match/fixture
type Match struct {
	ID                string        `json:"id" db:"id"`
	TournamentID      string        `json:"tournament_id" db:"tournament_id"`
	RoundNumber       int           `json:"round_number" db:"round_number"`
	MatchNumber       int           `json:"match_number" db:"match_number"`
	Stage             string        `json:"stage" db:"stage"`
	GroupName         *string       `json:"group_name,omitempty" db:"group_name"`
	Participant1ID    *string       `json:"participant1_id,omitempty" db:"participant1_id"`
	Participant1      *Participant  `json:"participant1,omitempty"`
	Participant2ID    *string       `json:"participant2_id,omitempty" db:"participant2_id"`
	Participant2      *Participant  `json:"participant2,omitempty"`
	WinnerID          *string       `json:"winner_id,omitempty" db:"winner_id"`
	Score1            *int          `json:"score1,omitempty" db:"score1"`
	Score2            *int          `json:"score2,omitempty" db:"score2"`
	ScoreDetails      *ScoreDetails `json:"score_details,omitempty" db:"score_details"`
	Status            MatchStatus   `json:"status" db:"status"`
	ScheduledDatetime *time.Time    `json:"scheduled_datetime,omitempty" db:"scheduled_datetime"`
	ActualStartTime   *time.Time    `json:"actual_start_time,omitempty" db:"actual_start_time"`
	ActualEndTime     *time.Time    `json:"actual_end_time,omitempty" db:"actual_end_time"`
	VenueID           *string       `json:"venue_id,omitempty" db:"venue_id"`
	Venue             *Venue        `json:"venue,omitempty"`
	RefereeID         *string       `json:"referee_id,omitempty" db:"referee_id"`
	NextMatchID       *string       `json:"next_match_id,omitempty" db:"next_match_id"`
	Notes             *string       `json:"notes,omitempty" db:"notes"`
	CreatedAt         time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" db:"updated_at"`
}

// MatchStatus represents the current state of a match
type MatchStatus string

const (
	MatchPending    MatchStatus = "pending"
	MatchScheduled  MatchStatus = "scheduled"
	MatchInProgress MatchStatus = "in_progress"
	MatchCompleted  MatchStatus = "completed"
	MatchCancelled  MatchStatus = "cancelled"
	MatchPostponed  MatchStatus = "postponed"
	MatchWalkover   MatchStatus = "walkover"
)

// ScoreDetails stores sport-specific scoring information
type ScoreDetails struct {
	Sets   []SetScore             `json:"sets,omitempty"`
	Games  []GameScore            `json:"games,omitempty"`
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// SetScore represents a set score (for tennis, volleyball, etc.)
type SetScore struct {
	Player1Score int `json:"player1_score"`
	Player2Score int `json:"player2_score"`
}

// GameScore represents a game score
type GameScore struct {
	Player1Score int `json:"player1_score"`
	Player2Score int `json:"player2_score"`
}

// Implement sql.Scanner and driver.Valuer for ScoreDetails
func (s *ScoreDetails) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ScoreDetails", value)
	}
	return json.Unmarshal(bytes, s)
}

func (s ScoreDetails) Value() (driver.Value, error) {
	return json.Marshal(s)
}
