package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user account and their profile metrics.
type User struct {
	ID                primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Username          string               `bson:"username" json:"username"`
	Email             string               `bson:"email" json:"email"`
	PasswordHash      string               `bson:"password_hash" json:"-"`
	AvatarURL         string               `bson:"avatar_url,omitempty" json:"avatar_url"`
	Bio               string               `bson:"bio,omitempty" json:"bio"`
	ReputationScore   int                  `bson:"reputation_score" json:"reputation_score"`
	FollowedUniverses []primitive.ObjectID `bson:"followed_universes" json:"followed_universes"`
	CreatedAt         time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time            `bson:"updated_at" json:"updated_at"`
}

// RegisterRequest represents the user registration payload.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the user login payload.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response containing the user profile and JWT.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Universe represents a pop-culture franchise (MCU, DC, Stranger Things).
type Universe struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name             string             `bson:"name" json:"name"`
	Slug             string             `bson:"slug" json:"slug"`
	Description      string             `bson:"description" json:"description"`
	Type             string             `bson:"type" json:"type"` // "movie", "tv", "mixed"
	PrimaryLanguages []string           `bson:"primary_languages" json:"primary_languages"`
	RegionTags       []string           `bson:"region_tags" json:"region_tags"`
	Platforms        []string           `bson:"platforms" json:"platforms"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// Title represents a nested watchable unit inside a Universe.
type Title struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UniverseID   primitive.ObjectID `bson:"universe_id" json:"universe_id"`
	Name         string             `bson:"name" json:"name"`
	Slug         string             `bson:"slug" json:"slug"`
	ReleaseDate  string             `bson:"release_date" json:"release_date"`
	Synopsis     string             `bson:"synopsis" json:"synopsis"`
	PosterArtURL string             `bson:"poster_art_url,omitempty" json:"poster_art_url"`
	CastList     []string           `bson:"cast_list" json:"cast_list"`
	Tags         []string           `bson:"tags" json:"tags"` // e.g. genre, PG rating
	Status       string             `bson:"status" json:"status"` // "announced", "in_production", "released", "ended"
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}

// Moment represents a news card (the atom of the app feed).
type Moment struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UniverseID *primitive.ObjectID `bson:"universe_id,omitempty" json:"universe_id,omitempty"`
	TitleID    *primitive.ObjectID `bson:"title_id,omitempty" json:"title_id,omitempty"`
	Title      string              `bson:"title" json:"title"`
	Summary    string              `bson:"summary" json:"summary"`
	Type       string              `bson:"type" json:"type"` // "official_news", "industry_move", "rumor_leak", "fan_theory"
	SourceURLs []string            `bson:"source_urls" json:"source_urls"`
	SourceType string              `bson:"source_type" json:"source_type"` // "studio", "trade", "gossip", "user"
	CreatedAt  time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time           `bson:"updated_at" json:"updated_at"`
}

// TimelineEvent represents a nested timeline node in an IssueHub.
type TimelineEvent struct {
	MomentID  *primitive.ObjectID `bson:"moment_id,omitempty" json:"moment_id,omitempty"`
	EventText string              `bson:"event_text" json:"event_text"`
	Date      string              `bson:"date" json:"date"`
}

// EvidenceItem represents a nested data metric in an IssueHub.
type EvidenceItem struct {
	MetricName string `bson:"metric_name" json:"metric_name"`
	Value      string `bson:"value" json:"value"`
	Type       string `bson:"type" json:"type"` // "box_office", "critical_rating", "statement"
}

// IssueHub represents a long-running community debate hub.
type IssueHub struct {
	ID            primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UniverseID    *primitive.ObjectID `bson:"universe_id,omitempty" json:"universe_id,omitempty"`
	TitleID       *primitive.ObjectID `bson:"title_id,omitempty" json:"title_id,omitempty"`
	Title         string              `bson:"title" json:"title"`
	Slug          string              `bson:"slug" json:"slug"`
	Overview      string              `bson:"overview" json:"overview"`
	Timeline      []TimelineEvent     `bson:"timeline" json:"timeline"`
	EvidenceStack []EvidenceItem      `bson:"evidence_stack" json:"evidence_stack"`
	CreatedAt     time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time           `bson:"updated_at" json:"updated_at"`
}

// Prediction represents a structured question open to users.
type Prediction struct {
	ID                 primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	MomentID           *primitive.ObjectID `bson:"moment_id,omitempty" json:"moment_id,omitempty"`
	IssueHubID         *primitive.ObjectID `bson:"issue_hub_id,omitempty" json:"issue_hub_id,omitempty"`
	Question           string              `bson:"question" json:"question"`
	Type               string              `bson:"type" json:"type"` // "binary", "threshold"
	Options            []string            `bson:"options" json:"options"` // e.g. ["Yes", "No"]
	ResolutionCriteria string              `bson:"resolution_criteria" json:"resolution_criteria"`
	Status             string              `bson:"status" json:"status"` // "open", "resolved", "cancelled"
	CorrectOption      *string             `bson:"correct_option,omitempty" json:"correct_option,omitempty"`
	ResolutionDate     *time.Time          `bson:"resolution_date,omitempty" json:"resolution_date,omitempty"`
	ResolutionSource   *string             `bson:"resolution_source,omitempty" json:"resolution_source,omitempty"`
	CreatorID          primitive.ObjectID  `bson:"creator_id" json:"creator_id"`
	CreatedAt          time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time           `bson:"updated_at" json:"updated_at"`
}

// PredictionParticipation represents a user's forecast and confidence level.
type PredictionParticipation struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PredictionID   primitive.ObjectID `bson:"prediction_id" json:"prediction_id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	SelectedOption string             `bson:"selected_option" json:"selected_option"`
	Confidence     string             `bson:"confidence" json:"confidence"` // "low", "medium", "high"
	PointsAwarded  int                `bson:"points_awarded" json:"points_awarded"`
	IsCorrect      bool               `bson:"is_correct" json:"is_correct"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// PredictionVoteRequest represents the user input to place a prediction.
type PredictionVoteRequest struct {
	SelectedOption string `json:"selected_option" binding:"required"`
	Confidence     string `json:"confidence" binding:"required,oneof=low medium high"`
}

// PredictionResolveRequest represents the administrator's request to resolve a prediction.
type PredictionResolveRequest struct {
	CorrectOption string `json:"correct_option" binding:"required"`
	Source        string `json:"source" binding:"required"`
}

// SentimentPoll represents a quick Fandom Pulse poll.
type SentimentPoll struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	IssueHubID *primitive.ObjectID `bson:"issue_hub_id,omitempty" json:"issue_hub_id,omitempty"`
	TitleID    *primitive.ObjectID `bson:"title_id,omitempty" json:"title_id,omitempty"`
	Type       string              `bson:"type" json:"type"` // "hype", "support", "satisfaction"
	Question   string              `bson:"question" json:"question"`
	Options    []string            `bson:"options" json:"options"`
	CreatedAt  time.Time           `bson:"created_at" json:"created_at"`
}

// SentimentVote represents a vote placed in a sentiment poll, capturing demographic cohort fields.
type SentimentVote struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SentimentPollID primitive.ObjectID `bson:"sentiment_poll_id" json:"sentiment_poll_id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	SelectedOption  string             `bson:"selected_option" json:"selected_option"`
	Region          string             `bson:"region" json:"region"`
	AgeBand         string             `bson:"age_band" json:"age_band"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

// SentimentVoteRequest represents the payload to vote in a sentiment poll.
type SentimentVoteRequest struct {
	SelectedOption string `json:"selected_option" binding:"required"`
	Region         string `json:"region" binding:"required"`
	AgeBand        string `json:"age_band" binding:"required"`
}

// ContextCard represents a community context fact note.
type ContextCard struct {
	ID               primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	MomentID         *primitive.ObjectID `bson:"moment_id,omitempty" json:"moment_id,omitempty"`
	IssueHubID       *primitive.ObjectID `bson:"issue_hub_id,omitempty" json:"issue_hub_id,omitempty"`
	PredictionID     *primitive.ObjectID `bson:"prediction_id,omitempty" json:"prediction_id,omitempty"`
	UserID           primitive.ObjectID  `bson:"user_id" json:"user_id"`
	WhatsAccurate    string              `bson:"whats_accurate" json:"whats_accurate"`
	WhatsMissing     string              `bson:"whats_missing" json:"whats_missing"`
	WhatsSpeculative string              `bson:"whats_speculative" json:"whats_speculative"`
	Status           string              `bson:"status" json:"status"` // "pending", "approved", "rejected"
	HelpfulCount     int                 `bson:"helpful_count" json:"helpful_count"`
	NotHelpfulCount  int                 `bson:"not_helpful_count" json:"not_helpful_count"`
	BiasedCount      int                 `bson:"biased_count" json:"biased_count"`
	CreatedAt        time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time           `bson:"updated_at" json:"updated_at"`
}

// ContextCardRequest represents the request body to write a context card.
type ContextCardRequest struct {
	MomentID         string `json:"moment_id,omitempty"`
	IssueHubID       string `json:"issue_hub_id,omitempty"`
	PredictionID     string `json:"prediction_id,omitempty"`
	WhatsAccurate    string `json:"whats_accurate" binding:"required,min=10"`
	WhatsMissing     string `json:"whats_missing" binding:"required,min=10"`
	WhatsSpeculative string `json:"whats_speculative" binding:"required,min=10"`
}

// ContextCardRating represents a peer review on neutrality and clarity.
type ContextCardRating struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ContextCardID primitive.ObjectID `bson:"context_card_id" json:"context_card_id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Rating        string             `bson:"rating" json:"rating"` // "helpful", "not_helpful", "biased"
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
}

// ContextCardRateRequest represents the payload to review a card.
type ContextCardRateRequest struct {
	Rating string `json:"rating" binding:"required,oneof=helpful not_helpful biased"`
}

// ReputationHistory represents detailed points logs for dynamic calculations.
type ReputationHistory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	ActionType  string             `bson:"action_type" json:"action_type"` // "prediction_correct", "context_approved", "context_helpful", "poll_vote"
	Points      int                `bson:"points" json:"points"`
	ReferenceID primitive.ObjectID `bson:"reference_id,omitempty" json:"reference_id,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// FollowFandomRequest represents the payload to follow a universe.
type FollowFandomRequest struct {
	UniverseID string `json:"universe_id" binding:"required"`
}

// LeaderboardEntry represents a scoreboard record.
type LeaderboardEntry struct {
	Username        string `json:"username"`
	ReputationScore int    `json:"reputation_score"`
	Rank            int    `json:"rank"`
}
