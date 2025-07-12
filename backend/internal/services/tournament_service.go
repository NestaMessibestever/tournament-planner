// internal/services/tournament_service.go
// Core tournament business logic including constraint-based capacity calculation

package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"tournament-planner/internal/models"
	"tournament-planner/internal/repositories"
	"tournament-planner/internal/utils"
)

// TournamentService handles all tournament-related business logic
type TournamentService struct {
	repos        *repositories.Container
	cache        *CacheService
	notification *NotificationService
	logger       *log.Logger
}

// NewTournamentService creates a new tournament service
func NewTournamentService(
	repos *repositories.Container,
	cache *CacheService,
	notification *NotificationService,
	logger *log.Logger,
) *TournamentService {
	return &TournamentService{
		repos:        repos,
		cache:        cache,
		notification: notification,
		logger:       logger,
	}
}

// CreateTournamentRequest represents the data needed to create a tournament
type CreateTournamentRequest struct {
	Name                 string                  `json:"name" binding:"required,min=3,max=255"`
	Description          string                  `json:"description" binding:"max=1000"`
	SportID              *string                 `json:"sport_id"`
	FormatType           models.TournamentFormat `json:"format_type" binding:"required"`
	FormatConfig         *models.FormatConfig    `json:"format_config"`
	StartDate            time.Time               `json:"start_date" binding:"required"`
	EndDate              time.Time               `json:"end_date" binding:"required,gtfield=StartDate"`
	Timezone             string                  `json:"timezone" binding:"required,timezone"`
	MaxMatchesPerDay     int                     `json:"max_matches_per_day" binding:"required,min=1"`
	OperationalHours     models.OperationalHours `json:"operational_hours" binding:"required"`
	AvgMatchDuration     int                     `json:"avg_match_duration" binding:"required,min=5,max=480"`
	BufferTime           int                     `json:"buffer_time" binding:"min=0,max=60"`
	RegistrationDeadline *time.Time              `json:"registration_deadline"`
	EntryFee             float64                 `json:"entry_fee" binding:"min=0"`
	AllowOnsitePayment   bool                    `json:"allow_onsite_payment"`
	CustomFields         []models.CustomField    `json:"custom_fields"`
	Venues               []CreateVenueRequest    `json:"venues" binding:"required,min=1,dive"`
}

// CreateVenueRequest represents venue creation data
type CreateVenueRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Type              string                 `json:"type" binding:"required,oneof=court field table mat custom"`
	AvailabilityRules map[string]interface{} `json:"availability_rules"`
}

