package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/utils"
	"verbalforge-backend/pkg/logger"
)

// AuthMiddleware validates JWT tokens and sets user context
// It also handles automatic token refresh when the token is close to expiry
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.UnauthorizedResponse(c, "Authorization header required")
			c.Abort()
			return
		}

		// Check for Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			utils.UnauthorizedResponse(c, "Bearer token required")
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.ValidateJWT(tokenString, jwtSecret)
		if err != nil {
			logger.Errorf("JWT validation error: %v", err)
			utils.UnauthorizedResponse(c, "Invalid token")
			c.Abort()
			return
		}

		// Set user context
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)

		// Check if token needs refresh
		// Exclude /auth/me endpoint from automatic refresh
		path := c.Request.URL.Path
		if !strings.HasSuffix(path, "/auth/me") && utils.ShouldRefreshToken(claims) {
			// Generate new token
			newToken, err := utils.GenerateJWT(claims.UserID, claims.Email, claims.Username, jwtSecret)
			if err != nil {
				logger.Errorf("Token refresh error: %v", err)
			} else {
				// Send new token in response header
				c.Header("X-New-Token", newToken)
				logger.Infof("Token refreshed for user: %s", claims.Username)
			}
		}

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUsername extracts username from context
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetEmail extracts email from context
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	return email.(string), true
}

// RequireAuth ensures user is authenticated and aborts if not
func RequireAuth(c *gin.Context) (string, bool) {
	userID, exists := GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		c.Abort()
		return "", false
	}
	return userID, true
}
