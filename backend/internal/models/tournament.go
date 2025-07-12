// internal/models/tournament.go
// Domain models representing core business entities

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Tournament represents a tournament with all its configuration
type Tournament struct {
	ID                   string           `json:"id" db:"id"`
	OrganizerID          string           `json:"organizer_id" db:"organizer_id"`
	Name                 string           `json:"name" db:"name"`
	Description          string           `json:"description" db:"description"`
	SportID              *string          `json:"sport_id,omitempty" db:"sport_id"`
	FormatType           TournamentFormat `json:"format_type" db:"format_type"`
	FormatConfig         *FormatConfig    `json:"format_config,omitempty" db:"format_config"`
	StartDate            time.Time        `json:"start_date" db:"start_date"`
	EndDate              time.Time        `json:"end_date" db:"end_date"`
	Timezone             string           `json:"timezone" db:"timezone"`
	MaxMatchesPerDay     int              `json:"max_matches_per_day" db:"max_matches_per_day"`
	OperationalHours     OperationalHours `json:"operational_hours" db:"operational_hours"`
	AvgMatchDuration     int              `json:"avg_match_duration" db:"avg_match_duration"`
	BufferTime           int              `json:"buffer_time" db:"buffer_time"`
	RegistrationDeadline *time.Time       `json:"registration_deadline,omitempty" db:"registration_deadline"`
	EntryFee             float64          `json:"entry_fee" db:"entry_fee"`
	AllowOnsitePayment   bool             `json:"allow_onsite_payment" db:"allow_onsite_payment"`
	CapacityLimit        int              `json:"capacity_limit" db:"capacity_limit"`
	CurrentParticipants  int              `json:"current_participants" db:"current_participants"`
	Status               TournamentStatus `json:"status" db:"status"`
	IsPublic             bool             `json:"is_public" db:"is_public"`
	CustomFields         []CustomField    `json:"custom_fields,omitempty" db:"custom_fields"`
	CreatedAt            time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at" db:"updated_at"`
}

// TournamentFormat represents different tournament formats
type TournamentFormat string

const (
	FormatSingleElimination TournamentFormat = "single_elimination"
	FormatDoubleElimination TournamentFormat = "double_elimination"
	FormatRoundRobin        TournamentFormat = "round_robin"
	FormatSwiss             TournamentFormat = "swiss"
	FormatGroupToKnockout   TournamentFormat = "group_to_knockout"
)

// TournamentStatus represents the current state of a tournament
type TournamentStatus string

const (
	StatusDraft              TournamentStatus = "draft"
	StatusPublished          TournamentStatus = "published"
	StatusRegistrationOpen   TournamentStatus = "registration_open"
	StatusRegistrationClosed TournamentStatus = "registration_closed"
	StatusInProgress         TournamentStatus = "in_progress"
	StatusCompleted          TournamentStatus = "completed"
	StatusCancelled          TournamentStatus = "cancelled"
)

// FormatConfig stores format-specific configuration
type FormatConfig struct {
	NumberOfGroups  int    `json:"number_of_groups,omitempty"`
	GroupSize       int    `json:"group_size,omitempty"`
	AdvancementRule string `json:"advancement_rule,omitempty"`
	Consolation     bool   `json:"consolation,omitempty"`
	ThirdPlaceMatch bool   `json:"third_place_match,omitempty"`
	NumberOfRounds  int    `json:"number_of_rounds,omitempty"`
}

// OperationalHours defines when the tournament can run each day
type OperationalHours map[string]DayHours

// DayHours represents operational hours for a single day
type DayHours struct {
	StartTime string `json:"start_time"` // Format: "09:00"
	EndTime   string `json:"end_time"`   // Format: "18:00"
}

// CustomField represents a custom registration field
type CustomField struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Type       string   `json:"type"` // text, number, select, checkbox, date, file
	Required   bool     `json:"required"`
	Options    []string `json:"options,omitempty"`
	Validation string   `json:"validation,omitempty"`
}

// Implement sql.Scanner and driver.Valuer for JSON fields
func (f *FormatConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into FormatConfig", value)
	}
	return json.Unmarshal(bytes, f)
}

func (f FormatConfig) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (o *OperationalHours) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into OperationalHours", value)
	}
	return json.Unmarshal(bytes, o)
}

func (o OperationalHours) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (c *CustomField) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into CustomField", value)
	}
	return json.Unmarshal(bytes, c)
}

func (c CustomField) Value() (driver.Value, error) {
	return json.Marshal(c)
}
