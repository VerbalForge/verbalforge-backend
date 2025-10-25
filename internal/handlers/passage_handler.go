package handlers

import (
	"net/http"

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
