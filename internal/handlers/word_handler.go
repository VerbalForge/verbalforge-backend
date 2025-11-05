package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// WordHandler handles word HTTP requests
type WordHandler struct {
	wordService *services.WordService
}

// NewWordHandler creates a new word handler
func NewWordHandler(wordService *services.WordService) *WordHandler {
	return &WordHandler{
		wordService: wordService,
	}
}

// AdminGetAllWords retrieves all words with filters
func (h *WordHandler) AdminGetAllWords(c *gin.Context) {
	// Parse query parameters
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "50"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("skip", "0"), 10, 64)
	search := c.Query("search")
	sourcesParam := c.Query("sources")

	// Parse sources (comma-separated)
	var sources []string
	if sourcesParam != "" {
		sources = strings.Split(sourcesParam, ",")
		// Trim whitespace
		for i, source := range sources {
			sources[i] = strings.TrimSpace(source)
		}
	}

	filters := models.WordFilters{
		Sources: sources,
		Search:  search,
		Limit:   int(limit),
		Offset:  int(offset),
	}

	response, err := h.wordService.GetWords(c.Request.Context(), filters, nil)
	if err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, response)
}

// AdminGetWordByID retrieves a word by ID
func (h *WordHandler) AdminGetWordByID(c *gin.Context) {
	id := c.Param("id")

	word, err := h.wordService.GetWordByID(c.Request.Context(), id)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, word)
}

// AdminCreateWord creates a new word
func (h *WordHandler) AdminCreateWord(c *gin.Context) {
	var word models.Word
	if err := c.ShouldBindJSON(&word); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := h.wordService.CreateWord(c.Request.Context(), &word); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"message": "Word created successfully",
		"word":    word,
	})
}

// AdminUpdateWord updates an existing word
func (h *WordHandler) AdminUpdateWord(c *gin.Context) {
	id := c.Param("id")

	var word models.Word
	if err := c.ShouldBindJSON(&word); err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	if err := h.wordService.UpdateWord(c.Request.Context(), id, &word); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Word updated successfully",
	})
}

// AdminDeleteWord deletes a word
func (h *WordHandler) AdminDeleteWord(c *gin.Context) {
	id := c.Param("id")

	if err := h.wordService.DeleteWord(c.Request.Context(), id); err != nil {
		utils.InternalServerErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Word deleted successfully",
	})
}