// Create creates a new tournament with constraint-based capacity calculation
func (s *TournamentService) Create(ctx context.Context, organizerID string, req CreateTournamentRequest) (*models.Tournament, error) {
	// Step 1: Calculate tournament capacity based on constraints
	// This is the KEY DIFFERENTIATOR - we calculate capacity BEFORE registration
	capacity := s.calculateTournamentCapacity(req)
	s.logger.Printf("Calculated capacity for tournament: %d participants", capacity)

	// Step 2: Validate the calculated capacity
	if capacity < 2 {
		return nil, fmt.Errorf("tournament constraints too restrictive - minimum 2 participants required")
	}

	// Step 3: Create tournament entity
	tournament := &models.Tournament{
		ID:                   utils.GenerateUUID(),
		OrganizerID:          organizerID,
		Name:                 req.Name,
		Description:          req.Description,
		SportID:              req.SportID,
		FormatType:           req.FormatType,
		FormatConfig:         req.FormatConfig,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		Timezone:             req.Timezone,
		MaxMatchesPerDay:     req.MaxMatchesPerDay,
		OperationalHours:     req.OperationalHours,
		AvgMatchDuration:     req.AvgMatchDuration,
		BufferTime:           req.BufferTime,
		RegistrationDeadline: req.RegistrationDeadline,
		EntryFee:             req.EntryFee,
		AllowOnsitePayment:   req.AllowOnsitePayment,
		CapacityLimit:        capacity,
		CurrentParticipants:  0,
		Status:               models.StatusDraft,
		IsPublic:             false,
		CustomFields:         req.CustomFields,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Step 4: Begin transaction for atomicity
	tx, err := s.repos.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Step 5: Save tournament
	if err := s.repos.Tournament.CreateWithTx(tx, tournament); err != nil {
		return nil, fmt.Errorf("failed to create tournament: %w", err)
	}

	// Step 6: Create venues
	for _, venueReq := range req.Venues {
		venue := &models.Venue{
			ID:           utils.GenerateUUID(),
			TournamentID: tournament.ID,
			Name:         venueReq.Name,
			Type:         venueReq.Type,
			IsActive:     true,
			CreatedAt:    time.Now(),
		}

		if venueReq.AvailabilityRules != nil {
			venue.AvailabilityRules = utils.MustMarshalJSON(venueReq.AvailabilityRules)
		}

		if err := s.repos.Venue.CreateWithTx(tx, venue); err != nil {
			return nil, fmt.Errorf("failed to create venue: %w", err)
		}
	}

	// Step 7: Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 8: Clear any cached data
	s.cache.Delete(fmt.Sprintf("organizer_tournaments_%s", organizerID))

	// Step 9: Log analytics event
	go s.logTournamentCreated(tournament)

	return tournament, nil
}

// calculateTournamentCapacity calculates the maximum number of participants
// based on tournament format and daily match constraints.
// This is the CORE INNOVATION of the platform!
func (s *TournamentService) calculateTournamentCapacity(req CreateTournamentRequest) int {
	// Calculate total available match slots
	days := s.calculateTournamentDays(req.StartDate, req.EndDate)
	totalMatchSlots := req.MaxMatchesPerDay * days

	s.logger.Printf("Capacity calculation: %d days × %d matches/day = %d total match slots",
		days, req.MaxMatchesPerDay, totalMatchSlots)

	// Calculate operational hours per day
	dailyMinutes := s.calculateDailyOperationalMinutes(req.OperationalHours)
	matchesPerVenuePerDay := dailyMinutes / (req.AvgMatchDuration + req.BufferTime)
	totalVenueCapacity := matchesPerVenuePerDay * len(req.Venues) * days

	// Use the more restrictive constraint
	if totalVenueCapacity < totalMatchSlots {
		totalMatchSlots = totalVenueCapacity
		s.logger.Printf("Venue capacity is more restrictive: %d matches", totalVenueCapacity)
	}

	// Apply format-specific calculations
	var capacity int
	switch req.FormatType {
	case models.FormatSingleElimination:
		// Single elimination: n participants need n-1 matches
		capacity = totalMatchSlots + 1

	case models.FormatDoubleElimination:
		// Double elimination: approximately 2n-2 matches for n participants
		// So n ≈ (totalMatchSlots + 2) / 2
		capacity = (totalMatchSlots + 2) / 2

	case models.FormatRoundRobin:
		// Round robin: n(n-1)/2 matches for n participants
		// Solving quadratic equation: n² - n - 2×totalMatchSlots = 0
		// Using quadratic formula: n = (1 + √(1 + 8×totalMatchSlots)) / 2
		discriminant := 1 + 8*float64(totalMatchSlots)
		n := (1 + math.Sqrt(discriminant)) / 2
		capacity = int(n)
		// Verify we don't exceed capacity
		if capacity*(capacity-1)/2 > totalMatchSlots {
			capacity--
		}

	case models.FormatGroupToKnockout:
		// Complex calculation for group stage + knockout
		if req.FormatConfig != nil && req.FormatConfig.GroupSize > 0 && req.FormatConfig.NumberOfGroups > 0 {
			groupSize := req.FormatConfig.GroupSize
			numGroups := req.FormatConfig.NumberOfGroups

			// Group stage: each group plays round robin
			matchesPerGroup := groupSize * (groupSize - 1) / 2
			groupStageMatches := numGroups * matchesPerGroup

			// Knockout stage (assume top 2 from each group advance)
			knockoutTeams := numGroups * 2
			knockoutMatches := knockoutTeams - 1

			totalRequired := groupStageMatches + knockoutMatches
			if totalRequired <= totalMatchSlots {
				capacity = numGroups * groupSize
			} else {
				// Scale down proportionally
				scaleFactor := float64(totalMatchSlots) / float64(totalRequired)
				capacity = int(float64(numGroups*groupSize) * scaleFactor)
			}
		} else {
			// Conservative fallback
			capacity = totalMatchSlots / 3
		}

	case models.FormatSwiss:
		// Swiss system: each participant plays a fixed number of rounds
		rounds := 5 // Default Swiss rounds
		if req.FormatConfig != nil && req.FormatConfig.NumberOfRounds > 0 {
			rounds = req.FormatConfig.NumberOfRounds
		}
		// Each round has n/2 matches for n participants
		capacity = (totalMatchSlots * 2) / rounds

	default:
		// Conservative estimate for custom formats
		capacity = totalMatchSlots / 3
	}

	s.logger.Printf("Final calculated capacity: %d participants for %s format",
		capacity, req.FormatType)

	return capacity
}

// calculateTournamentDays calculates the number of days in a tournament
func (s *TournamentService) calculateTournamentDays(start, end time.Time) int {
	// Add 1 because both start and end dates are inclusive
	return int(end.Sub(start).Hours()/24) + 1
}

// calculateDailyOperationalMinutes calculates average operational minutes per day
func (s *TournamentService) calculateDailyOperationalMinutes(hours models.OperationalHours) int {
	totalMinutes := 0
	daysCount := 0

	for _, dayHours := range hours {
		startTime, _ := time.Parse("15:04", dayHours.StartTime)
		endTime, _ := time.Parse("15:04", dayHours.EndTime)
		dailyMinutes := int(endTime.Sub(startTime).Minutes())
		if dailyMinutes > 0 {
			totalMinutes += dailyMinutes
			daysCount++
		}
	}

	if daysCount == 0 {
		return 0
	}

	return totalMinutes / daysCount
}

// GetByID retrieves a tournament by ID
func (s *TournamentService) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("tournament_%s", id)
	var tournament models.Tournament
	if err := s.cache.Get(cacheKey, &tournament); err == nil {
		return &tournament, nil
	}

	// Fetch from database
	t, err := s.repos.Tournament.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	s.cache.Set(cacheKey, t, 5*time.Minute)

	return t, nil
}

