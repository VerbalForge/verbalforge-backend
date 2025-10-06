package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/models"
)

var passageController *models.PassageModel

// InitializePassageController initializes the passage controller with required dependencies
func InitializePassageController(db *mongo.Database) {
	passageController = models.NewPassageModel(db)
}

// GetPassages returns a paginated list of passages
func GetPassages(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	difficultyLevel := c.Query("difficulty")
	showNew := c.Query("new") == "true"

	// Build filter
	filter := bson.M{}
	if difficultyLevel != "" {
		filter["difficulty_level"] = difficultyLevel
	}
	if showNew {
		// Passages generated in the last 24 hours
		oneDayAgo := time.Now().Add(-24 * time.Hour)
		filter["metadata.created_at"] = bson.M{"$gte": oneDayAgo.Format(time.RFC3339)}
	}

	// Calculate skip
	skip := (page - 1) * limit

	passages, err := passageController.GetPartialPassages(filter, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch passages"})
		return
	}

	// Get total count
	total, err := passageController.CountPassages(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count passages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"passages":   passages,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + limit - 1) / limit,
	})
}

// GetPassageByID returns a single passage by its ID
func GetPassageByID(c *gin.Context) {
	passageID := c.Param("id")

	passage, err := passageController.GetPassageByID(passageID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Passage not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch passage"})
		return
	}

	c.JSON(http.StatusOK, passage)
}
