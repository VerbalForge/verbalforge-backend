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

// GoogleLogin handles Google OAuth login
// @Summary Google OAuth login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.GoogleOAuthRequest true "Google ID token"
// @Success 200 {object} utils.APIResponse
// @Router /auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req models.GoogleOAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	user, token, err := h.authService.GoogleLogin(c.Request.Context(), req.Token)
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

// RefreshToken generates a new JWT token for the authenticated user
// @Summary Refresh authentication token
// @Tags auth
// @Produce json
// @Success 200 {object} utils.APIResponse
// @Router /auth/refresh [post]
// @Security BearerAuth
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	email, _ := middleware.GetEmail(c)
	username, _ := middleware.GetUsername(c)

	// Generate new token
	newToken, err := h.authService.RefreshToken(userID, email, username)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"token": newToken,
	})
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

// ForgotPassword initiates the password reset process
// @Summary Request password reset
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Email address"
// @Success 200 {object} utils.APIResponse
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	token, err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	// In production, you would send this token via email
	// For now, we return it in the response for testing
	utils.SuccessMessageResponse(c, http.StatusOK, "Password reset instructions sent to your email", gin.H{
		"resetToken": token, // Remove this in production
	})
}

// VerifyResetToken verifies if a reset token is valid
// @Summary Verify password reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param token query string true "Reset token"
// @Success 200 {object} utils.APIResponse
// @Router /auth/verify-reset-token [get]
func (h *AuthHandler) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequestResponse(c, "Reset token is required")
		return
	}

	user, err := h.authService.VerifyResetToken(token)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"valid": true,
		"email": user.Email,
	})
}

// ResetPassword resets the password using a valid reset token
// @Summary Reset password with token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} utils.APIResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Password reset successfully - you can now login with your new password", nil)
}
