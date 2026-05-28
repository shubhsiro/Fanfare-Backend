package mongodb

import (
	"context"
	"errors"
	"time"

	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoContextRepository struct {
	cardCol   *mongo.Collection
	ratingCol *mongo.Collection
	repCol    *mongo.Collection
}

// NewMongoContextRepository instantiates a MongoDB context-governance repository.
func NewMongoContextRepository(db *mongo.Database) repository.ContextRepository {
	return &MongoContextRepository{
		cardCol:   db.Collection("context_cards"),
		ratingCol: db.Collection("context_card_ratings"),
		repCol:    db.Collection("reputation_history"),
	}
}

func (r *MongoContextRepository) CreateCard(ctx context.Context, card *models.ContextCard) error {
	card.ID = primitive.NewObjectID()
	card.Status = "approved" // Default auto-approve for the early MVP
	card.HelpfulCount = 0
	card.NotHelpfulCount = 0
	card.BiasedCount = 0
	card.CreatedAt = time.Now()
	card.UpdatedAt = time.Now()

	_, err := r.cardCol.InsertOne(ctx, card)
	return err
}

func (r *MongoContextRepository) GetCardByID(ctx context.Context, id primitive.ObjectID) (*models.ContextCard, error) {
	var card models.ContextCard
	err := r.cardCol.FindOne(ctx, bson.M{"_id": id}).Decode(&card)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &card, nil
}

func (r *MongoContextRepository) ListCardsForMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.ContextCard, error) {
	cursor, err := r.cardCol.Find(ctx, bson.M{"moment_id": momentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []models.ContextCard
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []models.ContextCard{}
	}
	return cards, nil
}

func (r *MongoContextRepository) ListCardsForIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.ContextCard, error) {
	cursor, err := r.cardCol.Find(ctx, bson.M{"issue_hub_id": hubID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []models.ContextCard
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []models.ContextCard{}
	}
	return cards, nil
}

func (r *MongoContextRepository) ListCardsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.ContextCard, error) {
	cursor, err := r.cardCol.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []models.ContextCard
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	if cards == nil {
		cards = []models.ContextCard{}
	}
	return cards, nil
}

func (r *MongoContextRepository) UpdateCard(ctx context.Context, card *models.ContextCard) error {
	card.UpdatedAt = time.Now()
	_, err := r.cardCol.UpdateOne(
		ctx,
		bson.M{"_id": card.ID},
		bson.M{
			"$set": bson.M{
				"status":            card.Status,
				"helpful_count":     card.HelpfulCount,
				"not_helpful_count": card.NotHelpfulCount,
				"biased_count":      card.BiasedCount,
				"updated_at":        card.UpdatedAt,
			},
		},
	)
	return err
}

func (r *MongoContextRepository) CreateRating(ctx context.Context, rating *models.ContextCardRating) error {
	rating.ID = primitive.NewObjectID()
	rating.CreatedAt = time.Now()

	_, err := r.ratingCol.InsertOne(ctx, rating)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user has already rated this context card")
		}
		return err
	}
	return nil
}

func (r *MongoContextRepository) GetRating(ctx context.Context, cardID, userID primitive.ObjectID) (*models.ContextCardRating, error) {
	var rating models.ContextCardRating
	err := r.ratingCol.FindOne(ctx, bson.M{"context_card_id": cardID, "user_id": userID}).Decode(&rating)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &rating, nil
}

func (r *MongoContextRepository) CreateReputationLog(ctx context.Context, log *models.ReputationHistory) error {
	log.ID = primitive.NewObjectID()
	log.CreatedAt = time.Now()
	_, err := r.repCol.InsertOne(ctx, log)
	return err
}

func (r *MongoContextRepository) GetReputationHistory(ctx context.Context, userID primitive.ObjectID) ([]models.ReputationHistory, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.repCol.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []models.ReputationHistory
	if err := cursor.All(ctx, &history); err != nil {
		return nil, err
	}
	if history == nil {
		history = []models.ReputationHistory{}
	}
	return history, nil
}
