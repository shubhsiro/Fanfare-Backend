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

type MongoContentRepository struct {
	uniCol    *mongo.Collection
	titleCol  *mongo.Collection
	momentCol *mongo.Collection
	issueCol  *mongo.Collection
}

// NewMongoContentRepository creates a MongoDB collection wrapper for catalogs and feeds.
func NewMongoContentRepository(db *mongo.Database) repository.ContentRepository {
	return &MongoContentRepository{
		uniCol:    db.Collection("universes"),
		titleCol:  db.Collection("titles"),
		momentCol: db.Collection("moments"),
		issueCol:  db.Collection("issue_hubs"),
	}
}

func (r *MongoContentRepository) CreateUniverse(ctx context.Context, uni *models.Universe) error {
	uni.ID = primitive.NewObjectID()
	uni.CreatedAt = time.Now()
	uni.UpdatedAt = time.Now()
	if uni.PrimaryLanguages == nil {
		uni.PrimaryLanguages = []string{}
	}
	if uni.RegionTags == nil {
		uni.RegionTags = []string{}
	}
	if uni.Platforms == nil {
		uni.Platforms = []string{}
	}
	_, err := r.uniCol.InsertOne(ctx, uni)
	return err
}

func (r *MongoContentRepository) GetUniverseBySlug(ctx context.Context, slug string) (*models.Universe, error) {
	var uni models.Universe
	err := r.uniCol.FindOne(ctx, bson.M{"slug": slug}).Decode(&uni)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &uni, nil
}

func (r *MongoContentRepository) ListUniverses(ctx context.Context) ([]models.Universe, error) {
	cursor, err := r.uniCol.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var unis []models.Universe
	if err := cursor.All(ctx, &unis); err != nil {
		return nil, err
	}
	if unis == nil {
		unis = []models.Universe{}
	}
	return unis, nil
}

func (r *MongoContentRepository) CreateTitle(ctx context.Context, title *models.Title) error {
	title.ID = primitive.NewObjectID()
	title.CreatedAt = time.Now()
	title.UpdatedAt = time.Now()
	if title.CastList == nil {
		title.CastList = []string{}
	}
	if title.Tags == nil {
		title.Tags = []string{}
	}
	_, err := r.titleCol.InsertOne(ctx, title)
	return err
}

func (r *MongoContentRepository) GetTitleBySlug(ctx context.Context, slug string) (*models.Title, error) {
	var title models.Title
	err := r.titleCol.FindOne(ctx, bson.M{"slug": slug}).Decode(&title)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &title, nil
}

func (r *MongoContentRepository) ListTitlesByUniverse(ctx context.Context, uniID primitive.ObjectID) ([]models.Title, error) {
	cursor, err := r.titleCol.Find(ctx, bson.M{"universe_id": uniID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var titles []models.Title
	if err := cursor.All(ctx, &titles); err != nil {
		return nil, err
	}
	if titles == nil {
		titles = []models.Title{}
	}
	return titles, nil
}

func (r *MongoContentRepository) CreateMoment(ctx context.Context, moment *models.Moment) error {
	moment.ID = primitive.NewObjectID()
	moment.CreatedAt = time.Now()
	moment.UpdatedAt = time.Now()
	if moment.SourceURLs == nil {
		moment.SourceURLs = []string{}
	}
	_, err := r.momentCol.InsertOne(ctx, moment)
	return err
}

func (r *MongoContentRepository) GetMomentByID(ctx context.Context, id primitive.ObjectID) (*models.Moment, error) {
	var moment models.Moment
	err := r.momentCol.FindOne(ctx, bson.M{"_id": id}).Decode(&moment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &moment, nil
}

func (r *MongoContentRepository) ListMoments(ctx context.Context, limit int, offset int) ([]models.Moment, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.momentCol.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var moments []models.Moment
	if err := cursor.All(ctx, &moments); err != nil {
		return nil, err
	}
	if moments == nil {
		moments = []models.Moment{}
	}
	return moments, nil
}

func (r *MongoContentRepository) CreateIssueHub(ctx context.Context, hub *models.IssueHub) error {
	hub.ID = primitive.NewObjectID()
	hub.CreatedAt = time.Now()
	hub.UpdatedAt = time.Now()
	if hub.Timeline == nil {
		hub.Timeline = []models.TimelineEvent{}
	}
	if hub.EvidenceStack == nil {
		hub.EvidenceStack = []models.EvidenceItem{}
	}
	_, err := r.issueCol.InsertOne(ctx, hub)
	return err
}

func (r *MongoContentRepository) GetIssueHubBySlug(ctx context.Context, slug string) (*models.IssueHub, error) {
	var hub models.IssueHub
	err := r.issueCol.FindOne(ctx, bson.M{"slug": slug}).Decode(&hub)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &hub, nil
}

func (r *MongoContentRepository) ListIssueHubs(ctx context.Context, sortOption string) ([]models.IssueHub, error) {
	opts := options.Find()
	if sortOption == "activity" {
		opts.SetSort(bson.D{{Key: "updated_at", Value: -1}})
	} else {
		opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	}

	cursor, err := r.issueCol.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var hubs []models.IssueHub
	if err := cursor.All(ctx, &hubs); err != nil {
		return nil, err
	}
	if hubs == nil {
		hubs = []models.IssueHub{}
	}
	return hubs, nil
}
