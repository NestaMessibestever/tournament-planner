// internal/repositories/container.go
// Repository container for dependency injection

package repositories

import (
	"context"
	"database/sql"
	"tournament-planner/internal/database"
)

// Container holds all repository instances
type Container struct {
	User                  *UserRepository
	Tournament            *TournamentRepository
	TournamentParticipant *TournamentParticipantRepository
	Match                 *MatchRepository
	Venue                 *VenueRepository
	Payment               *PaymentRepository
	UserPreferences       *UserPreferencesRepository
	Participant           *ParticipantRepository
	db                    *sql.DB
}

// NewContainer creates a new repository container
func NewContainer(conn *database.Connections) *Container {
	return &Container{
		User:                  NewUserRepository(conn.MySQL),
		Tournament:            NewTournamentRepository(conn.MySQL),
		TournamentParticipant: NewTournamentParticipantRepository(conn.MySQL),
		Match:                 NewMatchRepository(conn.MySQL),
		Venue:                 NewVenueRepository(conn.MySQL),
		Payment:               NewPaymentRepository(conn.MySQL),
		Participant:           NewParticipantRepository(conn.MySQL),
		UserPreferences:       NewUserPreferencesRepository(conn.MongoDB),
		db:                    conn.MySQL,
	}
}

// BeginTx starts a new database transaction
func (c *Container) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, nil)
}
