package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/models"
)

var questionController *models.QuestionModel

// InitializeQuestionController initializes the question controller with required dependencies
func InitializeQuestionController(db *mongo.Database) {
	questionController = models.NewQuestionModel(db)
}

// GetQuestions returns a paginated list of questions with optional filters
func GetQuestions(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	difficultyLevel := c.Query("difficulty")
	questionType := c.Query("type")
	topic := c.Query("topic")
	showNew := c.Query("new") == "true"

	// Build filter
	filter := bson.M{}
	if difficultyLevel != "" {
		filter["difficulty_level"] = difficultyLevel
	}
	if questionType != "" {
		// Use regex to match question types that contain the base type
		// This allows "reading_comprehension" to match "READING_COMPREHENSION_SINGLE", etc.
		upperType := strings.ToUpper(questionType)
		filter["question_type"] = bson.M{"$regex": upperType, "$options": "i"}
	}
	if topic != "" {
		filter["topic"] = topic
	}
	if showNew {
		// Questions generated in the last 24 hours
		yesterday := time.Now().Add(-24 * time.Hour)
		filter["metadata.created_at"] = bson.M{"$gte": yesterday}
	}

	// Calculate skip
	skip := (page - 1) * limit

	// Get partial questions (only essential fields)
	questions, err := questionController.GetPartialQuestions(filter, limit, skip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch questions"})
		return
	}

	// Truncate question text to 150 characters
	for i := range questions {
		if len(questions[i].QuestionText) > 150 {
			questions[i].QuestionText = questions[i].QuestionText[:150] + "..."
		}
	}

	// Get total count
	total, err := questionController.CountQuestions(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count questions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"questions":  questions,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + limit - 1) / limit,
	})
}

// GetQuestionByID returns a single question by its ID
func GetQuestionByID(c *gin.Context) {
	questionIDStr := c.Param("id")

	// Try to parse as ObjectID first
	questionID, err := primitive.ObjectIDFromHex(questionIDStr)
	if err != nil {
		// If not a valid ObjectID, try to find by question_id
		question, err := questionController.GetQuestionByQuestionID(questionIDStr)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch question"})
			return
		}
		c.JSON(http.StatusOK, question)
		return
	}

	question, err := questionController.GetQuestionByID(questionID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
			return
		}
		log.Printf("Error fetching question by ObjectID %s: %v", questionID.Hex(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch question"})
		return
	}

	c.JSON(http.StatusOK, question)
}
