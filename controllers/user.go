package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/models"
)

var userController *models.UserModel

// UpdateProfileRequest represents the profile update payload
type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2"`
	Phone string `json:"phone" binding:"omitempty,min=10"`
	Bio   string `json:"bio" binding:"omitempty,max=500"`
}

// UpdatePreferenceRequest represents a single preference update
type UpdatePreferenceRequest struct {
	Field string `json:"field" binding:"required,oneof=visibility progress leaderboard"`
	Value bool   `json:"value"`
}

// UpdateThemeRequest represents a theme update
type UpdateThemeRequest struct {
	Theme string `json:"theme" binding:"required,oneof=light dark"`
}

// InitializeUserController initializes the user controller with required dependencies
func InitializeUserController(db *mongo.Database) {
	userController = models.NewUserModel(db)
}

// GetProfile returns the authenticated user's profile information
func GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := userController.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile handles profile update requests
func UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build update map with only provided fields
	updates := bson.M{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}

	// Check if there's anything to update
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Update user
	err := userController.UpdateUser(userID.(string), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Get updated user
	user, err := userController.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

// GetUserStats returns statistics about the user
func GetUserStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Implement actual stats logic
	// For now, return mock data
	stats := gin.H{
		"userID":           userID,
		"totalPractices":   0,
		"currentLevel":     3,
		"experiencePoints": 0,
		"achievements":     []string{},
	}

	c.JSON(http.StatusOK, stats)
}

// UpdatePreference handles updating a single preference field
func UpdatePreference(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build the update path based on the field
	updateField := "preferences.profile." + req.Field
	updates := bson.M{updateField: req.Value}

	// Update preference
	err := userController.UpdateUser(userID.(string), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preference"})
		return
	}

	// Get updated user
	user, err := userController.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Preference updated successfully",
		"preferences": user.Preferences,
	})
}

// GetPreferences returns user preferences
func GetPreferences(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	user, err := userController.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.Preferences)
}

// UpdateTheme handles updating the user's theme preference
func UpdateTheme(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update theme preference
	updates := bson.M{"preferences.theme": req.Theme}
	err := userController.UpdateUser(userID.(string), updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update theme"})
		return
	}

	// Get updated user
	user, err := userController.GetUserByID(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Theme updated successfully",
		"theme":   user.Preferences.Theme,
	})
}
