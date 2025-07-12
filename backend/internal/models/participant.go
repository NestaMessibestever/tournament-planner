// internal/models/participant.go
// Participant (player/team) related models

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Participant represents a tournament participant (individual or team)
type Participant struct {
	ID                 string          `json:"id" db:"id"`
	UserID             *string         `json:"user_id,omitempty" db:"user_id"`
	Name               string          `json:"name" db:"name"`
	Type               ParticipantType `json:"type" db:"type"`
	ContactEmail       *string         `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone       *string         `json:"contact_phone,omitempty" db:"contact_phone"`
	TotalMatchesPlayed int             `json:"total_matches_played" db:"total_matches_played"`
	TotalMatchesWon    int             `json:"total_matches_won" db:"total_matches_won"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`

	// Tournament-specific fields (populated from junction table)
	Seed             *int                   `json:"seed,omitempty" db:"seed"`
	Division         *string                `json:"division,omitempty" db:"division"`
	GroupName        *string                `json:"group_name,omitempty" db:"group_name"`
	PaymentStatus    *PaymentStatus         `json:"payment_status,omitempty" db:"payment_status"`
	CheckedIn        *bool                  `json:"checked_in,omitempty" db:"checked_in"`
	RegistrationData map[string]interface{} `json:"registration_data,omitempty" db:"registration_data"`
}

// ParticipantType defines whether a participant is an individual or team
type ParticipantType string

const (
	ParticipantIndividual ParticipantType = "individual"
	ParticipantTeam       ParticipantType = "team"
)

// PaymentStatus represents the payment state for a participant
type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "pending"
	PaymentPaid     PaymentStatus = "paid"
	PaymentRefunded PaymentStatus = "refunded"
	PaymentWaived   PaymentStatus = "waived"
)

// Implement sql.Scanner and driver.Valuer for RegistrationData
func (r *map[string]interface{}) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into RegistrationData", value)
	}
	return json.Unmarshal(bytes, r)
}

func (r map[string]interface{}) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	}
	return json.Marshal(r)
}
