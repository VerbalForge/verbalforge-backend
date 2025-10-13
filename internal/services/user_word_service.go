package services

import (
	"context"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

type UserWordService struct {
	repo *repository.UserWordRepository
}

func NewUserWordService(repo *repository.UserWordRepository) *UserWordService {
	return &UserWordService{
		repo: repo,
	}
}

func (s *UserWordService) GetUserWords(ctx context.Context, userID string) (*models.UserWord, error) {
	return s.repo.GetUserWords(ctx, userID)
}

func (s *UserWordService) UpdateWordStatus(ctx context.Context, userID, wordID, status string) error {
	switch status {
	case "known":
		return s.repo.AddToKnown(ctx, userID, wordID)
	case "practice":
		return s.repo.AddToPractice(ctx, userID, wordID)
	case "reset":
		return s.repo.RemoveWordStatus(ctx, userID, wordID)
	default:
		return s.repo.RemoveWordStatus(ctx, userID, wordID)
	}
}

func (s *UserWordService) ResetUserProgress(ctx context.Context, userID string) error {
	return s.repo.ResetUserProgress(ctx, userID)
}