// Update updates tournament information
func (s *TournamentService) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	// Get existing tournament
	tournament, err := s.repos.Tournament.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Apply updates (validation should happen at handler level)
	// This is simplified - in production you'd validate each field
	if name, ok := updates["name"].(string); ok {
		tournament.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		tournament.Description = description
	}
	// ... other fields

	tournament.UpdatedAt = time.Now()

	// Save updates
	if err := s.repos.Tournament.Update(ctx, tournament); err != nil {
		return err
	}

	// Clear cache
	s.cache.Delete(fmt.Sprintf("tournament_%s", id))

	return nil
}

// List retrieves tournaments with filters
func (s *TournamentService) List(ctx context.Context, filter repositories.ListFilter) ([]*models.Tournament, int, error) {
	return s.repos.Tournament.List(ctx, filter)
}

// Publish makes a tournament public and opens registration
func (s *TournamentService) Publish(ctx context.Context, id string) error {
	tournament, err := s.repos.Tournament.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Validate tournament is ready to publish
	if tournament.Status != models.StatusDraft {
		return fmt.Errorf("tournament must be in draft status to publish")
	}

	// Check if venues are configured
	venues, err := s.repos.Venue.GetByTournamentID(ctx, id)
	if err != nil {
		return err
	}
	if len(venues) == 0 {
		return ErrNoVenues
	}

	// Update status
	tournament.Status = models.StatusRegistrationOpen
	tournament.IsPublic = true

	if err := s.repos.Tournament.Update(ctx, tournament); err != nil {
		return err
	}

	// Clear cache
	s.cache.Delete(fmt.Sprintf("tournament_%s", id))

	// Send notifications
	go s.notification.NotifyTournamentPublished(tournament)

	return nil
}

// IsOwner checks if a user owns a tournament
func (s *TournamentService) IsOwner(ctx context.Context, tournamentID, userID string) (bool, error) {
	tournament, err := s.repos.Tournament.GetByID(ctx, tournamentID)
	if err != nil {
		return false, err
	}

	return tournament.OrganizerID == userID, nil
}

// SeedingData represents participant seeding information
type SeedingData struct {
	ParticipantID string `json:"participant_id"`
	Seed          int    `json:"seed"`
}

