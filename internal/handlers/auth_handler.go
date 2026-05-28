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

// AuthHandler exposes registration, login, profile, and fandom endpoints.
type AuthHandler struct {
	authService *services.AuthService
	repService  *services.ReputationService
	ctxRepo     repository.ContextRepository
}

// NewAuthHandler creates an AuthHandler with the required services.
func NewAuthHandler(authService *services.AuthService, repService *services.ReputationService, ctxRepo repository.ContextRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		repService:  repService,
		ctxRepo:     ctxRepo,
	}
}

// Register handles POST /api/v1/auth/register
// @Summary Register a new user account
// @Description Creates a new user with username, email, and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.RegisterRequest true "Registration details"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login handles POST /api/v1/auth/login
// @Summary Authenticate and obtain a JWT
// @Description Validates credentials and returns a signed JWT token with user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMe handles GET /api/v1/users/me (Auth required)
// @Summary Get the currently authenticated user's profile and stats
// @Description Returns the full profile of the token holder — no need to know your own username.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.UserProfileStats
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	username, _ := c.Get(middleware.ContextUsername)
	stats, err := h.repService.GetUserProfileStats(c.Request.Context(), username.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetUserProfile handles GET /api/v1/users/:username
// @Summary Get any user's public profile with dynamic stats and badges
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username"
// @Success 200 {object} services.UserProfileStats
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{username} [get]
func (h *AuthHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")
	stats, err := h.repService.GetUserProfileStats(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// FollowFandom handles POST /api/v1/users/follow-fandom (Auth required)
// @Summary Follow a franchise/universe
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body models.FollowFandomRequest true "Universe ID to follow"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /api/v1/users/follow-fandom [post]
func (h *AuthHandler) FollowFandom(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user session"})
		return
	}

	var req models.FollowFandomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	universeID, err := primitive.ObjectIDFromHex(req.UniverseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid universe_id format"})
		return
	}

	if err := h.repService.FollowUniverse(c.Request.Context(), userID, universeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully followed universe"})
}

// GetScoreboard handles GET /api/v1/scoreboard
// @Summary Get the global or fandom-specific leaderboard
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param universe_id query string false "Filter by universe ID"
// @Param limit query int false "Max entries (default 10)"
// @Success 200 {array} models.LeaderboardEntry
// @Failure 401 {object} map[string]string
// @Router /api/v1/scoreboard [get]
func (h *AuthHandler) GetScoreboard(c *gin.Context) {
	limit := 10
	universeIDStr := c.Query("universe_id")

	if universeIDStr != "" {
		universeID, err := primitive.ObjectIDFromHex(universeIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid universe_id"})
			return
		}
		entries, err := h.repService.GetFandomScoreboard(c.Request.Context(), universeID, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, entries)
		return
	}

	entries, err := h.repService.GetGlobalScoreboard(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// GetReputationHistory handles GET /api/v1/users/me/reputation (Auth required)
// @Summary Get the authenticated user's reputation points audit trail
// @Description Returns a chronological list of every reputation-earning action
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ReputationHistory
// @Failure 401 {object} map[string]string
// @Router /api/v1/users/me/reputation [get]
func (h *AuthHandler) GetReputationHistory(c *gin.Context) {
	userIDStr, _ := c.Get(middleware.ContextUserID)
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user session"})
		return
	}

	history, err := h.ctxRepo.GetReputationHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if history == nil {
		history = []models.ReputationHistory{}
	}
	c.JSON(http.StatusOK, history)
}
