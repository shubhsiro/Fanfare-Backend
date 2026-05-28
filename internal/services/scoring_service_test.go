package services

import (
	"context"
	"testing"
	"time"

	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockPredictionRepository mocks repository.PredictionRepository for resolving tests.
type MockPredictionRepository struct {
	Predictions    map[primitive.ObjectID]*models.Prediction
	Participations map[primitive.ObjectID][]models.PredictionParticipation
	UpdatedPreds   map[primitive.ObjectID]bool
	UpdatedParts   map[primitive.ObjectID]bool
}

func NewMockPredictionRepository() *MockPredictionRepository {
	return &MockPredictionRepository{
		Predictions:    make(map[primitive.ObjectID]*models.Prediction),
		Participations: make(map[primitive.ObjectID][]models.PredictionParticipation),
		UpdatedPreds:   make(map[primitive.ObjectID]bool),
		UpdatedParts:   make(map[primitive.ObjectID]bool),
	}
}

func (m *MockPredictionRepository) CreatePrediction(ctx context.Context, pred *models.Prediction) error {
	return nil
}

func (m *MockPredictionRepository) GetPredictionByID(ctx context.Context, id primitive.ObjectID) (*models.Prediction, error) {
	pred, ok := m.Predictions[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return pred, nil
}

func (m *MockPredictionRepository) ListPredictionsByMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.Prediction, error) {
	return nil, nil
}

func (m *MockPredictionRepository) ListPredictionsByIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.Prediction, error) {
	return nil, nil
}

func (m *MockPredictionRepository) UpdatePrediction(ctx context.Context, pred *models.Prediction) error {
	m.Predictions[pred.ID] = pred
	m.UpdatedPreds[pred.ID] = true
	return nil
}

func (m *MockPredictionRepository) CreateParticipation(ctx context.Context, part *models.PredictionParticipation) error {
	return nil
}

func (m *MockPredictionRepository) GetParticipation(ctx context.Context, predID, userID primitive.ObjectID) (*models.PredictionParticipation, error) {
	return nil, nil
}

func (m *MockPredictionRepository) ListParticipationsByPrediction(ctx context.Context, predID primitive.ObjectID) ([]models.PredictionParticipation, error) {
	return m.Participations[predID], nil
}

func (m *MockPredictionRepository) ListParticipationsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.PredictionParticipation, error) {
	return nil, nil
}

func (m *MockPredictionRepository) UpdateParticipation(ctx context.Context, part *models.PredictionParticipation) error {
	m.UpdatedParts[part.ID] = true
	// Update matching participation in map
	parts := m.Participations[part.PredictionID]
	for i, p := range parts {
		if p.ID == part.ID {
			m.Participations[part.PredictionID][i] = *part
			break
		}
	}
	return nil
}

func (m *MockPredictionRepository) CreateSentimentPoll(ctx context.Context, poll *models.SentimentPoll) error {
	return nil
}

func (m *MockPredictionRepository) GetSentimentPollByID(ctx context.Context, id primitive.ObjectID) (*models.SentimentPoll, error) {
	return nil, nil
}

func (m *MockPredictionRepository) GetSentimentPollsForTarget(ctx context.Context, targetType string, targetID primitive.ObjectID) ([]models.SentimentPoll, error) {
	return nil, nil
}

func (m *MockPredictionRepository) CreateSentimentVote(ctx context.Context, vote *models.SentimentVote) error {
	return nil
}

func (m *MockPredictionRepository) GetSentimentVote(ctx context.Context, pollID, userID primitive.ObjectID) (*models.SentimentVote, error) {
	return nil, nil
}

func (m *MockPredictionRepository) GetSentimentDistribution(ctx context.Context, pollID primitive.ObjectID) (map[string]int, error) {
	return nil, nil
}

func (m *MockPredictionRepository) GetSentimentDistributionByCohort(ctx context.Context, pollID primitive.ObjectID, cohortType string) (map[string]map[string]int, error) {
	return nil, nil
}

// MockContextRepository mocks repository.ContextRepository for unit testing logs.
type MockContextRepository struct {
	Logs []models.ReputationHistory
}

func (m *MockContextRepository) CreateCard(ctx context.Context, card *models.ContextCard) error {
	return nil
}

func (m *MockContextRepository) GetCardByID(ctx context.Context, id primitive.ObjectID) (*models.ContextCard, error) {
	return nil, nil
}

func (m *MockContextRepository) ListCardsForMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.ContextCard, error) {
	return nil, nil
}

func (m *MockContextRepository) ListCardsForIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.ContextCard, error) {
	return nil, nil
}

func (m *MockContextRepository) ListCardsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.ContextCard, error) {
	return nil, nil
}

func (m *MockContextRepository) UpdateCard(ctx context.Context, card *models.ContextCard) error {
	return nil
}

func (m *MockContextRepository) CreateRating(ctx context.Context, rating *models.ContextCardRating) error {
	return nil
}

func (m *MockContextRepository) GetRating(ctx context.Context, cardID, userID primitive.ObjectID) (*models.ContextCardRating, error) {
	return nil, nil
}

func (m *MockContextRepository) CreateReputationLog(ctx context.Context, log *models.ReputationHistory) error {
	m.Logs = append(m.Logs, *log)
	return nil
}

