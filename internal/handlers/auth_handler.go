package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User registration details"
// @Success 201 {object} utils.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	user, token, err := h.authService.Register(&req)
	if err != nil {
		utils.ConflictResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

// Login handles user login
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} utils.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	user, token, err := h.authService.Login(&req)
	if err != nil {
		utils.UnauthorizedResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}

// GetCurrentUser retrieves the current authenticated user
// @Summary Get current user
// @Tags auth
// @Produce json
// @Success 200 {object} utils.APIResponse
// @Router /auth/me [get]
// @Security BearerAuth
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, user)
}

// ChangePassword handles password change
// @Summary Change password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ChangePasswordRequest true "Password change request"
// @Success 200 {object} utils.APIResponse
// @Router /auth/change-password [post]
// @Security BearerAuth
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.authService.ChangePassword(userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Password changed successfully", nil)
}

// DeleteAccount handles account deletion
// @Summary Delete account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Password confirmation"
// @Success 200 {object} utils.APIResponse
// @Router /auth/delete-account [delete]
// @Security BearerAuth
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.authService.DeleteAccount(userID, req.Password)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Account deleted successfully", nil)
}
