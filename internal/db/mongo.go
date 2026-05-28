package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB wraps the client and database references.
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// ConnectMongo initializes a MongoDB client and checks connectivity.
func ConnectMongo(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB client: %w", err)
	}

	// Verify the connection
	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Printf("Successfully connected to MongoDB database: %s", dbName)
	database := client.Database(dbName)

	dbObj := &MongoDB{
		Client:   client,
		Database: database,
	}

	// Ensure all collection indexes are initialized
	if err := dbObj.EnsureIndexes(); err != nil {
		log.Printf("Warning: Failed to fully ensure database indexes: %v", err)
	}

	return dbObj, nil
}

// EnsureIndexes creates necessary unique and performance indexes on startup.
func (db *MongoDB) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Users Indexes
	if err := createUniqueIndex(ctx, db.Database.Collection("users"), "username"); err != nil {
		return err
	}
	if err := createUniqueIndex(ctx, db.Database.Collection("users"), "email"); err != nil {
		return err
	}

	// 2. Universes Indexes
	if err := createUniqueIndex(ctx, db.Database.Collection("universes"), "slug"); err != nil {
		return err
	}

	// 3. Titles Indexes
	if err := createUniqueIndex(ctx, db.Database.Collection("titles"), "slug"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("titles"), "universe_id"); err != nil {
		return err
	}

	// 4. Moments Indexes (sorting feed)
	if err := createIndex(ctx, db.Database.Collection("moments"), "universe_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("moments"), "title_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("moments"), "created_at"); err != nil {
		return err
	}

	// 5. Issue Hubs Indexes
	if err := createUniqueIndex(ctx, db.Database.Collection("issue_hubs"), "slug"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("issue_hubs"), "universe_id"); err != nil {
		return err
	}

	// 6. Predictions Indexes
	if err := createIndex(ctx, db.Database.Collection("predictions"), "moment_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("predictions"), "issue_hub_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("predictions"), "status"); err != nil {
		return err
	}

	// 7. Prediction Participations Index (Strictly 1 vote per user per prediction)
	if err := createCompoundUniqueIndex(ctx, db.Database.Collection("prediction_participations"), "prediction_id", "user_id"); err != nil {
		return err
	}

	// 8. Sentiment Polls Indexes
	if err := createIndex(ctx, db.Database.Collection("sentiment_polls"), "issue_hub_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("sentiment_polls"), "title_id"); err != nil {
		return err
	}

	// 9. Sentiment Votes Index (Compound Unique + filter groups)
	if err := createCompoundUniqueIndex(ctx, db.Database.Collection("sentiment_votes"), "sentiment_poll_id", "user_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("sentiment_votes"), "region"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("sentiment_votes"), "age_band"); err != nil {
		return err
	}

	// 10. Context Cards Indexes
	if err := createIndex(ctx, db.Database.Collection("context_cards"), "moment_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("context_cards"), "issue_hub_id"); err != nil {
		return err
	}
	if err := createIndex(ctx, db.Database.Collection("context_cards"), "prediction_id"); err != nil {
		return err
	}

	// 11. Context Card Ratings Index (Strictly 1 rating per user per card)
	if err := createCompoundUniqueIndex(ctx, db.Database.Collection("context_card_ratings"), "context_card_id", "user_id"); err != nil {
		return err
	}

	// 12. Reputation History Indexes
	if err := createIndex(ctx, db.Database.Collection("reputation_history"), "user_id"); err != nil {
		return err
	}

	log.Println("MongoDB database indexes ensured and validated.")
	return nil
}

// Helpers for index creation

func createUniqueIndex(ctx context.Context, col *mongo.Collection, field string) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: field, Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := col.Indexes().CreateOne(ctx, indexModel)
	return err
}

func createIndex(ctx context.Context, col *mongo.Collection, field string) error {
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: field, Value: 1}},
	}
	_, err := col.Indexes().CreateOne(ctx, indexModel)
	return err
}

func createCompoundUniqueIndex(ctx context.Context, col *mongo.Collection, field1, field2 string) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: field1, Value: 1}, {Key: field2, Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := col.Indexes().CreateOne(ctx, indexModel)
	return err
}
