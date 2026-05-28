package main

import (
	"log"

	"fanfare-backend/internal/config"
	"fanfare-backend/internal/db"
	"fanfare-backend/internal/handlers"
	"fanfare-backend/internal/middleware"
	mongoRepo "fanfare-backend/internal/repository/mongodb"
	"fanfare-backend/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// ──────────────────────────────────────────────────────────────
	// 1. Load Configuration
	// ──────────────────────────────────────────────────────────────
	cfg := config.LoadConfig()
	log.Printf("Starting FanFare Backend on port %s ...", cfg.Port)

	// ──────────────────────────────────────────────────────────────
	// 2. Connect to MongoDB
	// ──────────────────────────────────────────────────────────────
	mongoDB, err := db.ConnectMongo(cfg.MongoURI, cfg.MongoDBName)
	if err != nil {
		log.Fatalf("FATAL: Could not connect to MongoDB: %v", err)
	}
	log.Println("MongoDB connection established and indexes ensured.")

	// ──────────────────────────────────────────────────────────────
	// 3. Initialize Repository Layer
	// ──────────────────────────────────────────────────────────────
	userRepo := mongoRepo.NewMongoUserRepository(mongoDB.Database)
	contentRepo := mongoRepo.NewMongoContentRepository(mongoDB.Database)
	predRepo := mongoRepo.NewMongoPredictionRepository(mongoDB.Database)
	ctxRepo := mongoRepo.NewMongoContextRepository(mongoDB.Database)

	// ──────────────────────────────────────────────────────────────
	// 4. Initialize Service Layer
	// ──────────────────────────────────────────────────────────────
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	scoringService := services.NewScoringService(predRepo, userRepo, ctxRepo)
	repService := services.NewReputationService(userRepo, predRepo, ctxRepo)

	// ──────────────────────────────────────────────────────────────
	// 5. Seed Database (populates empty collections with sample data)
	// ──────────────────────────────────────────────────────────────
	seeder := db.NewSeeder(mongoDB.Database)
	if err := seeder.SeedAll(); err != nil {
		log.Printf("WARNING: Seeder encountered errors: %v", err)
	}

	// ──────────────────────────────────────────────────────────────
	// 6. Initialize HTTP Handlers
	// ──────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(authService, repService, ctxRepo)
	contentHandler := handlers.NewContentHandler(contentRepo, predRepo)
	predictionHandler := handlers.NewPredictionHandler(predRepo, scoringService)
	sentimentHandler := handlers.NewSentimentHandler(predRepo, scoringService)
	contextHandler := handlers.NewContextHandler(ctxRepo, userRepo)

	// ──────────────────────────────────────────────────────────────
	// 7. Configure Gin Router
	// ──────────────────────────────────────────────────────────────
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// Health check (always public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "fanfare-backend"})
	})

	// ──────────────────────────────────────────────────────────────
	// 8. Register API Routes
	// ──────────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")
	{
		// ── PUBLIC ROUTES (no auth required) ──────────────────────
		// Only registration and login are accessible without a token.
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// ── PROTECTED ROUTES (JWT auth required) ─────────────────
		// All remaining endpoints require a valid Bearer token.
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Users & Profiles
			users := protected.Group("/users")
			{
				users.GET("/me", authHandler.GetMe)
				users.GET("/me/reputation", authHandler.GetReputationHistory)
				users.GET("/:username", authHandler.GetUserProfile)
				users.POST("/follow-fandom", authHandler.FollowFandom)
			}

			// Scoreboard / Leaderboard
			protected.GET("/scoreboard", authHandler.GetScoreboard)

			// Universes & Titles (content catalog)
			protected.GET("/universes", contentHandler.ListUniverses)
			protected.GET("/universes/:slug", contentHandler.GetUniverse)
			protected.GET("/titles/:slug", contentHandler.GetTitle)

			// Moments / News Feed
			protected.GET("/moments", contentHandler.ListMoments)
			protected.GET("/moments/:id", contentHandler.GetMoment)
			protected.GET("/moments/:id/predictions", contentHandler.ListMomentPredictions)

			// Issue Hubs / Debates
			protected.GET("/issues", contentHandler.ListIssueHubs)
			protected.GET("/issues/:slug", contentHandler.GetIssueHub)
			protected.GET("/issues/:slug/predictions", contentHandler.ListIssueHubPredictions)

			// Predictions
			predictions := protected.Group("/predictions")
			{
				predictions.POST("/:id/vote", predictionHandler.VotePrediction)
				predictions.POST("/:id/resolve", predictionHandler.ResolvePrediction)
			}

			// Sentiment Polls & Dashboards
			sentiment := protected.Group("/sentiment")
			{
				sentiment.GET("/dashboards/:target_type/:target_id", sentimentHandler.GetSentimentDashboard)
				sentiment.POST("/polls/:id/vote", sentimentHandler.VoteSentimentPoll)
			}

			// Context Cards
			context := protected.Group("/context")
			{
				context.POST("", contextHandler.CreateContextCard)
				context.POST("/:id/rate", contextHandler.RateContextCard)
			}
		}
	}

	// ──────────────────────────────────────────────────────────────
	// 9. Start Server
	// ──────────────────────────────────────────────────────────────
	log.Printf("FanFare API server running at http://localhost:%s", cfg.Port)
	log.Printf("Swagger docs available at: docs/swagger.yaml")
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("FATAL: Failed to start server: %v", err)
	}
}
