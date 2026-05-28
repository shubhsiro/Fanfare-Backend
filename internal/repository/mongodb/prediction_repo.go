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
)

type MongoPredictionRepository struct {
	predCol      *mongo.Collection
	partCol      *mongo.Collection
	pollCol      *mongo.Collection
	voteCol      *mongo.Collection
}

// NewMongoPredictionRepository instantiates a prediction and sentiment engine repository.
func NewMongoPredictionRepository(db *mongo.Database) repository.PredictionRepository {
	return &MongoPredictionRepository{
		predCol:      db.Collection("predictions"),
		partCol:      db.Collection("prediction_participations"),
		pollCol:      db.Collection("sentiment_polls"),
		voteCol:      db.Collection("sentiment_votes"),
	}
}

func (r *MongoPredictionRepository) CreatePrediction(ctx context.Context, pred *models.Prediction) error {
	pred.ID = primitive.NewObjectID()
	pred.Status = "open"
	pred.CreatedAt = time.Now()
	pred.UpdatedAt = time.Now()
	if pred.Options == nil {
		pred.Options = []string{}
	}
	_, err := r.predCol.InsertOne(ctx, pred)
	return err
}

func (r *MongoPredictionRepository) GetPredictionByID(ctx context.Context, id primitive.ObjectID) (*models.Prediction, error) {
	var pred models.Prediction
	err := r.predCol.FindOne(ctx, bson.M{"_id": id}).Decode(&pred)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &pred, nil
}

