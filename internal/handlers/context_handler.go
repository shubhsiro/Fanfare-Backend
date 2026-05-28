package handlers

import (
	"net/http"

	"fanfare-backend/internal/middleware"
	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContextHandler exposes endpoints for creating and rating community context cards.
type ContextHandler struct {
	ctxRepo  repository.ContextRepository
	userRepo repository.UserRepository
}

// NewContextHandler creates a ContextHandler.
func NewContextHandler(ctxRepo repository.ContextRepository, userRepo repository.UserRepository) *ContextHandler {
	return &ContextHandler{
		ctxRepo:  ctxRepo,
		userRepo: userRepo,
	}
}

// CreateContextCard handles POST /api/v1/context (Auth required)
// @Summary Submit a crowd-sourced context card attached to a moment, issue hub, or prediction
// @Tags Context
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.ContextCardRequest true "Context card content"
// @Success 201 {object} models.ContextCard
// @Failure 400 {object} map[string]string
// @Router /api/v1/context [post]
func (h *ContextHandler) CreateContextCard(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user session"})
		return
	}

	var req models.ContextCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	card := &models.ContextCard{
		UserID:           userID,
		WhatsAccurate:    req.WhatsAccurate,
		WhatsMissing:     req.WhatsMissing,
		WhatsSpeculative: req.WhatsSpeculative,
	}

	// Parse optional reference IDs
	if req.MomentID != "" {
		oid, err := primitive.ObjectIDFromHex(req.MomentID)
		if err == nil {
			card.MomentID = &oid
		}
	}
	if req.IssueHubID != "" {
		oid, err := primitive.ObjectIDFromHex(req.IssueHubID)
		if err == nil {
			card.IssueHubID = &oid
		}
	}
	if req.PredictionID != "" {
		oid, err := primitive.ObjectIDFromHex(req.PredictionID)
		if err == nil {
			card.PredictionID = &oid
		}
	}

	if err := h.ctxRepo.CreateCard(c.Request.Context(), card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create context card: " + err.Error()})
		return
	}

	// Award small reputation for context contribution
	repLog := &models.ReputationHistory{
		UserID:      userID,
		ActionType:  "context_approved",
		Points:      5,
		ReferenceID: card.ID,
	}
	_ = h.ctxRepo.CreateReputationLog(c.Request.Context(), repLog)

	// Update user score
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err == nil {
		user.ReputationScore += 5
		_ = h.userRepo.Update(c.Request.Context(), user)
	}

	c.JSON(http.StatusCreated, card)
}

// RateContextCard handles POST /api/v1/context/:id/rate (Auth required)
// @Summary Rate a context card as helpful, not helpful, or biased
// @Tags Context
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Context Card ObjectID"
// @Param body body models.ContextCardRateRequest true "Rating value"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/context/{id}/rate [post]
func (h *ContextHandler) RateContextCard(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user session"})
		return
	}

	cardIDStr := c.Param("id")
	cardID, err := primitive.ObjectIDFromHex(cardIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid context card ID"})
		return
	}

	var req models.ContextCardRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Fetch the card to update counts
	card, err := h.ctxRepo.GetCardByID(c.Request.Context(), cardID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Context card not found"})
		return
	}

	// Prevent self-rating
	if card.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot rate your own context card"})
		return
	}

	// Create rating record (compound unique index prevents duplicates)
	rating := &models.ContextCardRating{
		ContextCardID: cardID,
		UserID:        userID,
		Rating:        req.Rating,
	}
	if err := h.ctxRepo.CreateRating(c.Request.Context(), rating); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Update aggregated counts on the card
	switch req.Rating {
	case "helpful":
		card.HelpfulCount++
	case "not_helpful":
		card.NotHelpfulCount++
	case "biased":
		card.BiasedCount++
	}

	if err := h.ctxRepo.UpdateCard(c.Request.Context(), card); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card counts"})
		return
	}

	// Award reputation to the card author for helpful ratings
	if req.Rating == "helpful" {
		repLog := &models.ReputationHistory{
			UserID:      card.UserID,
			ActionType:  "context_helpful",
			Points:      2,
			ReferenceID: card.ID,
		}
		_ = h.ctxRepo.CreateReputationLog(c.Request.Context(), repLog)

		author, err := h.userRepo.GetByID(c.Request.Context(), card.UserID)
		if err == nil {
			author.ReputationScore += 2
			_ = h.userRepo.Update(c.Request.Context(), author)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rating submitted successfully"})
}
