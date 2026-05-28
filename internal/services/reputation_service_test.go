package services

import (
	"context"
	"testing"

	"fanfare-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockBadgePredictionRepository struct {
	participations []models.PredictionParticipation
}

func (m *MockBadgePredictionRepository) CreatePrediction(ctx context.Context, pred *models.Prediction) error { return nil }
func (m *MockBadgePredictionRepository) GetPredictionByID(ctx context.Context, id primitive.ObjectID) (*models.Prediction, error) { return nil, nil }
func (m *MockBadgePredictionRepository) ListPredictionsByMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.Prediction, error) { return nil, nil }
func (m *MockBadgePredictionRepository) ListPredictionsByIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.Prediction, error) { return nil, nil }
func (m *MockBadgePredictionRepository) UpdatePrediction(ctx context.Context, pred *models.Prediction) error { return nil }
func (m *MockBadgePredictionRepository) CreateParticipation(ctx context.Context, part *models.PredictionParticipation) error { return nil }
func (m *MockBadgePredictionRepository) GetParticipation(ctx context.Context, predID, userID primitive.ObjectID) (*models.PredictionParticipation, error) { return nil, nil }
func (m *MockBadgePredictionRepository) ListParticipationsByPrediction(ctx context.Context, predID primitive.ObjectID) ([]models.PredictionParticipation, error) { return nil, nil }
func (m *MockBadgePredictionRepository) UpdateParticipation(ctx context.Context, part *models.PredictionParticipation) error { return nil }
func (m *MockBadgePredictionRepository) ListParticipationsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.PredictionParticipation, error) {
	return m.participations, nil
}
func (m *MockBadgePredictionRepository) CreateSentimentPoll(ctx context.Context, poll *models.SentimentPoll) error { return nil }
func (m *MockBadgePredictionRepository) GetSentimentPollByID(ctx context.Context, id primitive.ObjectID) (*models.SentimentPoll, error) { return nil, nil }
func (m *MockBadgePredictionRepository) GetSentimentPollsForTarget(ctx context.Context, targetType string, targetID primitive.ObjectID) ([]models.SentimentPoll, error) { return nil, nil }
func (m *MockBadgePredictionRepository) CreateSentimentVote(ctx context.Context, vote *models.SentimentVote) error { return nil }
func (m *MockBadgePredictionRepository) GetSentimentVote(ctx context.Context, pollID, userID primitive.ObjectID) (*models.SentimentVote, error) { return nil, nil }
func (m *MockBadgePredictionRepository) GetSentimentDistribution(ctx context.Context, pollID primitive.ObjectID) (map[string]int, error) { return nil, nil }
func (m *MockBadgePredictionRepository) GetSentimentDistributionByCohort(ctx context.Context, pollID primitive.ObjectID, cohortType string) (map[string]map[string]int, error) { return nil, nil }

type MockBadgeContextRepository struct {
	cards   []models.ContextCard
	history []models.ReputationHistory
}

func (m *MockBadgeContextRepository) CreateCard(ctx context.Context, card *models.ContextCard) error { return nil }
func (m *MockBadgeContextRepository) GetCardByID(ctx context.Context, id primitive.ObjectID) (*models.ContextCard, error) { return nil, nil }
func (m *MockBadgeContextRepository) ListCardsForMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.ContextCard, error) { return nil, nil }
func (m *MockBadgeContextRepository) ListCardsForIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.ContextCard, error) { return nil, nil }
func (m *MockBadgeContextRepository) ListCardsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.ContextCard, error) {
	return m.cards, nil
}
func (m *MockBadgeContextRepository) UpdateCard(ctx context.Context, card *models.ContextCard) error { return nil }
func (m *MockBadgeContextRepository) CreateRating(ctx context.Context, rating *models.ContextCardRating) error { return nil }
func (m *MockBadgeContextRepository) GetRating(ctx context.Context, cardID, userID primitive.ObjectID) (*models.ContextCardRating, error) { return nil, nil }
func (m *MockBadgeContextRepository) CreateReputationLog(ctx context.Context, log *models.ReputationHistory) error { return nil }
func (m *MockBadgeContextRepository) GetReputationHistory(ctx context.Context, userID primitive.ObjectID) ([]models.ReputationHistory, error) {
	return m.history, nil
}

func TestBadgeCalculations(t *testing.T) {
	userRepo := NewMockUserRepository()
	predRepo := &MockBadgePredictionRepository{}
	ctxRepo := &MockBadgeContextRepository{}

	reputationService := NewReputationService(userRepo, predRepo, ctxRepo)

	userID := primitive.NewObjectID()
	user := &models.User{
		ID:              userID,
		Username:        "lore_analyst_leader",
		ReputationScore: 500,
	}
	_ = userRepo.Create(context.Background(), user)

	ctx := context.Background()

	// Case 1: Minimalist user -> Expect default "Fan" badge
	stats, err := reputationService.GetUserProfileStats(ctx, "lore_analyst_leader")
	if err != nil {
		t.Fatalf("Failed to compile user profile stats: %v", err)
	}
	if len(stats.Badges) != 1 || stats.Badges[0] != "Fan" {
		t.Errorf("Expected only 'Fan' badge for clean user, got: %+v", stats.Badges)
	}

	// Case 2: Make user qualify for Analyst (>= 5 resolved predictions, >= 60% accuracy)
	// Let's add 6 predictions with 4 correct (accuracy = 66.6%)
	predRepo.participations = []models.PredictionParticipation{
		{IsCorrect: true},
		{IsCorrect: true},
		{IsCorrect: true},
		{IsCorrect: true},
		{IsCorrect: false},
		{IsCorrect: false},
	}

	// Case 3: Make user qualify for Lore Master (>= 3 context cards, >= 15 net helpful rating)
	ctxRepo.cards = []models.ContextCard{
		{HelpfulCount: 10, NotHelpfulCount: 1, BiasedCount: 1}, // net = 8
		{HelpfulCount: 5, NotHelpfulCount: 0, BiasedCount: 0},  // net = 5
		{HelpfulCount: 4, NotHelpfulCount: 1, BiasedCount: 0},  // net = 3 -> total net = 16 >= 15
	}

	// Case 4: Make user qualify for Hype Leader (>= 10 sentiment poll votes)
	ctxRepo.history = make([]models.ReputationHistory, 10)
	for i := 0; i < 10; i++ {
		ctxRepo.history[i] = models.ReputationHistory{ActionType: "poll_vote"}
	}

	// Run evaluation
	stats, err = reputationService.GetUserProfileStats(ctx, "lore_analyst_leader")
	if err != nil {
		t.Fatalf("Failed to compile user stats: %v", err)
	}

	// Verify all badges are earned
	hasAnalyst := false
	hasLoreMaster := false
	hasHypeLeader := false
	for _, b := range stats.Badges {
		if b == "Analyst" {
			hasAnalyst = true
		}
		if b == "Lore Master" {
			hasLoreMaster = true
		}
		if b == "Hype Leader" {
			hasHypeLeader = true
		}
	}

	if !hasAnalyst || !hasLoreMaster || !hasHypeLeader {
		t.Errorf("Expected user to earn Analyst, Lore Master, and Hype Leader badges, got: %+v", stats.Badges)
	}
}