// GenerateFixtures generates tournament fixtures based on format and seeding
func (s *TournamentService) GenerateFixtures(ctx context.Context, tournamentID string, seedingMethod string, seedingData []SeedingData) ([]*models.Match, error) {
	// Fetch tournament with full details
	tournament, err := s.repos.Tournament.GetByIDWithDetails(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("tournament not found: %w", err)
	}

	// Validate tournament state
	if tournament.Status != models.StatusRegistrationClosed {
		return nil, fmt.Errorf("fixtures can only be generated after registration is closed")
	}

	// Fetch all registered participants
	participants, err := s.repos.TournamentParticipant.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch participants: %w", err)
	}

	if len(participants) < 2 {
		return nil, ErrInsufficientParticipants
	}

	// Apply seeding
	seededParticipants := s.applySeedingMethod(participants, seedingMethod, seedingData)

	// Generate fixtures based on format
	var fixtures []*models.Match

	switch tournament.FormatType {
	case models.FormatSingleElimination:
		fixtures = s.generateSingleEliminationFixtures(tournament, seededParticipants)

	case models.FormatDoubleElimination:
		fixtures = s.generateDoubleEliminationFixtures(tournament, seededParticipants)

	case models.FormatRoundRobin:
		fixtures = s.generateRoundRobinFixtures(tournament, seededParticipants)

	case models.FormatGroupToKnockout:
		fixtures = s.generateGroupToKnockoutFixtures(tournament, seededParticipants)

	case models.FormatSwiss:
		// Swiss system generates pairings round by round
		fixtures = s.generateSwissFirstRound(tournament, seededParticipants)

	default:
		return nil, fmt.Errorf("unsupported tournament format: %s", tournament.FormatType)
	}

	// CRITICAL VALIDATION: Ensure fixtures don't exceed capacity
	maxPossibleMatches := tournament.MaxMatchesPerDay * s.calculateTournamentDays(tournament.StartDate, tournament.EndDate)
	if len(fixtures) > maxPossibleMatches {
		return nil, fmt.Errorf("%w: %d fixtures generated but capacity only allows %d matches",
			ErrCapacityExceeded, len(fixtures), maxPossibleMatches)
	}

	// Save fixtures in transaction
	tx, err := s.repos.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, fixture := range fixtures {
		if err := s.repos.Match.CreateWithTx(tx, fixture); err != nil {
			return nil, fmt.Errorf("failed to create fixture: %w", err)
		}
	}

	// Update tournament status
	if err := s.repos.Tournament.UpdateStatusWithTx(tx, tournamentID, models.StatusInProgress); err != nil {
		return nil, fmt.Errorf("failed to update tournament status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Clear caches
	s.cache.Delete(fmt.Sprintf("tournament_%s", tournamentID))
	s.cache.Delete(fmt.Sprintf("tournament_bracket_%s", tournamentID))

	// Send notifications
	go s.notification.NotifyFixturesGenerated(tournamentID, participants)

	return fixtures, nil
}

// applySeedingMethod applies the selected seeding method to participants
func (s *TournamentService) applySeedingMethod(participants []*models.Participant, method string, data []SeedingData) []*models.Participant {
	switch method {
	case "manual":
		// Apply manual seeding from data
		seedMap := make(map[string]int)
		for _, sd := range data {
			seedMap[sd.ParticipantID] = sd.Seed
		}

		// Sort participants by seed
		sort.Slice(participants, func(i, j int) bool {
			seedI, okI := seedMap[participants[i].ID]
			seedJ, okJ := seedMap[participants[j].ID]

			if okI && okJ {
				return seedI < seedJ
			}
			if okI {
				return true
			}
			if okJ {
				return false
			}
			return participants[i].Name < participants[j].Name
		})

		// Update seed values
		for i, p := range participants {
			seed := i + 1
			p.Seed = &seed
		}

	case "random":
		// Shuffle participants randomly
		for i := len(participants) - 1; i > 0; i-- {
			j := utils.RandomInt(i + 1)
			participants[i], participants[j] = participants[j], participants[i]
		}

		// Assign sequential seeds
		for i, p := range participants {
			seed := i + 1
			p.Seed = &seed
		}

	case "skill":
		// Sort by skill rating if available
		// This would use custom registration data
		// For now, fallback to name order
		sort.Slice(participants, func(i, j int) bool {
			return participants[i].Name < participants[j].Name
		})

		for i, p := range participants {
			seed := i + 1
			p.Seed = &seed
		}

	default:
		// Default to no seeding change
	}

	return participants
}

// generateSingleEliminationFixtures creates a single elimination bracket
func (s *TournamentService) generateSingleEliminationFixtures(tournament *models.Tournament, participants []*models.Participant) []*models.Match {
	n := len(participants)
	rounds := int(math.Ceil(math.Log2(float64(n))))
	totalMatches := n - 1
	fixtures := make([]*models.Match, 0, totalMatches)

	// Calculate number of byes needed for perfect bracket
	targetSize := int(math.Pow(2, float64(rounds)))
	byes := targetSize - n

	// Create first round matches
	matchNumber := 1
	firstRoundMatches := (n - byes) / 2

	// Create bracket structure that ensures proper seeding
	// Higher seeds should face lower seeds in later rounds
	bracketPositions := s.createBracketPositions(targetSize)

	// Place participants in bracket positions
	participantPositions := make(map[int]*models.Participant)
	for i, p := range participants {
		if p.Seed != nil {
			participantPositions[*p.Seed-1] = p
		} else {
			participantPositions[i] = p
		}
	}

	// Generate matches round by round
	for round := 1; round <= rounds; round++ {
		roundMatches := targetSize / int(math.Pow(2, float64(round)))

		for i := 0; i < roundMatches; i++ {
			match := &models.Match{
				ID:           utils.GenerateUUID(),
				TournamentID: tournament.ID,
				RoundNumber:  round,
				MatchNumber:  matchNumber,
				Stage:        "main",
				Status:       models.MatchPending,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// For first round, assign participants
			if round == 1 {
				pos1 := bracketPositions[i*2]
				pos2 := bracketPositions[i*2+1]

				if p1, exists := participantPositions[pos1]; exists && pos1 < n {
					match.Participant1ID = &p1.ID
				}
				if p2, exists := participantPositions[pos2]; exists && pos2 < n {
					match.Participant2ID = &p2.ID
				}
			}

			fixtures = append(fixtures, match)
			matchNumber++
		}
	}

	// Link matches for bracket progression
	s.linkBracketProgression(fixtures, rounds)

	return fixtures
}

// createBracketPositions creates the proper bracket ordering for seeding
func (s *TournamentService) createBracketPositions(size int) []int {
	if size == 2 {
		return []int{0, 1}
	}

	// Recursively build bracket positions
	half := size / 2
	left := s.createBracketPositions(half)
	right := s.createBracketPositions(half)

	// Interleave to create proper matchups
	positions := make([]int, size)
	for i := 0; i < half; i++ {
		positions[i*2] = left[i]
		positions[i*2+1] = right[half-1-i] + half
	}

	return positions
}

// linkBracketProgression sets up the next_match_id links for bracket progression
func (s *TournamentService) linkBracketProgression(matches []*models.Match, rounds int) {
	matchesByRound := make(map[int][]*models.Match)

	// Group matches by round
	for _, match := range matches {
		matchesByRound[match.RoundNumber] = append(matchesByRound[match.RoundNumber], match)
	}

	// Link each match to its next match
	for round := 1; round < rounds; round++ {
		currentRoundMatches := matchesByRound[round]
		nextRoundMatches := matchesByRound[round+1]

		for i := 0; i < len(currentRoundMatches); i += 2 {
			nextMatchIndex := i / 2
			if nextMatchIndex < len(nextRoundMatches) {
				currentRoundMatches[i].NextMatchID = &nextRoundMatches[nextMatchIndex].ID
				if i+1 < len(currentRoundMatches) {
					currentRoundMatches[i+1].NextMatchID = &nextRoundMatches[nextMatchIndex].ID
				}
			}
		}
	}
}

// generateRoundRobinFixtures creates round robin matches
func (s *TournamentService) generateRoundRobinFixtures(tournament *models.Tournament, participants []*models.Participant) []*models.Match {
	n := len(participants)
	fixtures := make([]*models.Match, 0, n*(n-1)/2)
	matchNumber := 1

	// Generate all possible pairings
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			match := &models.Match{
				ID:             utils.GenerateUUID(),
				TournamentID:   tournament.ID,
				RoundNumber:    1, // In round robin, we'll need to optimize this later
				MatchNumber:    matchNumber,
				Stage:          "main",
				Participant1ID: &participants[i].ID,
				Participant2ID: &participants[j].ID,
				Status:         models.MatchPending,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			fixtures = append(fixtures, match)
			matchNumber++
		}
	}

	// TODO: Optimize match order to minimize back-to-back games for teams

	return fixtures
}

// Additional helper methods would continue here...
// Including generateDoubleEliminationFixtures, generateGroupToKnockoutFixtures, etc.

// logTournamentCreated logs analytics event
func (s *TournamentService) logTournamentCreated(tournament *models.Tournament) {
	// Log to analytics service
	s.analytics.LogEvent(context.Background(), "tournament_created", map[string]interface{}{
		"tournament_id": tournament.ID,
		"organizer_id":  tournament.OrganizerID,
		"format":        tournament.FormatType,
		"capacity":      tournament.CapacityLimit,
	})
}
