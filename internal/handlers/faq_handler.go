package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// FAQHandler handles FAQ HTTP requests
type FAQHandler struct {
	service *services.FAQService
}

// NewFAQHandler creates a new FAQ handler
func NewFAQHandler(service *services.FAQService) *FAQHandler {
	return &FAQHandler{
		service: service,
	}
}

// CreateFAQ handles the creation of a new FAQ
// @Summary Create a new FAQ
// @Tags FAQ
// @Accept json
// @Produce json
// @Param request body models.CreateFAQRequest true "FAQ details"
// @Success 201 {object} models.FAQ
// @Router /faq [post]
func (h *FAQHandler) CreateFAQ(c *gin.Context) {
	var req models.CreateFAQRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	faq, err := h.service.CreateFAQ(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create FAQ")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, faq)
}

// GetAllFAQs retrieves all FAQs (Admin only)
// @Summary Get all FAQs
// @Tags FAQ
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {array} models.FAQ
// @Router /admin/faq [get]
func (h *FAQHandler) GetAllFAQs(c *gin.Context) {
	status := c.Query("status")

	faqs, err := h.service.GetAllFAQs(c.Request.Context(), status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve FAQs")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"faqs": faqs})
}

// UpdateFAQStatus updates the status of an FAQ (Admin only)
// @Summary Update FAQ status
// @Tags FAQ
// @Accept json
// @Produce json
// @Param id path string true "FAQ ID"
// @Param request body models.UpdateFAQStatusRequest true "Status update"
// @Success 200
// @Router /admin/faq/{id}/status [patch]
func (h *FAQHandler) UpdateFAQStatus(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateFAQStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.service.UpdateFAQStatus(c.Request.Context(), id, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update FAQ status")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "FAQ status updated successfully"})
}

// DeleteFAQ deletes an FAQ (Admin only)
// @Summary Delete FAQ
// @Tags FAQ
// @Param id path string true "FAQ ID"
// @Success 200
// @Router /admin/faq/{id} [delete]
func (h *FAQHandler) DeleteFAQ(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteFAQ(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete FAQ")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "FAQ deleted successfully"})
}
