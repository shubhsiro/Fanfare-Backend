package middleware

import (
	"net/http"
	"strings"

	"fanfare-backend/internal/services"

	"github.com/gin-gonic/gin"
)

const (
	// AuthHeaderKey is the HTTP header used to carry the Bearer token.
	AuthHeaderKey = "Authorization"
	// BearerPrefix is the expected token scheme prefix.
	BearerPrefix = "Bearer "
	// ContextUserID is the key stored in gin.Context after successful auth.
	ContextUserID = "user_id"
	// ContextUsername is the key stored in gin.Context after successful auth.
	ContextUsername = "username"
)

// AuthMiddleware validates the JWT from the Authorization header and injects
// user claims into the Gin request context for downstream handlers.
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthHeaderKey)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must use Bearer scheme"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Inject user identity into request context for downstream handlers
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Next()
	}
}
