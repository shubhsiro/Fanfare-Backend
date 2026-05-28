package services

import (
	"context"
	"errors"
	"fmt"

	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReputationService struct {
	userRepo repository.UserRepository
	predRepo repository.PredictionRepository
	ctxRepo  repository.ContextRepository
}

// UserProfileStats represents the detailed track record statistics for a user profile view.
type UserProfileStats struct {
	User                 models.User `json:"user"`
	PredictionsAttempted int         `json:"predictions_attempted"`
	PredictionsCorrect   int         `json:"predictions_correct"`
	PredictionAccuracy   float64     `json:"prediction_accuracy"`
	ContextCardsAdded    int         `json:"context_cards_added"`
	ContextNetHelpful    int         `json:"context_net_helpful"`
	PollsVoted           int         `json:"polls_voted"`
	Badges               []string    `json:"badges"`
}

// NewReputationService creates a service for managing user reputation score trails and badges.
func NewReputationService(
	userRepo repository.UserRepository,
	predRepo repository.PredictionRepository,
	ctxRepo repository.ContextRepository,
) *ReputationService {
	return &ReputationService{
		userRepo: userRepo,
		predRepo: predRepo,
		ctxRepo:  ctxRepo,
	}
}

// GetUserProfileStats compiles real-time profile statistics and evaluates active credentials.
func (s *ReputationService) GetUserProfileStats(ctx context.Context, username string) (*UserProfileStats, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// 1. Fetch prediction records
	participations, err := s.predRepo.ListParticipationsByUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user participations: %w", err)
	}

	predictionsAttempted := len(participations)
	predictionsCorrect := 0
	for _, part := range participations {
		if part.IsCorrect {
			predictionsCorrect++
		}
	}

	predictionAccuracy := 0.0
	if predictionsAttempted > 0 {
		predictionAccuracy = float64(predictionsCorrect) / float64(predictionsAttempted) * 100.0
	}

	// 2. Fetch context card details
	cards, err := s.ctxRepo.ListCardsByUser(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user context cards: %w", err)
	}

	contextCardsAdded := len(cards)
	contextNetHelpful := 0
	for _, card := range cards {
		netRating := card.HelpfulCount - card.NotHelpfulCount - card.BiasedCount
		contextNetHelpful += netRating
	}

	// 3. Fetch reputation log to count sentiment votes
	history, err := s.ctxRepo.GetReputationHistory(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reputation history: %w", err)
	}

	pollsVoted := 0
	for _, log := range history {
		if log.ActionType == "poll_vote" {
			pollsVoted++
		}
	}

	// 4. Badge Evaluation Logic
	var badges []string

	// Analyst Badge: >= 5 predictions and >= 60% accuracy
	if predictionsAttempted >= 5 && predictionAccuracy >= 60.0 {
		badges = append(badges, "Analyst")
	}

	// Lore Master Badge: >= 3 cards added and >= 15 net helpful rating
	if contextCardsAdded >= 3 && contextNetHelpful >= 15 {
		badges = append(badges, "Lore Master")
	}

	// Hype Leader Badge: >= 10 polls voted
	if pollsVoted >= 10 {
		badges = append(badges, "Hype Leader")
	}

	// Default fallback badge for everyone
	if len(badges) == 0 {
		badges = append(badges, "Fan")
	}

	return &UserProfileStats{
		User:                 *user,
		PredictionsAttempted: predictionsAttempted,
		PredictionsCorrect:   predictionsCorrect,
		PredictionAccuracy:   predictionAccuracy,
		ContextCardsAdded:    contextCardsAdded,
		ContextNetHelpful:    contextNetHelpful,
		PollsVoted:           pollsVoted,
		Badges:               badges,
	}, nil
}

// GetGlobalScoreboard retrieves the top ranked players in the fanbase community.
func (s *ReputationService) GetGlobalScoreboard(ctx context.Context, limit int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.GetLeaderboard(ctx, limit)
}

// GetFandomScoreboard retrieves the top ranked players for a specific universe category.
func (s *ReputationService) GetFandomScoreboard(ctx context.Context, universeID primitive.ObjectID, limit int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.userRepo.GetFandomLeaderboard(ctx, universeID, limit)
}

// FollowUniverse adds a universe to the user's followed fandoms list.
func (s *ReputationService) FollowUniverse(ctx context.Context, userID, universeID primitive.ObjectID) error {
	return s.userRepo.FollowUniverse(ctx, userID, universeID)
}
// AwardPollVotePoints logs a tiny reputation log (+1 pt) for answering quick opinion polls.
func (s *ScoringService) AwardPollVotePoints(ctx context.Context, userID primitive.ObjectID, pollID primitive.ObjectID) error {
	// Create log trail
	logEntry := &models.ReputationHistory{
		UserID:      userID,
		ActionType:  "poll_vote",
		Points:      1,
		ReferenceID: pollID,
	}

	if err := s.ctxRepo.CreateReputationLog(ctx, logEntry); err != nil {
		return err
	}

	// Increment user reputation
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil // Don't error out if user profile fetch fails
		}
		return err
	}
	user.ReputationScore += 1
	return s.userRepo.Update(ctx, user)
}
