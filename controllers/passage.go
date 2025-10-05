package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Build filter
	filter := bson.M{}

	// Calculate skip
	skip := (page - 1) * limit

	// Get passages
	passages, err := passageController.GetPassages(filter, limit, skip)
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
	passageIDStr := c.Param("id")

	// Parse as ObjectID
	passageID, err := primitive.ObjectIDFromHex(passageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid passage ID"})
		return
	}

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
