// internal/services/match_service.go
// Match management and progression logic

package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"tournament-planner/internal/models"
	"tournament-planner/internal/repositories"
)

// MatchService handles match-related business logic
type MatchService struct {
	repos        *repositories.Container
	cache        *CacheService
	notification *NotificationService
	logger       *log.Logger
}

// NewMatchService creates a new match service
func NewMatchService(
	repos *repositories.Container,
	cache *CacheService,
	notification *NotificationService,
	logger *log.Logger,
) *MatchService {
	return &MatchService{
		repos:        repos,
		cache:        cache,
		notification: notification,
		logger:       logger,
	}
}

// GetByID retrieves a match by ID
func (s *MatchService) GetByID(ctx context.Context, id string) (*models.Match, error) {
	return s.repos.Match.GetByID(ctx, id)
}

// GetByTournamentID retrieves all matches for a tournament
func (s *MatchService) GetByTournamentID(ctx context.Context, tournamentID string) ([]*models.Match, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("tournament_matches_%s", tournamentID)
	var matches []*models.Match
	if err := s.cache.Get(cacheKey, &matches); err == nil {
		return matches, nil
	}

	// Fetch from database
	matches, err := s.repos.Match.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, err
	}

	// Cache for 1 minute (short because matches update frequently)
	s.cache.Set(cacheKey, matches, 1*time.Minute)

	return matches, nil
}

// UpdateSchedule updates match schedule information
func (s *MatchService) UpdateSchedule(ctx context.Context, matchID string, scheduledTime time.Time, venueID string) error {
	// Get match
	match, err := s.repos.Match.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	// Validate venue exists
	if venueID != "" {
		venue, err := s.repos.Venue.GetByID(ctx, venueID)
		if err != nil {
			return fmt.Errorf("venue not found: %w", err)
		}
		if venue.TournamentID != match.TournamentID {
			return fmt.Errorf("venue does not belong to this tournament")
		}
	}

	// Update match
	match.ScheduledDatetime = &scheduledTime
	match.VenueID = &venueID
	match.Status = models.MatchScheduled

	if err := s.repos.Match.Update(ctx, match); err != nil {
		return err
	}

	// Clear cache
	s.cache.Delete(fmt.Sprintf("tournament_matches_%s", match.TournamentID))

	// Send notifications
	if match.Participant1ID != nil && match.Participant2ID != nil {
		go s.notification.NotifyMatchScheduled(match, []string{*match.Participant1ID, *match.Participant2ID})
	}

	return nil
}

// ReportScore reports match score and handles bracket progression
func (s *MatchService) ReportScore(ctx context.Context, matchID string, score1, score2 int, scoreDetails *models.ScoreDetails) error {
	// Get match
	match, err := s.repos.Match.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	// Validate match can have score reported
	if match.Status != models.MatchScheduled && match.Status != models.MatchInProgress {
		return fmt.Errorf("match is not in a state where score can be reported")
	}

	// Determine winner
	var winnerID string
	if score1 > score2 && match.Participant1ID != nil {
		winnerID = *match.Participant1ID
	} else if score2 > score1 && match.Participant2ID != nil {
		winnerID = *match.Participant2ID
	} else {
		return fmt.Errorf("tie score not allowed - must have a winner")
	}

	// Begin transaction
	tx, err := s.repos.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update match score
	if err := s.repos.Match.UpdateScore(ctx, matchID, score1, score2, winnerID, scoreDetails); err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	// Handle bracket progression
	if match.NextMatchID != nil {
		nextMatch, err := s.repos.Match.GetByID(ctx, *match.NextMatchID)
		if err != nil {
			return fmt.Errorf("failed to get next match: %w", err)
		}

		// Determine which slot to fill in next match
		if nextMatch.Participant1ID == nil {
			nextMatch.Participant1ID = &winnerID
		} else if nextMatch.Participant2ID == nil {
			nextMatch.Participant2ID = &winnerID
		} else {
			return fmt.Errorf("next match already has both participants")
		}

		// Update next match
		if err := s.repos.Match.Update(ctx, nextMatch); err != nil {
			return fmt.Errorf("failed to update next match: %w", err)
		}

		// If next match now has both participants, notify them
		if nextMatch.Participant1ID != nil && nextMatch.Participant2ID != nil {
			go s.notification.NotifyMatchScheduled(nextMatch, []string{*nextMatch.Participant1ID, *nextMatch.Participant2ID})
		}
	}

	// Update participant statistics
	if match.Participant1ID != nil {
		matchesWon := 0
		if winnerID == *match.Participant1ID {
			matchesWon = 1
		}
		s.repos.Participant.UpdateStats(ctx, *match.Participant1ID, 1, matchesWon)
	}

	if match.Participant2ID != nil {
		matchesWon := 0
		if winnerID == *match.Participant2ID {
			matchesWon = 1
		}
		s.repos.Participant.UpdateStats(ctx, *match.Participant2ID, 1, matchesWon)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// Clear caches
	s.cache.Delete(fmt.Sprintf("tournament_matches_%s", match.TournamentID))
	s.cache.Delete(fmt.Sprintf("tournament_bracket_%s", match.TournamentID))

	// Send result notifications
	if match.Participant1ID != nil && match.Participant2ID != nil {
		go s.notification.NotifyMatchResult(match, []string{*match.Participant1ID, *match.Participant2ID})
	}

	return nil
}

// StartMatch marks a match as in progress
func (s *MatchService) StartMatch(ctx context.Context, matchID string) error {
	return s.repos.Match.UpdateStatus(ctx, matchID, models.MatchInProgress)
}

// CancelMatch cancels a match
func (s *MatchService) CancelMatch(ctx context.Context, matchID string, reason string) error {
	match, err := s.repos.Match.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if match.Status == models.MatchCompleted {
		return fmt.Errorf("cannot cancel a completed match")
	}

	match.Status = models.MatchCancelled
	match.Notes = &reason

	return s.repos.Match.Update(ctx, match)
}

// HasAccess checks if a user can access/modify a match
func (s *MatchService) HasAccess(ctx context.Context, matchID, userID string) (bool, error) {
	match, err := s.repos.Match.GetByID(ctx, matchID)
	if err != nil {
		return false, err
	}

	// Check if user is tournament organizer
	tournament, err := s.repos.Tournament.GetByID(ctx, match.TournamentID)
	if err != nil {
		return false, err
	}

	if tournament.OrganizerID == userID {
		return true, nil
	}

	// Check if user is a participant in this match
	participant, err := s.repos.Participant.GetByUserID(ctx, userID)
	if err != nil {
		return false, nil // User is not a participant
	}

	if match.Participant1ID != nil && *match.Participant1ID == participant.ID {
		return true, nil
	}
	if match.Participant2ID != nil && *match.Participant2ID == participant.ID {
		return true, nil
	}

	// TODO: Check if user is assigned referee

	return false, nil
}

// GetScheduleByVenueAndDate retrieves matches for a specific venue and date
func (s *MatchService) GetScheduleByVenueAndDate(ctx context.Context, venueID string, date time.Time) ([]*models.Match, error) {
	return s.repos.Match.ListByVenueAndDate(ctx, venueID, date)
}
