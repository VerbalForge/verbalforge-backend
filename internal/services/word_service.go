package services

import (
	"context"
	"fmt"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

type WordService struct {
	wordRepo *repository.WordRepository
}

func NewWordService(wordRepo *repository.WordRepository) *WordService {
	return &WordService{
		wordRepo: wordRepo,
	}
}

// GetWords retrieves words with filters and pagination
func (s *WordService) GetWords(ctx context.Context, filters models.WordFilters, userWords *models.UserWord) (*models.WordsResponse, error) {
	// Validate filters
	if filters.Limit < 0 {
		filters.Limit = 50
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	response, err := s.wordRepo.GetWords(ctx, filters, userWords)
	if err != nil {
		return nil, err
	}

	// Calculate pagination info
	if filters.Limit > 0 {
		response.TotalPages = int((response.Total + int64(filters.Limit) - 1) / int64(filters.Limit))
		response.Page = filters.Page
	}

	return response, nil
}

// GetWordByID retrieves a single word by ID
func (s *WordService) GetWordByID(ctx context.Context, id string) (*models.Word, error) {
	if id == "" {
		return nil, fmt.Errorf("word ID is required")
	}

	return s.wordRepo.GetWordByID(ctx, id)
}

// GetAvailableSources returns all available sources for filtering
func (s *WordService) GetAvailableSources(ctx context.Context) ([]string, error) {
	return s.wordRepo.GetAvailableSources(ctx)
}

// SearchWords performs a text search on words
func (s *WordService) SearchWords(ctx context.Context, query string, sources []string, limit int, userWords *models.UserWord) (*models.WordsResponse, error) {
	filters := models.WordFilters{
		Search:  query,
		Sources: sources,
		Limit:   limit,
		Offset:  0,
	}

	return s.wordRepo.GetWords(ctx, filters, userWords)
}
