package db

import (
	"context"
	"log"
	"time"

	"fanfare-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Seeder populates empty MongoDB collections with sample data on first startup.
type Seeder struct {
	db *mongo.Database
}

// NewSeeder creates a Seeder bound to the given database.
func NewSeeder(db *mongo.Database) *Seeder {
	return &Seeder{db: db}
}

// SeedAll checks every collection and seeds it if empty.
func (s *Seeder) SeedAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.seedUniverses(ctx); err != nil {
		return err
	}
	if err := s.seedTitles(ctx); err != nil {
		return err
	}
	if err := s.seedMoments(ctx); err != nil {
		return err
	}
	if err := s.seedIssueHubs(ctx); err != nil {
		return err
	}
	if err := s.seedPredictions(ctx); err != nil {
		return err
	}
	if err := s.seedSentimentPolls(ctx); err != nil {
		return err
	}

	log.Println("Database seeding complete.")
	return nil
}

func (s *Seeder) isEmpty(ctx context.Context, colName string) bool {
	count, err := s.db.Collection(colName).CountDocuments(ctx, bson.M{})
	return err != nil || count == 0
}

func (s *Seeder) seedUniverses(ctx context.Context) error {
	if !s.isEmpty(ctx, "universes") {
		return nil
	}
	log.Println("Seeding universes...")

	universes := []interface{}{
		models.Universe{
			ID:               primitive.NewObjectID(),
			Name:             "Marvel Cinematic Universe",
			Slug:             "mcu",
			Description:      "The massive shared universe containing all Marvel Studios theatrical and Disney+ releases.",
			Type:             "mixed",
			PrimaryLanguages: []string{"English"},
			RegionTags:       []string{"Global", "North America"},
			Platforms:        []string{"Theaters", "Disney+"},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		models.Universe{
			ID:               primitive.NewObjectID(),
			Name:             "DC Universe",
			Slug:             "dcu",
			Description:      "The new DC Studios cinematic universe led by James Gunn.",
			Type:             "mixed",
			PrimaryLanguages: []string{"English"},
			RegionTags:       []string{"Global", "North America"},
			Platforms:        []string{"Theaters", "Max"},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		models.Universe{
			ID:               primitive.NewObjectID(),
			Name:             "Stranger Things",
			Slug:             "stranger-things",
			Description:      "The beloved Netflix sci-fi horror franchise set in the 1980s in Hawkins, Indiana.",
			Type:             "tv",
			PrimaryLanguages: []string{"English"},
			RegionTags:       []string{"Global"},
			Platforms:        []string{"Netflix"},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		models.Universe{
			ID:               primitive.NewObjectID(),
			Name:             "House of the Dragon",
			Slug:             "house-of-the-dragon",
			Description:      "The Game of Thrones prequel series following the Targaryen civil war.",
			Type:             "tv",
			PrimaryLanguages: []string{"English"},
			RegionTags:       []string{"Global"},
			Platforms:        []string{"Max", "HBO"},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		models.Universe{
			ID:               primitive.NewObjectID(),
			Name:             "Star Wars",
			Slug:             "star-wars",
			Description:      "The iconic space opera franchise spanning films, series, and animated projects.",
			Type:             "mixed",
			PrimaryLanguages: []string{"English"},
			RegionTags:       []string{"Global"},
			Platforms:        []string{"Theaters", "Disney+"},
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	_, err := s.db.Collection("universes").InsertMany(ctx, universes)
	return err
}

func (s *Seeder) seedTitles(ctx context.Context) error {
	if !s.isEmpty(ctx, "titles") {
		return nil
	}
	log.Println("Seeding titles...")

	// Fetch MCU universe to link titles
	var mcu models.Universe
	err := s.db.Collection("universes").FindOne(ctx, bson.M{"slug": "mcu"}).Decode(&mcu)
	if err != nil {
		return nil // Skip if no universe seeded
	}

	var st models.Universe
	_ = s.db.Collection("universes").FindOne(ctx, bson.M{"slug": "stranger-things"}).Decode(&st)

	titles := []interface{}{
		models.Title{
			ID:          primitive.NewObjectID(),
			UniverseID:  mcu.ID,
			Name:        "Avengers: Secret Wars",
			Slug:        "avengers-secret-wars",
			ReleaseDate: "2027-05-01",
			Synopsis:    "The culmination of the Multiverse Saga, bringing heroes from across realities.",
			CastList:    []string{"Robert Downey Jr.", "Chris Evans", "Tom Holland", "Simu Liu"},
			Tags:        []string{"Action", "Sci-Fi", "PG-13"},
			Status:      "in_production",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		models.Title{
			ID:          primitive.NewObjectID(),
			UniverseID:  mcu.ID,
			Name:        "Avengers: Doomsday",
			Slug:        "avengers-doomsday",
			ReleaseDate: "2026-05-01",
			Synopsis:    "Doctor Doom arrives in the MCU, threatening the Avengers like never before.",
			CastList:    []string{"Robert Downey Jr.", "Chris Hemsworth", "Scarlett Johansson"},
			Tags:        []string{"Action", "Sci-Fi", "PG-13"},
			Status:      "announced",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		models.Title{
			ID:          primitive.NewObjectID(),
			UniverseID:  st.ID,
			Name:        "Stranger Things: Season 5",
			Slug:        "stranger-things-season-5",
			ReleaseDate: "2025-11-01",
			Synopsis:    "The final chapter of the Hawkins saga. Will Eleven and her friends defeat Vecna once and for all?",
			CastList:    []string{"Millie Bobby Brown", "Finn Wolfhard", "David Harbour"},
			Tags:        []string{"Sci-Fi", "Horror", "TV-14"},
			Status:      "released",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	_, err = s.db.Collection("titles").InsertMany(ctx, titles)
	return err
}

func (s *Seeder) seedMoments(ctx context.Context) error {
	if !s.isEmpty(ctx, "moments") {
		return nil
	}
	log.Println("Seeding moments...")

	var mcu models.Universe
	_ = s.db.Collection("universes").FindOne(ctx, bson.M{"slug": "mcu"}).Decode(&mcu)

	moments := []interface{}{
		models.Moment{
			ID:         primitive.NewObjectID(),
			UniverseID: &mcu.ID,
			Title:      "Robert Downey Jr. Confirmed to Return as Doctor Doom",
			Summary:    "Marvel Studios officially confirmed RDJ will portray Victor Von Doom in the upcoming Avengers films.",
			Type:       "official_news",
			SourceURLs: []string{"https://variety.com/marvel-rdj-doom"},
			SourceType: "trade",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		models.Moment{
			ID:         primitive.NewObjectID(),
			UniverseID: &mcu.ID,
			Title:      "Leaked Set Photos Show Massive Battle Scene for Secret Wars",
			Summary:    "Paparazzi photos from the London filming location show what appears to be a multi-universe battle sequence.",
			Type:       "rumor_leak",
			SourceURLs: []string{"https://cosmicbooknews.com/secret-wars-set-leak"},
			SourceType: "gossip",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now().Add(-24 * time.Hour),
		},
		models.Moment{
			ID:        primitive.NewObjectID(),
			Title:     "Disney Acquires New IP Rights for Massive Crossover Event",
			Summary:   "Industry reports suggest Disney is negotiating cross-franchise deals that could reshape streaming content.",
			Type:      "industry_move",
			SourceURLs: []string{"https://deadline.com/disney-crossover-deal"},
			SourceType: "trade",
			CreatedAt:  time.Now().Add(-48 * time.Hour),
			UpdatedAt:  time.Now().Add(-48 * time.Hour),
		},
	}

	_, err := s.db.Collection("moments").InsertMany(ctx, moments)
	return err
}

func (s *Seeder) seedIssueHubs(ctx context.Context) error {
	if !s.isEmpty(ctx, "issue_hubs") {
		return nil
	}
	log.Println("Seeding issue hubs...")

	var mcu models.Universe
	_ = s.db.Collection("universes").FindOne(ctx, bson.M{"slug": "mcu"}).Decode(&mcu)

	hubs := []interface{}{
		models.IssueHub{
			ID:         primitive.NewObjectID(),
			UniverseID: &mcu.ID,
			Title:      "Is superhero fatigue finally fading?",
			Slug:       "superhero-fatigue-fading",
			Overview:   "A detailed analysis of box office recoveries, critical response shifts, and audience sentiment for 2025-2026 superhero releases.",
			Timeline: []models.TimelineEvent{
				{EventText: "Deadpool & Wolverine crosses $1.3B globally", Date: "2024-08-15"},
				{EventText: "Thunderbolts* receives mixed reviews", Date: "2025-05-02"},
				{EventText: "RDJ return announcement boosts MCU hype", Date: "2025-07-20"},
			},
			EvidenceStack: []models.EvidenceItem{
				{MetricName: "Deadpool 3 Global Box Office", Value: "$1.3B", Type: "box_office"},
				{MetricName: "Rotten Tomatoes Average 2024 vs 2025", Value: "+12% net approval increase", Type: "critical_rating"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	_, err := s.db.Collection("issue_hubs").InsertMany(ctx, hubs)
	return err
}

func (s *Seeder) seedPredictions(ctx context.Context) error {
	if !s.isEmpty(ctx, "predictions") {
		return nil
	}
	log.Println("Seeding predictions...")

	// Fetch a moment to link to
	var moment models.Moment
	_ = s.db.Collection("moments").FindOne(ctx, bson.M{}).Decode(&moment)

	predictions := []interface{}{
		models.Prediction{
			ID:                 primitive.NewObjectID(),
			MomentID:           &moment.ID,
			Question:           "Will Avengers: Secret Wars beat Endgame's opening weekend ($357M)?",
			Type:               "threshold",
			Options:            []string{"Yes (Above $357M)", "No (Below $357M)"},
			ResolutionCriteria: "Official Box Office Mojo numbers for domestic opening weekend.",
			Status:             "open",
			CreatorID:          primitive.NewObjectID(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
		models.Prediction{
			ID:                 primitive.NewObjectID(),
			MomentID:           &moment.ID,
			Question:           "Will RDJ's Doctor Doom receive a standalone film?",
			Type:               "binary",
			Options:            []string{"Yes", "No"},
			ResolutionCriteria: "Official Marvel Studios announcement or D23/SDCC confirmation.",
			Status:             "open",
			CreatorID:          primitive.NewObjectID(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		},
	}

	_, err := s.db.Collection("predictions").InsertMany(ctx, predictions)
	return err
}

func (s *Seeder) seedSentimentPolls(ctx context.Context) error {
	if !s.isEmpty(ctx, "sentiment_polls") {
		return nil
	}
	log.Println("Seeding sentiment polls...")

	// Fetch an issue hub to link to
	var hub models.IssueHub
	_ = s.db.Collection("issue_hubs").FindOne(ctx, bson.M{}).Decode(&hub)

	polls := []interface{}{
		models.SentimentPoll{
			ID:         primitive.NewObjectID(),
			IssueHubID: &hub.ID,
			Type:       "hype",
			Question:   "How hyped are you for RDJ's return to the MCU?",
			Options:    []string{"Massively Hyped", "Mildly Hyped", "Neutral", "Fatigued/Opposed"},
			CreatedAt:  time.Now(),
		},
		models.SentimentPoll{
			ID:         primitive.NewObjectID(),
			IssueHubID: &hub.ID,
			Type:       "satisfaction",
			Question:   "Are you satisfied with Marvel's creative direction in Phase 6?",
			Options:    []string{"Very Satisfied", "Somewhat Satisfied", "Neutral", "Disappointed"},
			CreatedAt:  time.Now(),
		},
	}

	_, err := s.db.Collection("sentiment_polls").InsertMany(ctx, polls)
	return err
}
