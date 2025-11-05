package services

import (
	"context"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

type UserWordService struct {
	repo            *repository.UserWordRepository
	wordRepo        *repository.WordRepository
	activityService *UserActivityService
}

func NewUserWordService(repo *repository.UserWordRepository, wordRepo *repository.WordRepository, activityService *UserActivityService) *UserWordService {
	return &UserWordService{
		repo:            repo,
		wordRepo:        wordRepo,
		activityService: activityService,
	}
}

func (s *UserWordService) GetUserWords(ctx context.Context, userID string) (*models.UserWord, error) {
	return s.repo.GetUserWords(ctx, userID)
}

func (s *UserWordService) UpdateWordStatus(ctx context.Context, userID, wordID, status string) error {
	// Get current user words to determine the previous state
	userWords, err := s.repo.GetUserWords(ctx, userID)
	if err != nil {
		return err
	}

	// Get word details for activity logging
	word, err := s.wordRepo.GetWordByID(ctx, wordID)
	if err != nil {
		// Continue even if word details can't be fetched
		word = &models.Word{ID: wordID, Word: "unknown"}
	}

	// Determine previous state
	wasKnown := contains(userWords.Known, wordID)
	wasPractice := contains(userWords.Practice, wordID)

	// Determine action type based on status change
	var actionType models.WordActionType
	switch status {
	case "known":
		actionType = models.ActionWordMarkedKnown
		err = s.repo.AddToKnown(ctx, userID, wordID)
	case "practice":
		actionType = models.ActionWordMarkedPractice
		err = s.repo.AddToPractice(ctx, userID, wordID)
	case "reset":
		// Determine which action to log based on previous state
		if wasKnown {
			actionType = models.ActionWordUnmarkedKnown
		} else if wasPractice {
			actionType = models.ActionWordUnmarkedPractice
		}
		err = s.repo.RemoveWordStatus(ctx, userID, wordID)
	default:
		err = s.repo.RemoveWordStatus(ctx, userID, wordID)
	}

	if err != nil {
		return err
	}

	// Get updated counts
	updatedWords, _ := s.repo.GetUserWords(ctx, userID)
	knownCount := len(updatedWords.Known)
	practiceCount := len(updatedWords.Practice)

	// Log activity only if there was an actual change
	if actionType != "" {
		metadata := models.WordActivityMetadata{
			WordID:        wordID,
			Word:          word.Word,
			ActionType:    actionType,
			KnownCount:    knownCount,
			PracticeCount: practiceCount,
		}

		// Log the word activity (this will also update daily summary)
		_ = s.activityService.LogWordActivity(userID, metadata)
	}

	return nil
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *UserWordService) ResetUserProgress(ctx context.Context, userID string) error {
	return s.repo.ResetUserProgress(ctx, userID)
}
