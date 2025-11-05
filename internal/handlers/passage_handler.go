package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
) // PassageHandler handles passage HTTP requests
type PassageHandler struct {
	passageService *services.PassageService
}

// NewPassageHandler creates a new passage handler
func NewPassageHandler(passageService *services.PassageService) *PassageHandler {
	return &PassageHandler{
		passageService: passageService,
	}
}

// GetPassageByID retrieves a passage by ID
func (h *PassageHandler) GetPassageByID(c *gin.Context) {
	id := c.Param("id")

	passage, err := h.passageService.GetPassageByID(id)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, passage)
}

// GetPassageProgress retrieves user progress for a passage
func (h *PassageHandler) GetPassageProgress(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	passageID := c.Param("passage_id")

	progress, err := h.passageService.GetPassageProgress(userID, passageID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, progress)
}

// GetBulkPassageProgress retrieves progress for multiple passages
func (h *PassageHandler) GetBulkPassageProgress(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.BulkProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	progress, err := h.passageService.GetBulkPassageProgress(userID, req.IDs)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, progress)
}

// SubmitPassageAttempt submits attempts for all questions in a passage
func (h *PassageHandler) SubmitPassageAttempt(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	passageID := c.Param("passage_id")

	var req models.SubmitPassageAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	results, err := h.passageService.SubmitPassageAttempt(userID, passageID, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Get passage progress
	passageProgress, err := h.passageService.GetPassageProgress(userID, passageID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	// Return the first question's progress (since frontend expects single question response)
	// and passage progress
	response := map[string]interface{}{
		"question": results[0],
		"passage":  passageProgress,
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// Admin-specific handlers

// AdminGetAllPassages returns all passages with filters for admin
func (h *PassageHandler) AdminGetAllPassages(c *gin.Context) {
	published := c.Query("published") // "true", "false", or "" (all)
	difficulty := c.Query("difficulty")
	search := c.Query("search")
	limitStr := c.DefaultQuery("limit", "0") // 0 = no limit for admin
	skipStr := c.DefaultQuery("skip", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	skip, _ := strconv.ParseInt(skipStr, 10, 64)

	passages, total, err := h.passageService.GetPassages(published, difficulty, search, limit, skip)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"passages": passages,
		"total":    total,
	})
}

// AdminUpdatePassage updates a passage (full update)
func (h *PassageHandler) AdminUpdatePassage(c *gin.Context) {
	passageID := c.Param("id")

	var updateData models.Passage
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := h.passageService.UpdatePassage(passageID, &updateData); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Passage updated successfully"})
}

// AdminDeletePassage deletes a passage and handles cleanup
func (h *PassageHandler) AdminDeletePassage(c *gin.Context) {
	passageID := c.Param("id")

	if err := h.passageService.DeletePassage(passageID); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Passage deleted successfully"})
}

// AdminPublishPassage toggles publish status
func (h *PassageHandler) AdminPublishPassage(c *gin.Context) {
	passageID := c.Param("id")

	var req struct {
		Publish bool `json:"publish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	publishedAt, err := h.passageService.PublishPassage(passageID, req.Publish)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	status := "unpublished"
	if req.Publish {
		status = "published"
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Passage " + status + " successfully",
		"published_at": publishedAt,
	})
}

// AdminBulkDeletePassages deletes multiple passages
func (h *PassageHandler) AdminBulkDeletePassages(c *gin.Context) {
	var req struct {
		PassageIDs []string `json:"passage_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if len(req.PassageIDs) == 0 {
		utils.BadRequestResponse(c, "No passage IDs provided")
		return
	}

	if err := h.passageService.BulkDeletePassages(req.PassageIDs); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Passages deleted successfully",
		"count":   len(req.PassageIDs),
	})
}

// AdminBulkPublishPassages publishes or unpublishes multiple passages
func (h *PassageHandler) AdminBulkPublishPassages(c *gin.Context) {
	var req struct {
		PassageIDs []string `json:"passage_ids" binding:"required"`
		Publish    bool     `json:"publish"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if len(req.PassageIDs) == 0 {
		utils.BadRequestResponse(c, "No passage IDs provided")
		return
	}

	if err := h.passageService.BulkPublishPassages(req.PassageIDs, req.Publish); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	status := "unpublished"
	if req.Publish {
		status = "published"
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Passages " + status + " successfully",
		"count":   len(req.PassageIDs),
	})
}
