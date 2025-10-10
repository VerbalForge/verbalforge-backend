package services

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// FileService handles file upload business logic
type FileService struct {
	uploadDir string
}

// NewFileService creates a new file service
func NewFileService(uploadDir string) *FileService {
	return &FileService{
		uploadDir: uploadDir,
	}
}

// UploadFile handles file upload
func (s *FileService) UploadFile(file *multipart.FileHeader) (string, error) {
	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return "", errors.New("file size exceeds 5MB limit")
	}

	// Validate file type
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExtensions[ext] {
		return "", errors.New("invalid file type. Allowed: jpg, jpeg, png, gif, pdf")
	}

	// Generate unique filename
	filename := uuid.New().String() + ext

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(s.uploadDir, os.ModePerm); err != nil {
		return "", errors.New("failed to create upload directory")
	}

	// Save file
	src, err := file.Open()
	if err != nil {
		return "", errors.New("failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(s.uploadDir, filename))
	if err != nil {
		return "", errors.New("failed to create file")
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", errors.New("failed to save file")
	}

	return filename, nil
}
