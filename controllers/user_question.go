package controllers

import (
	"net/http"

	"verbalforge-backend/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

var userQuestionController *UserQuestionController

type UserQuestionController struct {
	model *models.UserQuestionModel
}

func NewUserQuestionController(model *models.UserQuestionModel) *UserQuestionController {
	return &UserQuestionController{model: model}
}

// InitializeUserQuestionController sets up the controller with database connection
func InitializeUserQuestionController(db *mongo.Database) {
	model := models.NewUserQuestionModel(db)
	userQuestionController = NewUserQuestionController(model)
}

// SubmitQuestionAttempt records a user's attempt on a standalone question (TC, SE)
// POST /user/questions/:question_id/attempt
func (uqc *UserQuestionController) SubmitQuestionAttempt(c *gin.Context) {
	questionID := c.Param("question_id")

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var requestBody struct {
		Solved    bool `json:"solved"`
		TimeTaken int  `json:"time_taken"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userQuestion, err := uqc.model.UpsertUserQuestion(userIDStr, questionID, requestBody.Solved, requestBody.TimeTaken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attempt"})
		return
	}

	c.JSON(http.StatusOK, userQuestion)
}

// SubmitPassageAttempt records a user's attempt on a passage question (RC)
// POST /user/passages/:passage_id/attempt
func (uqc *UserQuestionController) SubmitPassageAttempt(c *gin.Context) {
	passageID := c.Param("passage_id")

	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var requestBody struct {
		QuestionID string `json:"question_id"`
		Solved     bool   `json:"solved"`
		TimeTaken  int    `json:"time_taken"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Record the question attempt
	userQuestion, err := uqc.model.UpsertUserQuestion(userIDStr, requestBody.QuestionID, requestBody.Solved, requestBody.TimeTaken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record question attempt"})
		return
	}

	// Derive passage progress
	passageProgress, err := uqc.model.GetPassageProgress(userIDStr, passageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate passage progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"question": userQuestion,
		"passage":  passageProgress,
	})
}

// GetQuestionProgress retrieves a user's progress on a specific question
// GET /user/questions/:question_id/progress
func (uqc *UserQuestionController) GetQuestionProgress(c *gin.Context) {
	questionID := c.Param("question_id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	progress, err := uqc.model.GetUserQuestion(userIDStr, questionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve progress"})
		return
	}

	if progress == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No progress found"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetPassageProgress retrieves a user's derived progress on a specific passage
// GET /user/passages/:passage_id/progress
func (uqc *UserQuestionController) GetPassageProgress(c *gin.Context) {
	passageID := c.Param("passage_id")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	progress, err := uqc.model.GetPassageProgress(userIDStr, passageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve passage progress"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetBulkQuestionProgress gets progress for multiple questions
// POST /user/questions/progress/bulk
func (uqc *UserQuestionController) GetBulkQuestionProgress(c *gin.Context) {
	var requestBody struct {
		QuestionIDs []string `json:"question_ids"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	progress, err := uqc.model.GetUserQuestionsByIDs(userIDStr, requestBody.QuestionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve question progress"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// GetBulkPassageProgress retrieves user's derived progress for multiple passages
// POST /user/passages/progress/bulk
func (uqc *UserQuestionController) GetBulkPassageProgress(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var requestBody struct {
		PassageIDs []string `json:"passage_ids"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	progress, err := uqc.model.GetBulkPassageProgress(userIDStr, requestBody.PassageIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve passage progress"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// Wrapper functions for router
func SubmitQuestionAttempt(c *gin.Context) {
	userQuestionController.SubmitQuestionAttempt(c)
}

func SubmitPassageAttempt(c *gin.Context) {
	userQuestionController.SubmitPassageAttempt(c)
}

func GetQuestionProgress(c *gin.Context) {
	userQuestionController.GetQuestionProgress(c)
}

func GetPassageProgress(c *gin.Context) {
	userQuestionController.GetPassageProgress(c)
}

func GetBulkQuestionProgress(c *gin.Context) {
	userQuestionController.GetBulkQuestionProgress(c)
}

func GetBulkPassageProgress(c *gin.Context) {
	userQuestionController.GetBulkPassageProgress(c)
}
