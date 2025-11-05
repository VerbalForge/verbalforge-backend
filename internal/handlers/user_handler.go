package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// UserHandler handles user HTTP requests
type UserHandler struct {
	userService     *services.UserService
	activityService *services.UserActivityService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, activityService *services.UserActivityService) *UserHandler {
	return &UserHandler{
		userService:     userService,
		activityService: activityService,
	}
}

// GetProfile retrieves user's own profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, user)
}

// GetUserProfile retrieves a user profile by username (public)
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	username := c.Param("username")

	profile, err := h.userService.GetProfileByUsername(username)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	// Check if the requesting user is viewing their own profile
	requestingUserID, exists := middleware.GetUserID(c)
	isOwnProfile := exists && requestingUserID == profile.User.ID

	// Check privacy settings - allow access if it's the user's own profile
	if !profile.User.Preferences.Profile.Visibility && !isOwnProfile {
		utils.ForbiddenResponse(c, "This profile is private")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, profile)
}

// UpdateProfile updates user profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, user)
}

// GetPreferences retrieves user preferences
func (h *UserHandler) GetPreferences(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, user.Preferences)
}

// UpdatePreference updates a user preference
func (h *UserHandler) UpdatePreference(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.UpdatePreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.userService.UpdatePreference(userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Preference updated successfully", nil)
}

// UpdateTheme updates user theme
func (h *UserHandler) UpdateTheme(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	var req models.UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.userService.UpdateTheme(userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Theme updated successfully", nil)
}

// IncrementProfileView increments profile view count
func (h *UserHandler) IncrementProfileView(c *gin.Context) {
	username := c.Param("username")

	profile, err := h.userService.GetProfileByUsername(username)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	err = h.userService.IncrementProfileView(profile.User.ID)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to increment view count")
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "View count incremented", nil)
}

// GetUserStats retrieves user statistics
func (h *UserHandler) GetUserStats(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	stats, err := h.userService.GetUserStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, stats)
}

// GetUserStatsByUsername retrieves user statistics by username
func (h *UserHandler) GetUserStatsByUsername(c *gin.Context) {
	username := c.Param("username")

	stats, err := h.userService.GetUserStatsByUsername(username)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, stats)
}

// GetUserRecentActivity retrieves recent activity for a user by username
func (h *UserHandler) GetUserRecentActivity(c *gin.Context) {
	username := c.Param("username")

	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	// Get user by username
	profile, err := h.userService.GetProfileByUsername(username)
	if err != nil {
		utils.NotFoundResponse(c, "User not found")
		return
	}

	// Get activities from activity service
	activities, err := h.activityService.GetRecentActivities(profile.User.ID, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, activities)
}

// GetLeaderboard retrieves the leaderboard
func (h *UserHandler) GetLeaderboard(c *gin.Context) {
	limit := 100
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			limit = parsedLimit
		}
	}

	leaderboard, err := h.userService.GetLeaderboard(limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, leaderboard)
}

// UpdateRanks updates all user ranks
func (h *UserHandler) UpdateRanks(c *gin.Context) {
	err := h.userService.UpdateRanks()
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Ranks updated successfully", nil)
}

// GetActivityCalendar retrieves activity calendar
func (h *UserHandler) GetActivityCalendar(c *gin.Context) {
	username := c.Param("username")

	profile, err := h.userService.GetProfileByUsername(username)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	days := 365
	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil {
			days = parsedDays
		}
	}

	// Get timezone from query param, default to UTC
	timezone := c.Query("timezone")
	if timezone == "" {
		timezone = "UTC"
	}

	activity, err := h.userService.GetActivityCalendar(profile.User.ID, days, timezone)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, activity)
}

// GetWordStatistics retrieves word learning statistics
func (h *UserHandler) GetWordStatistics(c *gin.Context) {
	username := c.Param("username")

	profile, err := h.userService.GetProfileByUsername(username)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	days := 30 // Default to last 30 days
	if daysParam := c.Query("days"); daysParam != "" {
		if parsedDays, err := strconv.Atoi(daysParam); err == nil {
			days = parsedDays
		}
	}

	stats, err := h.activityService.GetWordStatistics(profile.User.ID, days)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, stats)
}

// Admin-specific handlers

// AdminGetUserActivity returns detailed activity logs for a specific user
func (h *UserHandler) AdminGetUserActivity(c *gin.Context) {
	userID := c.Param("userId")
	questionType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "100")
	skipStr := c.DefaultQuery("skip", "0")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	skip, _ := strconv.ParseInt(skipStr, 10, 64)

	activities, total, err := h.activityService.GetUserActivityLogs(userID, questionType, startDate, endDate, limit, skip)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"activities": activities,
		"total":      total,
		"limit":      limit,
		"skip":       skip,
	})
}

// AdminGetUserStats returns aggregated stats for a specific user
func (h *UserHandler) AdminGetUserStats(c *gin.Context) {
	userID := c.Param("userId")

	stats, err := h.activityService.GetUserAggregatedStats(userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, stats)
}

// AdminGetTotalUsers returns the total number of registered users
func (h *UserHandler) AdminGetTotalUsers(c *gin.Context) {
	count, err := h.userService.GetTotalUserCount()
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"total": count})
}