func (r *MongoPredictionRepository) ListPredictionsByMoment(ctx context.Context, momentID primitive.ObjectID) ([]models.Prediction, error) {
	cursor, err := r.predCol.Find(ctx, bson.M{"moment_id": momentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var preds []models.Prediction
	if err := cursor.All(ctx, &preds); err != nil {
		return nil, err
	}
	if preds == nil {
		preds = []models.Prediction{}
	}
	return preds, nil
}

func (r *MongoPredictionRepository) ListPredictionsByIssueHub(ctx context.Context, hubID primitive.ObjectID) ([]models.Prediction, error) {
	cursor, err := r.predCol.Find(ctx, bson.M{"issue_hub_id": hubID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var preds []models.Prediction
	if err := cursor.All(ctx, &preds); err != nil {
		return nil, err
	}
	if preds == nil {
		preds = []models.Prediction{}
	}
	return preds, nil
}

func (r *MongoPredictionRepository) UpdatePrediction(ctx context.Context, pred *models.Prediction) error {
	pred.UpdatedAt = time.Now()
	_, err := r.predCol.UpdateOne(
		ctx,
		bson.M{"_id": pred.ID},
		bson.M{
			"$set": bson.M{
				"status":            pred.Status,
				"correct_option":    pred.CorrectOption,
				"resolution_date":   pred.ResolutionDate,
				"resolution_source": pred.ResolutionSource,
				"updated_at":        pred.UpdatedAt,
			},
		},
	)
	return err
}

func (r *MongoPredictionRepository) CreateParticipation(ctx context.Context, part *models.PredictionParticipation) error {
	part.ID = primitive.NewObjectID()
	part.PointsAwarded = 0
	part.IsCorrect = false
	part.CreatedAt = time.Now()

	_, err := r.partCol.InsertOne(ctx, part)
	if err != nil {
		// Detect duplicate key error (meaning user has already voted)
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user has already placed a prediction on this question")
		}
		return err
	}
	return nil
}

func (r *MongoPredictionRepository) GetParticipation(ctx context.Context, predID, userID primitive.ObjectID) (*models.PredictionParticipation, error) {
	var part models.PredictionParticipation
	err := r.partCol.FindOne(ctx, bson.M{"prediction_id": predID, "user_id": userID}).Decode(&part)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &part, nil
}

func (r *MongoPredictionRepository) ListParticipationsByPrediction(ctx context.Context, predID primitive.ObjectID) ([]models.PredictionParticipation, error) {
	cursor, err := r.partCol.Find(ctx, bson.M{"prediction_id": predID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var parts []models.PredictionParticipation
	if err := cursor.All(ctx, &parts); err != nil {
		return nil, err
	}
	if parts == nil {
		parts = []models.PredictionParticipation{}
	}
	return parts, nil
}

func (r *MongoPredictionRepository) ListParticipationsByUser(ctx context.Context, userID primitive.ObjectID) ([]models.PredictionParticipation, error) {
	cursor, err := r.partCol.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var parts []models.PredictionParticipation
	if err := cursor.All(ctx, &parts); err != nil {
		return nil, err
	}
	if parts == nil {
		parts = []models.PredictionParticipation{}
	}
	return parts, nil
}

func (r *MongoPredictionRepository) UpdateParticipation(ctx context.Context, part *models.PredictionParticipation) error {
	_, err := r.partCol.UpdateOne(
		ctx,
		bson.M{"_id": part.ID},
		bson.M{
			"$set": bson.M{
				"points_awarded": part.PointsAwarded,
				"is_correct":     part.IsCorrect,
			},
		},
	)
	return err
}

func (r *MongoPredictionRepository) CreateSentimentPoll(ctx context.Context, poll *models.SentimentPoll) error {
	poll.ID = primitive.NewObjectID()
	poll.CreatedAt = time.Now()
	if poll.Options == nil {
		poll.Options = []string{}
	}
	_, err := r.pollCol.InsertOne(ctx, poll)
	return err
}

func (r *MongoPredictionRepository) GetSentimentPollByID(ctx context.Context, id primitive.ObjectID) (*models.SentimentPoll, error) {
	var poll models.SentimentPoll
	err := r.pollCol.FindOne(ctx, bson.M{"_id": id}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &poll, nil
}

func (r *MongoPredictionRepository) GetSentimentPollsForTarget(ctx context.Context, targetType string, targetID primitive.ObjectID) ([]models.SentimentPoll, error) {
	filter := bson.M{}
	if targetType == "issue" {
		filter["issue_hub_id"] = targetID
	} else if targetType == "title" {
		filter["title_id"] = targetID
	} else {
		return nil, errors.New("invalid sentiment target type")
	}

	cursor, err := r.pollCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.SentimentPoll
	if err := cursor.All(ctx, &polls); err != nil {
		return nil, err
	}
	if polls == nil {
		polls = []models.SentimentPoll{}
	}
	return polls, nil
}

func (r *MongoPredictionRepository) CreateSentimentVote(ctx context.Context, vote *models.SentimentVote) error {
	vote.ID = primitive.NewObjectID()
	vote.CreatedAt = time.Now()

	_, err := r.voteCol.InsertOne(ctx, vote)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user has already voted in this sentiment poll")
		}
		return err
	}
	return nil
}

func (r *MongoPredictionRepository) GetSentimentVote(ctx context.Context, pollID, userID primitive.ObjectID) (*models.SentimentVote, error) {
	var vote models.SentimentVote
	err := r.voteCol.FindOne(ctx, bson.M{"sentiment_poll_id": pollID, "user_id": userID}).Decode(&vote)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &vote, nil
}

func (r *MongoPredictionRepository) GetSentimentDistribution(ctx context.Context, pollID primitive.ObjectID) (map[string]int, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"sentiment_poll_id": pollID}}},
		{{Key: "$group", Value: bson.M{"_id": "$selected_option", "count": bson.M{"$sum": 1}}}},
	}

	cursor, err := r.voteCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	dist := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			Option string `bson:"_id"`
			Count  int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		dist[result.Option] = result.Count
	}
	return dist, nil
}

func (r *MongoPredictionRepository) GetSentimentDistributionByCohort(ctx context.Context, pollID primitive.ObjectID, cohortType string) (map[string]map[string]int, error) {
	if cohortType != "region" && cohortType != "age_band" {
		return nil, errors.New("invalid cohort type, must be 'region' or 'age_band'")
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"sentiment_poll_id": pollID}}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"option": "$selected_option",
				"cohort": "$" + cohortType,
			},
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.voteCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Outer key: cohort name (e.g. "Asia"), Inner key: selected option (e.g. "Massively Hyped") -> Count
	dist := make(map[string]map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID struct {
				Option string `bson:"option"`
				Cohort string `bson:"cohort"`
			} `bson:"_id"`
			Count int `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		if dist[result.ID.Cohort] == nil {
			dist[result.ID.Cohort] = make(map[string]int)
		}
		dist[result.ID.Cohort][result.ID.Option] = result.Count
	}
	return dist, nil
}
