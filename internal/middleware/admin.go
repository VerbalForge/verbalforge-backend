package middleware

import (
	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/repository"
	"verbalforge-backend/internal/utils"
)

// AdminMiddleware checks if the authenticated user is an admin
// Must be used after AuthMiddleware
func AdminMiddleware(userRepo *repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			utils.UnauthorizedResponse(c, "User not authenticated")
			c.Abort()
			return
		}

		// Get user from database to check admin status
		user, err := userRepo.FindByID(userID)
		if err != nil {
			utils.UnauthorizedResponse(c, "User not found")
			c.Abort()
			return
		}

		// Check if user has admin privileges
		if !user.IsAdmin {
			utils.ForbiddenResponse(c, "Admin access required")
			c.Abort()
			return
		}

		c.Set("isAdmin", true)
		c.Next()
	}
}

// CheckAdminStatus checks if user is admin and returns boolean (doesn't abort)
func CheckAdminStatus(c *gin.Context, userRepo *repository.UserRepository) bool {
	userID, exists := GetUserID(c)
	if !exists {
		return false
	}

	user, err := userRepo.FindByID(userID)
	if err != nil {
		return false
	}

	return user.IsAdmin
}
