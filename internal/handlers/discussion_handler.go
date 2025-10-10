package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// DiscussionHandler handles discussion HTTP requests
type DiscussionHandler struct {
	discussionService *services.DiscussionService
}

// NewDiscussionHandler creates a new discussion handler
func NewDiscussionHandler(discussionService *services.DiscussionService) *DiscussionHandler {
	return &DiscussionHandler{
		discussionService: discussionService,
	}
}

// GetDiscussions retrieves all discussions
func (h *DiscussionHandler) GetDiscussions(c *gin.Context) {
	cursor := c.Query("cursor")
	sortBy := c.DefaultQuery("sortBy", "newest")

	limit := int64(20)
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseInt(limitParam, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var tags []string
	if tagsParam := c.Query("tags"); tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
	}

	response, err := h.discussionService.GetDiscussionsWithCursor(limit, cursor, sortBy, tags)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// GetDiscussionByID retrieves a discussion by ID
func (h *DiscussionHandler) GetDiscussionByID(c *gin.Context) {
	id := c.Param("id")

	discussion, err := h.discussionService.GetDiscussionByID(id)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, discussion)
}

// CreateDiscussion creates a new discussion
func (h *DiscussionHandler) CreateDiscussion(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	username, _ := middleware.GetUsername(c)

	var req models.CreateDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	discussion, err := h.discussionService.CreateDiscussion(userID, username, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, discussion)
}

// UpdateDiscussion updates a discussion
func (h *DiscussionHandler) UpdateDiscussion(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var req models.UpdateDiscussionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	discussion, err := h.discussionService.UpdateDiscussion(id, userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, discussion)
}

// DeleteDiscussion deletes a discussion
func (h *DiscussionHandler) DeleteDiscussion(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")

	err := h.discussionService.DeleteDiscussion(id, userID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Discussion deleted successfully", nil)
}

// IncrementDiscussionView increments view count
func (h *DiscussionHandler) IncrementDiscussionView(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")

	err := h.discussionService.IncrementView(id, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "View recorded", nil)
}

// ToggleDiscussionLike toggles like on a discussion
func (h *DiscussionHandler) ToggleDiscussionLike(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")

	liked, err := h.discussionService.ToggleLike(id, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"liked": liked})
}

// AddComment adds a comment to a discussion
func (h *DiscussionHandler) AddComment(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	username, _ := middleware.GetUsername(c)
	id := c.Param("id")

	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.discussionService.AddComment(id, userID, username, &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusCreated, "Comment added successfully", nil)
}

// UpdateComment updates a comment
func (h *DiscussionHandler) UpdateComment(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")
	commentID := c.Param("commentId")

	var req models.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	err := h.discussionService.UpdateComment(id, commentID, userID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Comment updated successfully", nil)
}

// DeleteComment deletes a comment
func (h *DiscussionHandler) DeleteComment(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")
	commentID := c.Param("commentId")

	err := h.discussionService.DeleteComment(id, commentID, userID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessMessageResponse(c, http.StatusOK, "Comment deleted successfully", nil)
}

// ToggleCommentLike toggles like on a comment
func (h *DiscussionHandler) ToggleCommentLike(c *gin.Context) {
	userID, ok := middleware.RequireAuth(c)
	if !ok {
		return
	}

	id := c.Param("id")
	commentID := c.Param("commentId")

	liked, err := h.discussionService.ToggleCommentLike(id, commentID, userID)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"liked": liked})
}

// GetUserDiscussions retrieves discussions by a user
func (h *DiscussionHandler) GetUserDiscussions(c *gin.Context) {
	username := c.Param("username")
	cursor := c.Query("cursor")

	limit := int64(20)
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseInt(limitParam, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	response, err := h.discussionService.GetUserDiscussionsWithCursor(username, limit, cursor)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// SearchDiscussions searches discussions
func (h *DiscussionHandler) SearchDiscussions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.BadRequestResponse(c, "Search query is required")
		return
	}

	cursor := c.Query("cursor")

	limit := int64(20)
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.ParseInt(limitParam, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	response, err := h.discussionService.SearchDiscussionsWithCursor(query, limit, cursor)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}
