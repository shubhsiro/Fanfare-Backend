package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ScoringService struct {
	predRepo repository.PredictionRepository
	userRepo repository.UserRepository
	ctxRepo  repository.ContextRepository
}

// NewScoringService instantiates the prediction points-resolution engine.
func NewScoringService(
	predRepo repository.PredictionRepository,
	userRepo repository.UserRepository,
	ctxRepo repository.ContextRepository,
) *ScoringService {
	return &ScoringService{
		predRepo: predRepo,
		userRepo: userRepo,
		ctxRepo:  ctxRepo,
	}
}

// ResolvePrediction closes an active prediction, evaluates participations, awards points, and logs reputation.
func (s *ScoringService) ResolvePrediction(ctx context.Context, predID primitive.ObjectID, correctOption string, source string) error {
	pred, err := s.predRepo.GetPredictionByID(ctx, predID)
	if err != nil {
		return err
	}

	if pred.Status != "open" {
		return errors.New("prediction is already resolved or cancelled")
	}

	// Validate options
	validOption := false
	for _, opt := range pred.Options {
		if opt == correctOption {
			validOption = true
			break
		}
	}
	if !validOption {
		return fmt.Errorf("option '%s' is not in the list of allowed choices", correctOption)
	}

	// Fetch all participations to compute statistics
	participations, err := s.predRepo.ListParticipationsByPrediction(ctx, predID)
	if err != nil {
		return err
	}

	totalVotes := len(participations)
	winningVotes := 0
	for _, part := range participations {
		if part.SelectedOption == correctOption {
			winningVotes++
		}
	}

	// Calculate minority ratio
	var minorityRatio float64
	if totalVotes > 0 {
		minorityRatio = float64(winningVotes) / float64(totalVotes)
	}

	basePoints := 100

	// Resolve each user's participation
	for _, part := range participations {
		isCorrect := part.SelectedOption == correctOption
		pointsAwarded := 0

		if isCorrect {
			// 1. Confidence scaling
			multiplier := 1.0
			switch part.Confidence {
			case "low":
				multiplier = 0.5
			case "medium":
				multiplier = 1.0
			case "high":
				multiplier = 1.5
			}

			points := float64(basePoints) * multiplier

			// 2. Minority Bonus (only if winning fraction is less than 50% of the total voters)
			if minorityRatio < 0.5 && totalVotes > 1 {
				bonus := float64(basePoints) * (1.0 - minorityRatio)
				points += bonus
			}

			pointsAwarded = int(points)
		}

		// Update participation record
		part.IsCorrect = isCorrect
		part.PointsAwarded = pointsAwarded
		if err := s.predRepo.UpdateParticipation(ctx, &part); err != nil {
			return fmt.Errorf("failed to update user %s participation: %w", part.UserID.Hex(), err)
		}

		// Update user reputation profile if correct
		if isCorrect && pointsAwarded > 0 {
			// Create reputation log
			logEntry := &models.ReputationHistory{
				UserID:      part.UserID,
				ActionType:  "prediction_correct",
				Points:      pointsAwarded,
				ReferenceID: part.ID,
			}
			if err := s.ctxRepo.CreateReputationLog(ctx, logEntry); err != nil {
				return fmt.Errorf("failed to create reputation history log: %w", err)
			}

			// Update User Profile Score
			user, err := s.userRepo.GetByID(ctx, part.UserID)
			if err != nil {
				return fmt.Errorf("failed to load user profile: %w", err)
			}
			user.ReputationScore += pointsAwarded
			if err := s.userRepo.Update(ctx, user); err != nil {
				return fmt.Errorf("failed to update user reputation score: %w", err)
			}
		}
	}

	// Finalize prediction status
	now := time.Now()
	pred.Status = "resolved"
	pred.CorrectOption = &correctOption
	pred.ResolutionDate = &now
	pred.ResolutionSource = &source

	if err := s.predRepo.UpdatePrediction(ctx, pred); err != nil {
		return fmt.Errorf("failed to save resolved prediction: %w", err)
	}

	return nil
}
