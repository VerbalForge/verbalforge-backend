package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
) // QuestionHandler handles question HTTP requests
type QuestionHandler struct {
	questionService *services.QuestionService
}

// NewQuestionHandler creates a new question handler
func NewQuestionHandler(questionService *services.QuestionService) *QuestionHandler {
	return &QuestionHandler{
		questionService: questionService,
	}
}

// GetQuestionByID retrieves a question by ID
func (h *QuestionHandler) GetQuestionByID(c *gin.Context) {
	id := c.Param("id")

	question, err := h.questionService.GetQuestionByID(id)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, question)
}

// GetQuestionProgress retrieves user progress for a question
func (h *QuestionHandler) GetQuestionProgress(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	questionID := c.Param("question_id")

	progress, err := h.questionService.GetQuestionProgress(userID, questionID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, progress)
}

// GetBulkQuestionProgress retrieves progress for multiple questions
func (h *QuestionHandler) GetBulkQuestionProgress(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.BulkProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	progress, err := h.questionService.GetBulkQuestionProgress(userID, req.IDs)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, progress)
}

// SubmitQuestionAttempt submits a question attempt
func (h *QuestionHandler) SubmitQuestionAttempt(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	questionID := c.Param("question_id")

	var req models.SubmitQuestionAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	userQuestion, err := h.questionService.SubmitQuestionAttempt(userID, questionID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, userQuestion)
}
