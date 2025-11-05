package services

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// PassageService handles passage business logic
type PassageService struct {
	passageRepo      *repository.PassageRepository
	questionRepo     *repository.QuestionRepository
	userQuestionRepo *repository.UserQuestionRepository
	userRepo         *repository.UserRepository
	activityService  *UserActivityService
	cacheService     *CacheService
}

// NewPassageService creates a new passage service
func NewPassageService(
	passageRepo *repository.PassageRepository,
	questionRepo *repository.QuestionRepository,
	userQuestionRepo *repository.UserQuestionRepository,
	userRepo *repository.UserRepository,
	activityService *UserActivityService,
	cacheService *CacheService,
) *PassageService {
	return &PassageService{
		passageRepo:      passageRepo,
		questionRepo:     questionRepo,
		userQuestionRepo: userQuestionRepo,
		userRepo:         userRepo,
		activityService:  activityService,
		cacheService:     cacheService,
	}
}

// GetPassages retrieves passages with filters (published, difficulty, search, pagination)
func (s *PassageService) GetPassages(published, difficulty, search string, limit, skip int64) ([]models.Passage, int64, error) {
	filter := bson.M{}

	// Published filter based on PublishedAt field
	if published == "true" {
		filter["metadata.published_at"] = bson.M{"$ne": nil}
	} else if published == "false" {
		filter["metadata.published_at"] = nil
	}
	// If published is empty, show all passages

	if difficulty != "" {
		filter["difficulty"] = difficulty
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"passage_text": bson.M{"$regex": search, "$options": "i"}},
			{"topic": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	passages, err := s.passageRepo.FindAll(filter, limit, skip)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.passageRepo.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return passages, total, nil
}

// GetPassageByID retrieves a passage by ID
func (s *PassageService) GetPassageByID(id string) (*models.Passage, error) {
	passage, err := s.passageRepo.FindByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("passage not found")
		}
		return nil, err
	}
	return passage, nil
}

// GetPassageProgress retrieves user progress for a passage
func (s *PassageService) GetPassageProgress(userID, passageID string) (*models.PassageProgress, error) {
	// First get passage to get question IDs
	passage, err := s.passageRepo.FindByID(passageID)
	if err != nil {
		return nil, err
	}

	// Get user's attempts for these questions
	userQuestions, err := s.userQuestionRepo.FindByUserAndQuestions(userID, passage.QuestionIDs)
	if err != nil {
		return nil, err
	}

	solvedQuestionIDs := []string{}
	attempted := false

	for qID, uq := range userQuestions {
		attempted = true
		if uq.Solved {
			solvedQuestionIDs = append(solvedQuestionIDs, qID)
		}
	}

	progress := &models.PassageProgress{
		PassageID:         passageID,
		SolvedQuestionIDs: solvedQuestionIDs,
		TotalQuestions:    len(passage.QuestionIDs),
		Solved:            len(solvedQuestionIDs) == len(passage.QuestionIDs) && len(passage.QuestionIDs) > 0,
		Attempted:         attempted,
	}

	return progress, nil
}

// GetBulkPassageProgress retrieves progress for multiple passages
func (s *PassageService) GetBulkPassageProgress(userID string, passageIDs []string) (map[string]*models.PassageProgress, error) {
	result := make(map[string]*models.PassageProgress)

	for _, passageID := range passageIDs {
		progress, err := s.GetPassageProgress(userID, passageID)
		if err != nil {
			continue
		}
		result[passageID] = progress
	}

	return result, nil
}

