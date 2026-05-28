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

type MongoUserRepository struct {
	col *mongo.Collection
}

// NewMongoUserRepository instantiates a MongoDB-specific UserRepository.
func NewMongoUserRepository(db *mongo.Database) repository.UserRepository {
	return &MongoUserRepository{
		col: db.Collection("users"),
	}
}

func (r *MongoUserRepository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.ReputationScore = 0
	user.FollowedUniverses = []primitive.ObjectID{}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.col.InsertOne(ctx, user)
	return err
}

func (r *MongoUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *MongoUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *MongoUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"avatar_url":         user.AvatarURL,
				"bio":                user.Bio,
				"reputation_score":   user.ReputationScore,
				"followed_universes": user.FollowedUniverses,
				"updated_at":         user.UpdatedAt,
			},
		},
	)
	return err
}

func (r *MongoUserRepository) FollowUniverse(ctx context.Context, userID, universeID primitive.ObjectID) error {
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$addToSet": bson.M{"followed_universes": universeID}},
	)
	return err
}

func (r *MongoUserRepository) GetLeaderboard(ctx context.Context, limit int) ([]models.LeaderboardEntry, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "reputation_score", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"username": 1, "reputation_score": 1})

	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []models.LeaderboardEntry
	rank := 1
	for cursor.Next(ctx) {
		var u models.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		entries = append(entries, models.LeaderboardEntry{
			Username:        u.Username,
			ReputationScore: u.ReputationScore,
			Rank:            rank,
		})
		rank++
	}
	return entries, nil
}

func (r *MongoUserRepository) GetFandomLeaderboard(ctx context.Context, universeID primitive.ObjectID, limit int) ([]models.LeaderboardEntry, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "reputation_score", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{"username": 1, "reputation_score": 1})

	// Find users who have followed_universes containing the universeID
	cursor, err := r.col.Find(ctx, bson.M{"followed_universes": universeID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []models.LeaderboardEntry
	rank := 1
	for cursor.Next(ctx) {
		var u models.User
		if err := cursor.Decode(&u); err != nil {
			return nil, err
		}
		entries = append(entries, models.LeaderboardEntry{
			Username:        u.Username,
			ReputationScore: u.ReputationScore,
			Rank:            rank,
		})
		rank++
	}
	return entries, nil
}
