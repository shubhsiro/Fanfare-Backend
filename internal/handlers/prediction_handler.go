package handlers

import (
	"net/http"

	"fanfare-backend/internal/middleware"
	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"
	"fanfare-backend/internal/services"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PredictionHandler exposes prediction voting and resolution endpoints.
type PredictionHandler struct {
	predRepo       repository.PredictionRepository
	scoringService *services.ScoringService
}

// NewPredictionHandler creates a PredictionHandler.
func NewPredictionHandler(predRepo repository.PredictionRepository, scoringService *services.ScoringService) *PredictionHandler {
	return &PredictionHandler{
		predRepo:       predRepo,
		scoringService: scoringService,
	}
}

// VotePrediction handles POST /api/v1/predictions/:id/vote (Auth required)
// @Summary Place a prediction vote with a confidence level
// @Tags Predictions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Prediction ObjectID"
// @Param body body models.PredictionVoteRequest true "Vote details"
// @Success 201 {object} models.PredictionParticipation
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/predictions/{id}/vote [post]
func (h *PredictionHandler) VotePrediction(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user session"})
		return
	}

	predIDStr := c.Param("id")
	predID, err := primitive.ObjectIDFromHex(predIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prediction ID"})
		return
	}

	var req models.PredictionVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Verify prediction exists and is open
	pred, err := h.predRepo.GetPredictionByID(c.Request.Context(), predID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Prediction not found"})
		return
	}
	if pred.Status != "open" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This prediction is no longer accepting votes"})
		return
	}

	// Validate selected option
	validOption := false
	for _, opt := range pred.Options {
		if opt == req.SelectedOption {
			validOption = true
			break
		}
	}
	if !validOption {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option selected"})
		return
	}

	part := &models.PredictionParticipation{
		PredictionID:   predID,
		UserID:         userID,
		SelectedOption: req.SelectedOption,
		Confidence:     req.Confidence,
	}

	if err := h.predRepo.CreateParticipation(c.Request.Context(), part); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, part)
}

// ResolvePrediction handles POST /api/v1/predictions/:id/resolve (Auth required, admin)
// @Summary Resolve a prediction with the actual outcome
// @Tags Predictions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Prediction ObjectID"
// @Param body body models.PredictionResolveRequest true "Resolution details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/predictions/{id}/resolve [post]
func (h *PredictionHandler) ResolvePrediction(c *gin.Context) {
	predIDStr := c.Param("id")
	predID, err := primitive.ObjectIDFromHex(predIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prediction ID"})
		return
	}

	var req models.PredictionResolveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	if err := h.scoringService.ResolvePrediction(c.Request.Context(), predID, req.CorrectOption, req.Source); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prediction resolved and scores calculated"})
}