// SubmitPassageAttempt submits attempts for all questions in a passage
func (s *PassageService) SubmitPassageAttempt(userID, passageID string, req *models.SubmitPassageAttemptRequest) ([]models.UserQuestion, error) {
	var results []models.UserQuestion
	totalXPGained := 0
	totalSolved := 0

	for _, attempt := range req.QuestionAttempts {
		// Get question to fetch difficulty and type
		question, err := s.questionRepo.FindByID(attempt.QuestionID)
		if err != nil {
			continue
		}

		// Calculate XP
		xpGained := 0
		if attempt.Solved {
			switch question.DifficultyLevel {
			case "easy":
				xpGained = 10
			case "medium":
				xpGained = 20
			case "hard":
				xpGained = 30
			default:
				xpGained = 15
			}
			totalSolved++
		}
		totalXPGained += xpGained

		// Upsert user question
		userQuestion, err := s.userQuestionRepo.Upsert(
			userID,
			attempt.QuestionID,
			attempt.Solved,
			attempt.TimeTaken,
			question.DifficultyLevel,
			question.QuestionType,
			xpGained,
		)

		if err == nil && userQuestion != nil {
			results = append(results, *userQuestion)

			// Log activity for this question (works for both attempted and solved)
			questionText := question.QuestionText
			if len(questionText) > 150 {
				questionText = questionText[:150] + "..."
			}

			activityMeta := models.QuestionActivityMetadata{
				QuestionID:      attempt.QuestionID,
				QuestionText:    questionText,
				Solved:          attempt.Solved,
				PassageID:       passageID,
				DifficultyLevel: question.DifficultyLevel,
				QuestionType:    question.QuestionType,
				TimeTaken:       attempt.TimeTaken,
				XPGained:        xpGained,
			}
			if err := s.activityService.LogQuestionActivity(userID, activityMeta); err != nil {
				fmt.Printf("Failed to log question activity: %v\n", err)
			}
		}
	}

	// Update user stats based on total results
	if totalSolved > 0 {
		// For now, increment for each solved question
		// TODO: Could optimize to do a single update with total counts
		for i := 0; i < totalSolved; i++ {
			if err := s.userRepo.IncrementSolved(userID, 0); err != nil {
				fmt.Printf("Failed to increment user solved count: %v\n", err)
			}
		}
		// Increment XP separately
		if totalXPGained > 0 {
			if err := s.userRepo.IncrementXP(userID, totalXPGained); err != nil {
				fmt.Printf("Failed to increment user XP: %v\n", err)
			}
		}
	} else {
		// Just increment attempts
		if err := s.userRepo.IncrementAttempts(userID); err != nil {
			fmt.Printf("Failed to increment user attempts: %v\n", err)
		}
	}

	return results, nil
}

// UpdatePassage updates a passage
func (s *PassageService) UpdatePassage(passageID string, updateData *models.Passage) error {
	updateData.ID = passageID
	updateData.Metadata.UpdatedAt = time.Now()

	return s.passageRepo.Update(passageID, updateData)
}

// DeletePassage deletes a passage and handles cleanup
func (s *PassageService) DeletePassage(passageID string) error {
	// Get the passage to find associated questions
	passage, err := s.passageRepo.FindByID(passageID)
	if err != nil {
		return errors.New("passage not found")
	}

	// Delete all associated questions
	for _, questionID := range passage.QuestionIDs {
		if err := s.questionRepo.Delete(questionID); err != nil {
			// Log error but continue deleting other questions
			fmt.Printf("Error deleting question %s: %v\n", questionID, err)
		}
	}

	return s.passageRepo.Delete(passageID)
}

// PublishPassage toggles publish status
func (s *PassageService) PublishPassage(passageID string, publish bool) (interface{}, error) {
	passage, err := s.passageRepo.FindByID(passageID)
	if err != nil {
		return nil, errors.New("passage not found")
	}

	// Set publish time
	var publishTime *time.Time
	if publish {
		now := time.Now()
		publishTime = &now
	} else {
		publishTime = nil
	}

	passage.Metadata.PublishedAt = publishTime
	passage.Metadata.UpdatedAt = time.Now()

	// Publish/unpublish all associated questions
	for _, questionID := range passage.QuestionIDs {
		question, err := s.questionRepo.FindByID(questionID)
		if err == nil {
			question.Metadata.PublishedAt = publishTime
			question.Metadata.UpdatedAt = time.Now()
			if err := s.questionRepo.Update(questionID, question); err != nil {
				// Log error but continue with other questions
				fmt.Printf("Error publishing question %s: %v\n", questionID, err)
			}
		}
	}

	if err := s.passageRepo.Update(passageID, passage); err != nil {
		return nil, err
	}

	// Invalidate practice cache when publishing/unpublishing
	if s.cacheService != nil {
		s.cacheService.InvalidateAll()
	}

	return passage.Metadata.PublishedAt, nil
}

// BulkDeletePassages deletes multiple passages
func (s *PassageService) BulkDeletePassages(passageIDs []string) error {
	for _, id := range passageIDs {
		if err := s.DeletePassage(id); err != nil {
			return fmt.Errorf("failed to delete passage %s: %w", id, err)
		}
	}
	return nil
}

// BulkPublishPassages publishes or unpublishes multiple passages
func (s *PassageService) BulkPublishPassages(passageIDs []string, publish bool) error {
	for _, id := range passageIDs {
		if _, err := s.PublishPassage(id, publish); err != nil {
			return fmt.Errorf("failed to publish passage %s: %w", id, err)
		}
	}
	return nil
}
