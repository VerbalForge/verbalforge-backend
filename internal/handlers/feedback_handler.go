package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// FeedbackHandler handles feedback HTTP requests
type FeedbackHandler struct {
	service *services.FeedbackService
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(service *services.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{
		service: service,
	}
}

// CreateFeedback handles the creation of new feedback
// @Summary Create new feedback
// @Tags Feedback
// @Accept json
// @Produce json
// @Param request body models.CreateFeedbackRequest true "Feedback details"
// @Success 201 {object} models.Feedback
// @Router /feedback [post]
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	var req models.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	feedback, err := h.service.CreateFeedback(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create feedback")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, feedback)
}

// GetAllFeedback retrieves all feedback (Admin only)
// @Summary Get all feedback
// @Tags Feedback
// @Produce json
// @Param type query string false "Filter by type"
// @Param status query string false "Filter by status"
// @Success 200 {array} models.Feedback
// @Router /admin/feedback [get]
func (h *FeedbackHandler) GetAllFeedback(c *gin.Context) {
	feedbackType := c.Query("type")
	status := c.Query("status")

	feedbacks, err := h.service.GetAllFeedback(c.Request.Context(), feedbackType, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve feedback")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"feedback": feedbacks})
}

// UpdateFeedbackStatus updates the status of feedback (Admin only)
// @Summary Update feedback status
// @Tags Feedback
// @Accept json
// @Produce json
// @Param id path string true "Feedback ID"
// @Param request body models.UpdateFeedbackStatusRequest true "Status update"
// @Success 200
// @Router /admin/feedback/{id}/status [patch]
func (h *FeedbackHandler) UpdateFeedbackStatus(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateFeedbackStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.service.UpdateFeedbackStatus(c.Request.Context(), id, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update feedback status")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Feedback status updated successfully"})
}

// DeleteFeedback deletes feedback (Admin only)
// @Summary Delete feedback
// @Tags Feedback
// @Param id path string true "Feedback ID"
// @Success 200
// @Router /admin/feedback/{id} [delete]
func (h *FeedbackHandler) DeleteFeedback(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteFeedback(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete feedback")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Feedback deleted successfully"})
}