func (m *MockContextRepository) GetReputationHistory(ctx context.Context, userID primitive.ObjectID) ([]models.ReputationHistory, error) {
	return nil, nil
}

func TestResolvePredictionMinorityAndMultiplier(t *testing.T) {
	predRepo := NewMockPredictionRepository()
	userRepo := NewMockUserRepository()
	ctxRepo := &MockContextRepository{}

	scoringService := NewScoringService(predRepo, userRepo, ctxRepo)

	// Create test prediction
	predID := primitive.NewObjectID()
	pred := &models.Prediction{
		ID:       predID,
		Question: "Will superhero fatigue fade?",
		Options:  []string{"Yes", "No"},
		Status:   "open",
	}
	predRepo.Predictions[predID] = pred

	// Create test users
	user1ID := primitive.NewObjectID()
	user2ID := primitive.NewObjectID()
	user3ID := primitive.NewObjectID()
	user4ID := primitive.NewObjectID()

	userRepo.Create(context.Background(), &models.User{ID: user1ID, Username: "user1"})
	userRepo.Create(context.Background(), &models.User{ID: user2ID, Username: "user2"})
	userRepo.Create(context.Background(), &models.User{ID: user3ID, Username: "user3"})
	userRepo.Create(context.Background(), &models.User{ID: user4ID, Username: "user4"})

	// Setup voting distributions:
	// - User 1: "Yes", high confidence (Core Winner)
	// - User 2: "No", medium confidence (Loser)
	// - User 3: "No", high confidence (Loser)
	// - User 4: "No", low confidence (Loser)
	// Output: Correct option is "Yes". Only 1 out of 4 correct (Minority ratio = 25% < 50%).
	participations := []models.PredictionParticipation{
		{
			ID:             primitive.NewObjectID(),
			PredictionID:   predID,
			UserID:         user1ID,
			SelectedOption: "Yes",
			Confidence:     "high",
		},
		{
			ID:             primitive.NewObjectID(),
			PredictionID:   predID,
			UserID:         user2ID,
			SelectedOption: "No",
			Confidence:     "medium",
		},
		{
			ID:             primitive.NewObjectID(),
			PredictionID:   predID,
			UserID:         user3ID,
			SelectedOption: "No",
			Confidence:     "high",
		},
		{
			ID:             primitive.NewObjectID(),
			PredictionID:   predID,
			UserID:         user4ID,
			SelectedOption: "No",
			Confidence:     "low",
		},
	}
	predRepo.Participations[predID] = participations

	ctx := context.Background()
	err := scoringService.ResolvePrediction(ctx, predID, "Yes", "Official box office data")
	if err != nil {
		t.Fatalf("Expected prediction to resolve smoothly, got: %v", err)
	}

	// Verify prediction status
	updatedPred, _ := predRepo.GetPredictionByID(ctx, predID)
	if updatedPred.Status != "resolved" || *updatedPred.CorrectOption != "Yes" {
		t.Error("Expected prediction status to be 'resolved' with correct option 'Yes'")
	}

	// Verify points awarded
	// User 1 expected points:
	// - Base: 100
	// - High confidence multiplier: 1.5x -> 150 points
	// - Minority ratio: 1/4 = 0.25 (winner is 25% of crowd, less than 50% threshold)
	// - Minority Winner Bonus: base points * (1 - ratio) -> 100 * 0.75 -> 75 points
	// - Total expected points = 150 + 75 = 225 points!
	usr1, _ := userRepo.GetByID(ctx, user1ID)
	if usr1.ReputationScore != 225 {
		t.Errorf("Expected User 1 to be awarded 225 points, got: %d", usr1.ReputationScore)
	}

	// Verify Losers are awarded 0 points
	usr2, _ := userRepo.GetByID(ctx, user2ID)
	if usr2.ReputationScore != 0 {
		t.Errorf("Expected User 2 (loser) to have 0 points, got: %d", usr2.ReputationScore)
	}

	// Verify reputation log created
	if len(ctxRepo.Logs) != 1 {
		t.Errorf("Expected exactly 1 reputation log generated for the winner, got: %d", len(ctxRepo.Logs))
	} else {
		log := ctxRepo.Logs[0]
		if log.UserID != user1ID || log.Points != 225 || log.ActionType != "prediction_correct" {
			t.Errorf("Reputation log contents invalid: %+v", log)
		}
	}
}

func TestResolveAlreadyResolvedPrediction(t *testing.T) {
	predRepo := NewMockPredictionRepository()
	userRepo := NewMockUserRepository()
	ctxRepo := &MockContextRepository{}
	scoringService := NewScoringService(predRepo, userRepo, ctxRepo)

	predID := primitive.NewObjectID()
	resDate := time.Now()
	correctOpt := "Yes"
	pred := &models.Prediction{
		ID:             predID,
		Question:       "Resolved already?",
		Options:        []string{"Yes", "No"},
		Status:         "resolved",
		CorrectOption:  &correctOpt,
		ResolutionDate: &resDate,
	}
	predRepo.Predictions[predID] = pred

	err := scoringService.ResolvePrediction(context.Background(), predID, "Yes", "Test source")
	if err == nil || err.Error() != "prediction is already resolved or cancelled" {
		t.Errorf("Expected already resolved error, got %v", err)
	}
}
