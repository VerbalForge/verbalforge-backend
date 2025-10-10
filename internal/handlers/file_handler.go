package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/services"
	"verbalforge-backend/internal/utils"
)

// FileHandler handles file upload HTTP requests
type FileHandler struct {
	fileService *services.FileService
}

// NewFileHandler creates a new file handler
func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// UploadFile handles file upload
func (h *FileHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequestResponse(c, "No file uploaded")
		return
	}

	filename, err := h.fileService.UploadFile(file)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{
		"filename": filename,
		"url":      "/uploads/" + filename,
	})
}
