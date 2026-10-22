package middleware

import (
	"net/http"
	"strings"

	"eastwater-backend/internal/core/common"
	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware validates incoming API Key if configured.
// If expectedKey is empty, authentication is bypassed (open mode).
func APIKeyAuthMiddleware(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.Next()
			return
		}

		// Check x-api-key header
		apiKey := c.GetHeader("x-api-key")
		if apiKey == "" {
			// Also support Authorization: Bearer <token>
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey != expectedKey {
			common.ErrorResponse(c, http.StatusUnauthorized, "Invalid or missing API key")
			c.Abort()
			return
		}

		c.Next()
	}
}
