package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type WordController struct {
	wordService     *services.WordService
	userWordService *services.UserWordService
}

func NewWordController(wordService *services.WordService, userWordService *services.UserWordService) *WordController {
	return &WordController{
		wordService:     wordService,
		userWordService: userWordService,
	}
}

// GetWords handles GET /words
func (c *WordController) GetWords(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	sourcesParam := ctx.Query("sources")
	search := ctx.Query("search")
	knownParam := ctx.Query("known")
	practiceParam := ctx.Query("practice")
	limitParam := ctx.DefaultQuery("limit", "30")
	pageParam := ctx.DefaultQuery("page", "1")

	// Parse limit and page
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}

	// Calculate offset from page
	offset := (page - 1) * limit

	// Parse sources (comma-separated)
	var sources []string
	if sourcesParam != "" {
		sources = strings.Split(sourcesParam, ",")
		// Trim whitespace
		for i, source := range sources {
			sources[i] = strings.TrimSpace(source)
		}
	}

	// Parse boolean filters
	var known, practice *bool
	if knownParam != "" {
		if knownBool, err := strconv.ParseBool(knownParam); err == nil {
			known = &knownBool
		}
	}
	if practiceParam != "" {
		if practiceBool, err := strconv.ParseBool(practiceParam); err == nil {
			practice = &practiceBool
		}
	}

	// Build filters
	filters := models.WordFilters{
		Sources:  sources,
		Search:   search,
		Known:    known,
		Practice: practice,
		Limit:    limit,
		Offset:   offset,
		Page:     page,
	}

	// Get user's word progress
	userWords, err := c.userWordService.GetUserWords(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user progress"})
		return
	}

	// Get words
	response, err := c.wordService.GetWords(ctx.Request.Context(), filters, userWords)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add user progress to response
	response.UserWords = userWords

	ctx.JSON(http.StatusOK, response)
}

// GetWordByID handles GET /words/:id
func (c *WordController) GetWordByID(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Word ID is required"})
		return
	}

	word, err := c.wordService.GetWordByID(ctx.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, word)
}

// GetSources handles GET /words/sources
func (c *WordController) GetSources(ctx *gin.Context) {
	sources, err := c.wordService.GetAvailableSources(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"sources": sources})
}

// SearchWords handles GET /words/search
func (c *WordController) SearchWords(ctx *gin.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	sourcesParam := ctx.Query("sources")
	limitParam := ctx.DefaultQuery("limit", "20")

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	var sources []string
	if sourcesParam != "" {
		sources = strings.Split(sourcesParam, ",")
		for i, source := range sources {
			sources[i] = strings.TrimSpace(source)
		}
	}

	// Get user words for progress filtering
	userID, _ := ctx.Get("userID")
	var userWords *models.UserWord
	if userID != nil {
		userWords, _ = c.userWordService.GetUserWords(ctx.Request.Context(), userID.(string))
	}

	response, err := c.wordService.SearchWords(ctx.Request.Context(), query, sources, limit, userWords)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateWordStatus handles POST/PUT /words/:id/status
func (c *WordController) UpdateWordStatus(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	wordID := ctx.Param("id")
	if wordID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Word ID is required"})
		return
	}

	var request struct {
		Status string `json:"status" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate status values
	if request.Status != "known" && request.Status != "practice" && request.Status != "reset" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'known', 'practice', or 'reset'"})
		return
	}

	err := c.userWordService.UpdateWordStatus(ctx.Request.Context(), userID.(string), wordID, request.Status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Word status updated successfully"})
}

// UpdateBulkWordStatus handles POST /words/bulk/status
func (c *WordController) UpdateBulkWordStatus(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request struct {
		Updates []struct {
			WordID string `json:"wordId" binding:"required"`
			Status string `json:"status" binding:"required"`
		} `json:"updates" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Process bulk updates
	for _, update := range request.Updates {
		// Validate status values
		if update.Status != "known" && update.Status != "practice" && update.Status != "reset" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be 'known', 'practice', or 'reset'"})
			return
		}

		err := c.userWordService.UpdateWordStatus(ctx.Request.Context(), userID.(string), update.WordID, update.Status)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Bulk word status updated successfully"})
}

// ResetUserProgress handles POST /words/progress/reset
func (c *WordController) ResetUserProgress(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	err := c.userWordService.ResetUserProgress(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User progress reset successfully"})
}

// GetUserProgress handles GET /words/progress
func (c *WordController) GetUserProgress(ctx *gin.Context) {
	// Get user ID from context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userWords, err := c.userWordService.GetUserWords(ctx.Request.Context(), userID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, userWords)
}
