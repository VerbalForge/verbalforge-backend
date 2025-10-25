package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// PracticeHandler handles practice content HTTP requests
type PracticeHandler struct {
	practiceService *services.PracticeService
}

// NewPracticeHandler creates a new practice handler
func NewPracticeHandler(practiceService *services.PracticeService) *PracticeHandler {
	return &PracticeHandler{
		practiceService: practiceService,
	}
}

// GetPracticeItems retrieves paginated practice items with filters
// GET /practice?page=1&limit=20&difficulty=medium&type=all
func (h *PracticeHandler) GetPracticeItems(c *gin.Context) {
	// Parse query parameters
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		if parsedPage, err := strconv.Atoi(pageParam); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			if parsedLimit > 100 {
				parsedLimit = 100 // Cap at 100
			}
			limit = parsedLimit
		}
	}

	difficulty := c.DefaultQuery("difficulty", "all")
	itemType := c.DefaultQuery("type", "all")

	// Get practice items
	response, err := h.practiceService.GetPracticeItems(page, limit, difficulty, itemType)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// InvalidatePracticeCache invalidates the practice cache
// POST /practice/invalidate?difficulty=medium&type=all
func (h *PracticeHandler) InvalidatePracticeCache(c *gin.Context) {
	difficulty := c.DefaultQuery("difficulty", "")
	itemType := c.DefaultQuery("type", "")

	h.practiceService.InvalidateCache(difficulty, itemType)

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Cache invalidated successfully",
	})
}
