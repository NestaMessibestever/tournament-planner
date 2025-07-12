// internal/models/venue.go
// Venue/Court related models

package models

import (
	"encoding/json"
	"time"
)

// Venue represents a playing venue or court
type Venue struct {
	ID                string          `json:"id" db:"id"`
	TournamentID      string          `json:"tournament_id" db:"tournament_id"`
	Name              string          `json:"name" db:"name"`
	Type              string          `json:"type" db:"type"`
	AvailabilityRules json.RawMessage `json:"availability_rules,omitempty" db:"availability_rules"`
	IsActive          bool            `json:"is_active" db:"is_active"`
	CreatedAt         time.Time       `json:"created_at" db:"created_at"`
}

// VenueType defines different venue types
type VenueType string

const (
	VenueCourt  VenueType = "court"
	VenueField  VenueType = "field"
	VenueTable  VenueType = "table"
	VenueMat    VenueType = "mat"
	VenueCustom VenueType = "custom"
)
