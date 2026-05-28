package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config represents the application's configuration parameters.
type Config struct {
	Port        string
	MongoURI    string
	MongoDBName string
	JWTSecret   string
}

// LoadConfig reads configuration from environment variables and an optional .env file.
func LoadConfig() *Config {
	// Attempt to load .env file; in production, environment variables are set directly.
	if err := godotenv.Load(); err != nil {
		log.Println("INFO: No .env file found or read failed. Using system environment variables.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		// Default to local MongoDB instance
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDBName := os.Getenv("MONGODB_DB_NAME")
	if mongoDBName == "" {
		mongoDBName = "fanfare"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretfanfarejwtkey2026!"
		log.Println("WARNING: JWT_SECRET is not set! Using default insecure fallback key.")
	}

	return &Config{
		Port:        port,
		MongoURI:    mongoURI,
		MongoDBName: mongoDBName,
		JWTSecret:   jwtSecret,
	}
}
