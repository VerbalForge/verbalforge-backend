package services

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
	"verbalforge-backend/internal/utils"
)

// PassageService handles passage business logic
type PassageService struct {
	passageRepo      *repository.PassageRepository
	questionRepo     *repository.QuestionRepository
	userQuestionRepo *repository.UserQuestionRepository
	userRepo         *repository.UserRepository
	activityService  *UserActivityService
}

// NewPassageService creates a new passage service
func NewPassageService(
	passageRepo *repository.PassageRepository,
	questionRepo *repository.QuestionRepository,
	userQuestionRepo *repository.UserQuestionRepository,
	userRepo *repository.UserRepository,
	activityService *UserActivityService,
) *PassageService {
	return &PassageService{
		passageRepo:      passageRepo,
		questionRepo:     questionRepo,
		userQuestionRepo: userQuestionRepo,
		userRepo:         userRepo,
		activityService:  activityService,
	}
}

// GetPassages retrieves passages with filters
func (s *PassageService) GetPassages(difficulty string, limit, skip int64) ([]models.Passage, error) {
	filter := bson.M{}

	if difficulty != "" {
		filter["difficulty"] = difficulty
	}

	return s.passageRepo.FindAll(filter, limit, skip)
}

// GetPartialPassages retrieves partial passages with filters
// GetPartialPassagesWithCursor retrieves passages with cursor-based pagination
func (s *PassageService) GetPartialPassagesWithCursor(difficulty string, limit int64, cursor string) (*models.PassagesResponse, error) {
	filter := bson.M{}

	if difficulty != "" {
		filter["difficulty"] = difficulty
	}

	// Decode cursor
	var lastID, lastCreatedAt string
	if cursor != "" {
		cursorData, err := utils.DecodeCursor(cursor)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		if cursorData != nil {
			lastID = cursorData.LastID
			lastCreatedAt = cursorData.LastCreatedAt
		}
	}

	// Fetch passages with cursor
	passages, err := s.passageRepo.FindPartialWithCursor(filter, limit, lastID, lastCreatedAt)
	if err != nil {
		return nil, err
	}

	// Determine if there are more results
	hasMore := int64(len(passages)) > limit
	if hasMore {
		passages = passages[:limit] // Remove the extra item
	}

	// Create next cursor if there are more results
	var nextCursor string
	if hasMore && len(passages) > 0 {
		lastPassage := passages[len(passages)-1]
		cursorData := utils.CursorData{
			LastID:        lastPassage.ID,
			LastCreatedAt: lastPassage.CreatedAt,
		}
		nextCursor, err = utils.EncodeCursor(cursorData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	// Create previous cursor if we have a current cursor (not on first page)
	var previousCursor string
	if cursor != "" && len(passages) > 0 {
		firstPassage := passages[0]
		cursorData := utils.CursorData{
			LastID:        firstPassage.ID,
			LastCreatedAt: firstPassage.CreatedAt,
		}
		previousCursor, err = utils.EncodeCursor(cursorData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode previous cursor: %w", err)
		}
	}

	// Get total count
	total, err := s.passageRepo.Count(filter)
	if err != nil {
		return nil, err
	}

	return &models.PassagesResponse{
		Passages:       passages,
		Total:          total,
		Limit:          int(limit),
		NextCursor:     nextCursor,
		PreviousCursor: previousCursor,
		HasMore:        hasMore,
	}, nil
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
