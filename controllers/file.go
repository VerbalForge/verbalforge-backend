package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var uploadDir string

type FileResponse struct {
	URL string `json:"url"`
}

// InitializeFileController sets up the upload directory
func InitializeFileController(dir string) {
	uploadDir = dir
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %v", err))
	}
}

// UploadFile handles file upload requests
func UploadFile(c *gin.Context) {
	// Get the file from the form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Validate file type (only images for now)
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only image files are allowed"})
		return
	}

	// Validate file size (max 5MB)
	const maxSize = 5 * 1024 * 1024 // 5MB
	if header.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large (max 5MB)"})
		return
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d_%s%s", timestamp, generateRandomString(8), ext)

	// Create full file path
	filePath := filepath.Join(uploadDir, filename)

	// Save the file
	if err := c.SaveUploadedFile(header, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return the file URL (you might want to use a CDN or cloud storage in production)
	fileURL := fmt.Sprintf("/uploads/%s", filename)

	c.JSON(http.StatusOK, FileResponse{
		URL: fileURL,
	})
}

// Simple random string generator (you might want to use a more robust solution in production)
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
