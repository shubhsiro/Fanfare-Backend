package repository

import (
	"context"

	"fanfare-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the persistence actions for user accounts, reputation, and fandom following.
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	FollowUniverse(ctx context.Context, userID, universeID primitive.ObjectID) error
	GetLeaderboard(ctx context.Context, limit int) ([]models.LeaderboardEntry, error)
	GetFandomLeaderboard(ctx context.Context, universeID primitive.ObjectID, limit int) ([]models.LeaderboardEntry, error)
}

// ContentRepository manages Pop Culture Universes, Titles, Moments feed, and IssueHub Debates.
type ContentRepository interface {
	CreateUniverse(ctx context.Context, uni *models.Universe) error
	GetUniverseBySlug(ctx context.Context, slug string) (*models.Universe, error)
	ListUniverses(ctx context.Context) ([]models.Universe, error)
	CreateTitle(ctx context.Context, title *models.Title) error
	GetTitleBySlug(ctx context.Context, slug string) (*models.Title, error)
	ListTitlesByUniverse(ctx context.Context, uniID primitive.ObjectID) ([]models.Title, error)
	CreateMoment(ctx context.Context, moment *models.Moment) error
	GetMomentByID(ctx context.Context, id primitive.ObjectID) (*models.Moment, error)
	ListMoments(ctx context.Context, limit int, offset int) ([]models.Moment, error)
	CreateIssueHub(ctx context.Context, hub *models.IssueHub) error
	GetIssueHubBySlug(ctx context.Context, slug string) (*models.IssueHub, error)
	ListIssueHubs(ctx context.Context, sortOption string) ([]models.IssueHub, error)
}

// PredictionRepository handles prediction creation, placing forecasts, sentiment polls, and cohort aggregation.
type PredictionRepository interface {
	CreatePrediction(ctx context.Context, pred *models.Prediction) error
	GetPredictionByID(ctx context.Context, id primitive.ObjectID) (*models.Prediction, error)
	ListPredictionsByMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.Prediction, error)
	ListPredictionsByIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.Prediction, error)
	UpdatePrediction(ctx context.Context, pred *models.Prediction) error

	CreateParticipation(ctx context.Context, part *models.PredictionParticipation) error
	GetParticipation(ctx context.Context, predID, userID primitive.ObjectID) (*models.PredictionParticipation, error)
	ListParticipationsByPrediction(ctx context.Context, predID primitive.ObjectID) ([]models.PredictionParticipation, error)
	ListParticipationsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.PredictionParticipation, error)
	UpdateParticipation(ctx context.Context, part *models.PredictionParticipation) error

	CreateSentimentPoll(ctx context.Context, poll *models.SentimentPoll) error
	GetSentimentPollByID(ctx context.Context, id primitive.ObjectID) (*models.SentimentPoll, error)
	GetSentimentPollsForTarget(ctx context.Context, targetType string, targetID primitive.ObjectID) ([]models.SentimentPoll, error)
	CreateSentimentVote(ctx context.Context, vote *models.SentimentVote) error
	GetSentimentVote(ctx context.Context, pollID, userID primitive.ObjectID) (*models.SentimentVote, error)
	GetSentimentDistribution(ctx context.Context, pollID primitive.ObjectID) (map[string]int, error)
	GetSentimentDistributionByCohort(ctx context.Context, pollID primitive.ObjectID, cohortType string) (map[string]map[string]int, error)
}

// ContextRepository governs community context card submissions, peer ratings, and point-history logs.
type ContextRepository interface {
	CreateCard(ctx context.Context, card *models.ContextCard) error
	GetCardByID(ctx context.Context, id primitive.ObjectID) (*models.ContextCard, error)
	ListCardsForMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.ContextCard, error)
	ListCardsForIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.ContextCard, error)
	ListCardsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.ContextCard, error)
	UpdateCard(ctx context.Context, card *models.ContextCard) error

	CreateRating(ctx context.Context, rating *models.ContextCardRating) error
	GetRating(ctx context.Context, cardID, userID primitive.ObjectID) (*models.ContextCardRating, error)

	CreateReputationLog(ctx context.Context, log *models.ReputationHistory) error
	GetReputationHistory(ctx context.Context, userID primitive.ObjectID) ([]models.ReputationHistory, error)
}
