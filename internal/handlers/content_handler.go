package handlers

import (
	"net/http"
	"strconv"

	"fanfare-backend/internal/models"
	"fanfare-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentHandler exposes endpoints for universes, titles, moments feed, and issue hubs.
type ContentHandler struct {
	contentRepo repository.ContentRepository
	predRepo    repository.PredictionRepository
}

// NewContentHandler creates a ContentHandler with the content and prediction repositories.
func NewContentHandler(contentRepo repository.ContentRepository, predRepo repository.PredictionRepository) *ContentHandler {
	return &ContentHandler{contentRepo: contentRepo, predRepo: predRepo}
}

// ListUniverses handles GET /api/v1/universes
// @Summary List all pop-culture universes/franchises
// @Tags Universes
// @Produce json
// @Success 200 {array} models.Universe
// @Router /api/v1/universes [get]
func (h *ContentHandler) ListUniverses(c *gin.Context) {
	unis, err := h.contentRepo.ListUniverses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, unis)
}

// GetUniverse handles GET /api/v1/universes/:slug
// @Summary Get a universe by its URL slug with its titles
// @Tags Universes
// @Produce json
// @Param slug path string true "Universe slug"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/v1/universes/{slug} [get]
func (h *ContentHandler) GetUniverse(c *gin.Context) {
	slug := c.Param("slug")
	uni, err := h.contentRepo.GetUniverseBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe not found"})
		return
	}

	// Also fetch titles belonging to this universe
	titles, err := h.contentRepo.ListTitlesByUniverse(c.Request.Context(), uni.ID)
	if err != nil {
		titles = nil // Non-critical: return universe without titles
	}

	c.JSON(http.StatusOK, gin.H{
		"universe": uni,
		"titles":   titles,
	})
}

// GetTitle handles GET /api/v1/titles/:slug
// @Summary Get a specific title by its URL slug
// @Tags Titles
// @Produce json
// @Param slug path string true "Title slug"
// @Success 200 {object} models.Title
// @Failure 404 {object} map[string]string
// @Router /api/v1/titles/{slug} [get]
func (h *ContentHandler) GetTitle(c *gin.Context) {
	slug := c.Param("slug")
	title, err := h.contentRepo.GetTitleBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Title not found"})
		return
	}
	c.JSON(http.StatusOK, title)
}

// ListMoments handles GET /api/v1/moments
// @Summary Get paginated news feed moments (infinite scroll)
// @Tags Moments
// @Produce json
// @Param limit query int false "Page size (default 20)"
// @Param offset query int false "Offset for pagination (default 0)"
// @Success 200 {array} models.Moment
// @Router /api/v1/moments [get]
func (h *ContentHandler) ListMoments(c *gin.Context) {
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	moments, err := h.contentRepo.ListMoments(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, moments)
}

// GetMoment handles GET /api/v1/moments/:id
// @Summary Get a specific moment by its ID
// @Tags Moments
// @Produce json
// @Param id path string true "Moment ObjectID"
// @Success 200 {object} models.Moment
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/moments/{id} [get]
func (h *ContentHandler) GetMoment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid moment ID format"})
		return
	}

	moment, err := h.contentRepo.GetMomentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Moment not found"})
		return
	}
	c.JSON(http.StatusOK, moment)
}

// ListIssueHubs handles GET /api/v1/issues
// @Summary List all long-running debate hubs
// @Tags Issues
// @Produce json
// @Param sort query string false "Sort order: 'activity' or 'recent' (default)"
// @Success 200 {array} models.IssueHub
// @Router /api/v1/issues [get]
func (h *ContentHandler) ListIssueHubs(c *gin.Context) {
	sortOption := c.DefaultQuery("sort", "recent")
	hubs, err := h.contentRepo.ListIssueHubs(c.Request.Context(), sortOption)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hubs)
}

// GetIssueHub handles GET /api/v1/issues/:slug
// @Summary Get a specific debate hub by slug with timeline and evidence
// @Tags Issues
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Issue hub slug"
// @Success 200 {object} models.IssueHub
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/issues/{slug} [get]
func (h *ContentHandler) GetIssueHub(c *gin.Context) {
	slug := c.Param("slug")
	hub, err := h.contentRepo.GetIssueHubBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue hub not found"})
		return
	}
	c.JSON(http.StatusOK, hub)
}

// ListMomentPredictions handles GET /api/v1/moments/:id/predictions (Auth required)
// @Summary List all prediction questions attached to a specific moment
// @Description Returns both open and resolved predictions linked to this news moment card
// @Tags Moments
// @Produce json
// @Security BearerAuth
// @Param id path string true "Moment ObjectID"
// @Success 200 {array} models.Prediction
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/moments/{id}/predictions [get]
func (h *ContentHandler) ListMomentPredictions(c *gin.Context) {
	idStr := c.Param("id")
	momentID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid moment ID format"})
		return
	}

	preds, err := h.predRepo.ListPredictionsByMoment(c.Request.Context(), momentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preds == nil {
		preds = []models.Prediction{}
	}
	c.JSON(http.StatusOK, preds)
}

// ListIssueHubPredictions handles GET /api/v1/issues/:slug/predictions (Auth required)
// @Summary List all prediction questions within an issue hub debate
// @Tags Issues
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Issue hub slug"
// @Success 200 {array} models.Prediction
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/issues/{slug}/predictions [get]
func (h *ContentHandler) ListIssueHubPredictions(c *gin.Context) {
	slug := c.Param("slug")
	hub, err := h.contentRepo.GetIssueHubBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issue hub not found"})
		return
	}

	preds, err := h.predRepo.ListPredictionsByIssueHub(c.Request.Context(), hub.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if preds == nil {
		preds = []models.Prediction{}
	}
	c.JSON(http.StatusOK, preds)
}
