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

// SentimentHandler exposes sentiment poll voting and dashboard aggregation endpoints.
type SentimentHandler struct {
	predRepo       repository.PredictionRepository
	scoringService *services.ScoringService
}

// NewSentimentHandler creates a SentimentHandler.
func NewSentimentHandler(predRepo repository.PredictionRepository, scoringService *services.ScoringService) *SentimentHandler {
	return &SentimentHandler{
		predRepo:       predRepo,
		scoringService: scoringService,
	}
}

// VoteSentimentPoll handles POST /api/v1/sentiment/polls/:id/vote (Auth required)
// @Summary Submit a quick sentiment opinion poll vote
// @Tags Sentiment
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Sentiment Poll ObjectID"
// @Param body body models.SentimentVoteRequest true "Vote details with region and age band"
// @Success 201 {object} models.SentimentVote
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/sentiment/polls/{id}/vote [post]
func (h *SentimentHandler) VoteSentimentPoll(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user session"})
		return
	}

	pollIDStr := c.Param("id")
	pollID, err := primitive.ObjectIDFromHex(pollIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	var req models.SentimentVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Verify poll exists
	poll, err := h.predRepo.GetSentimentPollByID(c.Request.Context(), pollID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sentiment poll not found"})
		return
	}

	// Validate selected option
	validOption := false
	for _, opt := range poll.Options {
		if opt == req.SelectedOption {
			validOption = true
			break
		}
	}
	if !validOption {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid option selected"})
		return
	}

	vote := &models.SentimentVote{
		SentimentPollID: pollID,
		UserID:          userID,
		SelectedOption:  req.SelectedOption,
		Region:          req.Region,
		AgeBand:         req.AgeBand,
	}

	if err := h.predRepo.CreateSentimentVote(c.Request.Context(), vote); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Award participation point (+1 reputation)
	_ = h.scoringService.AwardPollVotePoints(c.Request.Context(), userID, pollID)

	c.JSON(http.StatusCreated, vote)
}

// GetSentimentDashboard handles GET /api/v1/sentiment/dashboards/:target_type/:target_id
// @Summary Get aggregated sentiment distribution for an issue or title
// @Tags Sentiment
// @Produce json
// @Param target_type path string true "Target type: 'issue' or 'title'"
// @Param target_id path string true "Target ObjectID"
// @Param cohort query string false "Group by cohort: 'region' or 'age_band'"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /api/v1/sentiment/dashboards/{target_type}/{target_id} [get]
func (h *SentimentHandler) GetSentimentDashboard(c *gin.Context) {
	targetType := c.Param("target_type")
	targetIDStr := c.Param("target_id")

	targetID, err := primitive.ObjectIDFromHex(targetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid target ID"})
		return
	}

	// Fetch all polls for this target
	polls, err := h.predRepo.GetSentimentPollsForTarget(c.Request.Context(), targetType, targetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cohort := c.Query("cohort")

	type PollResult struct {
		Poll         interface{}            `json:"poll"`
		Distribution interface{}            `json:"distribution"`
	}

	var results []PollResult
	for _, poll := range polls {
		if cohort == "region" || cohort == "age_band" {
			dist, err := h.predRepo.GetSentimentDistributionByCohort(c.Request.Context(), poll.ID, cohort)
			if err != nil {
				continue
			}
			results = append(results, PollResult{Poll: poll, Distribution: dist})
		} else {
			dist, err := h.predRepo.GetSentimentDistribution(c.Request.Context(), poll.ID)
			if err != nil {
				continue
			}
			results = append(results, PollResult{Poll: poll, Distribution: dist})
		}
	}

	if results == nil {
		results = []PollResult{}
	}

	c.JSON(http.StatusOK, gin.H{
		"target_type": targetType,
		"target_id":   targetIDStr,
		"dashboards":  results,
	})
}
