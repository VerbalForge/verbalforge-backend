package handlers

import (
	"net/http"
	"strconv"

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

// Admin-specific handlers

// AdminGetAllQuestions returns all questions with filters for admin (uses existing service)
func (h *QuestionHandler) AdminGetAllQuestions(c *gin.Context) {
	// Admin check is done by AdminMiddleware in router

	published := c.Query("published") // "true", "false", or "" (all)
	questionType := c.Query("type")
	difficulty := c.Query("difficulty")
	search := c.Query("search")
	limitStr := c.DefaultQuery("limit", "0") // 0 = no limit for admin
	skipStr := c.DefaultQuery("skip", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	skip, _ := strconv.ParseInt(skipStr, 10, 64)

	questions, total, err := h.questionService.GetQuestions(published, questionType, difficulty, search, limit, skip)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"questions": questions,
		"total":     total,
	})
}

// AdminUpdateQuestion updates a question (full update)
func (h *QuestionHandler) AdminUpdateQuestion(c *gin.Context) {
	questionID := c.Param("id")

	var updateData models.Question
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := h.questionService.UpdateQuestion(questionID, &updateData); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Question updated successfully"})
}

// AdminDeleteQuestion deletes a question and handles cleanup
func (h *QuestionHandler) AdminDeleteQuestion(c *gin.Context) {
	questionID := c.Param("id")

	if err := h.questionService.DeleteQuestion(questionID); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Question deleted successfully"})
}

// AdminPublishQuestion toggles publish status
func (h *QuestionHandler) AdminPublishQuestion(c *gin.Context) {
	questionID := c.Param("id")

	var req struct {
		Publish bool `json:"publish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	publishedAt, err := h.questionService.PublishQuestion(questionID, req.Publish)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	status := "unpublished"
	if req.Publish {
		status = "published"
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Question " + status + " successfully",
		"published_at": publishedAt,
	})
}

// AdminBulkDeleteQuestions deletes multiple questions
func (h *QuestionHandler) AdminBulkDeleteQuestions(c *gin.Context) {
	var req struct {
		QuestionIDs []string `json:"question_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if len(req.QuestionIDs) == 0 {
		utils.BadRequestResponse(c, "No question IDs provided")
		return
	}

	if err := h.questionService.BulkDeleteQuestions(req.QuestionIDs); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Questions deleted successfully",
		"count":   len(req.QuestionIDs),
	})
}

// AdminBulkPublishQuestions publishes or unpublishes multiple questions
func (h *QuestionHandler) AdminBulkPublishQuestions(c *gin.Context) {
	var req struct {
		QuestionIDs []string `json:"question_ids" binding:"required"`
		Publish     bool     `json:"publish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if len(req.QuestionIDs) == 0 {
		utils.BadRequestResponse(c, "No question IDs provided")
		return
	}

	if err := h.questionService.BulkPublishQuestions(req.QuestionIDs, req.Publish); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	status := "unpublished"
	if req.Publish {
		status = "published"
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Questions " + status + " successfully",
		"count":   len(req.QuestionIDs),
	})
}
