package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// SupportTicketHandler handles support ticket HTTP requests
type SupportTicketHandler struct {
	service *services.SupportTicketService
}

// NewSupportTicketHandler creates a new support ticket handler
func NewSupportTicketHandler(service *services.SupportTicketService) *SupportTicketHandler {
	return &SupportTicketHandler{
		service: service,
	}
}

// CreateSupportTicket handles the creation of a new support ticket (Authenticated users only)
// @Summary Create a new support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param request body models.CreateSupportTicketRequest true "Support ticket details"
// @Success 201 {object} models.SupportTicket
// @Router /support [post]
func (h *SupportTicketHandler) CreateSupportTicket(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.CreateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	ticket, err := h.service.CreateSupportTicket(c.Request.Context(), userID, &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create support ticket")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, ticket)
}

// GetAllSupportTickets retrieves all support tickets (Admin only)
// @Summary Get all support tickets
// @Tags Support
// @Produce json
// @Param status query string false "Filter by status"
// @Success 200 {array} models.SupportTicket
// @Router /admin/support [get]
func (h *SupportTicketHandler) GetAllSupportTickets(c *gin.Context) {
	status := c.Query("status")

	tickets, err := h.service.GetAllSupportTickets(c.Request.Context(), status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve support tickets")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"tickets": tickets})
}

// GetUserSupportTickets retrieves support tickets for the authenticated user
// @Summary Get user's support tickets
// @Tags Support
// @Produce json
// @Success 200 {array} models.SupportTicket
// @Router /support/my-tickets [get]
func (h *SupportTicketHandler) GetUserSupportTickets(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	tickets, err := h.service.GetUserSupportTickets(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve support tickets")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"tickets": tickets})
}

// UpdateSupportTicket updates a support ticket (Admin only)
// @Summary Update support ticket
// @Tags Support
// @Accept json
// @Produce json
// @Param id path string true "Support ticket ID"
// @Param request body models.UpdateSupportTicketRequest true "Ticket update"
// @Success 200
// @Router /admin/support/{id} [patch]
func (h *SupportTicketHandler) UpdateSupportTicket(c *gin.Context) {
	id := c.Param("id")

	var req models.UpdateSupportTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := h.service.UpdateSupportTicket(c.Request.Context(), id, &req); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update support ticket")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Support ticket updated successfully"})
}

// DeleteSupportTicket deletes a support ticket (Admin only)
// @Summary Delete support ticket
// @Tags Support
// @Param id path string true "Support ticket ID"
// @Success 200
// @Router /admin/support/{id} [delete]
func (h *SupportTicketHandler) DeleteSupportTicket(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteSupportTicket(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete support ticket")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Support ticket deleted successfully"})
}
