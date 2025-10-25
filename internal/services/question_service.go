package services

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// QuestionService handles question business logic
type QuestionService struct {
	questionRepo     *repository.QuestionRepository
	userQuestionRepo *repository.UserQuestionRepository
	passageRepo      *repository.PassageRepository
	userRepo         *repository.UserRepository
	activityService  *UserActivityService
}

// NewQuestionService creates a new question service
func NewQuestionService(
	questionRepo *repository.QuestionRepository,
	userQuestionRepo *repository.UserQuestionRepository,
	passageRepo *repository.PassageRepository,
	userRepo *repository.UserRepository,
	activityService *UserActivityService,
) *QuestionService {
	return &QuestionService{
		questionRepo:     questionRepo,
		userQuestionRepo: userQuestionRepo,
		passageRepo:      passageRepo,
		userRepo:         userRepo,
		activityService:  activityService,
	}
}

// GetQuestions retrieves questions with filters
func (s *QuestionService) GetQuestions(questionType, difficulty string, limit, skip int64) ([]models.Question, error) {
	filter := bson.M{}

	if questionType != "" {
		filter["question_type"] = questionType
	}
	if difficulty != "" {
		filter["difficulty_level"] = difficulty
	}

	return s.questionRepo.FindAll(filter, limit, skip)
}

// GetQuestionByID retrieves a question by ID
func (s *QuestionService) GetQuestionByID(id string) (*models.Question, error) {
	question, err := s.questionRepo.FindByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("question not found")
		}
		return nil, err
	}
	return question, nil
}

// GetQuestionProgress retrieves user progress for a question
func (s *QuestionService) GetQuestionProgress(userID, questionID string) (*models.UserQuestionResponse, error) {
	userQuestion, err := s.userQuestionRepo.FindByUserAndQuestion(userID, questionID)
	if err != nil {
		return nil, err
	}

	if userQuestion == nil {
		return &models.UserQuestionResponse{
			UserID:     userID,
			QuestionID: questionID,
			Solved:     false,
			Attempted:  false,
			TimeTaken:  0,
			XPGained:   0,
		}, nil
	}

	return &models.UserQuestionResponse{
		UserID:          userQuestion.UserID,
		QuestionID:      questionID,
		Solved:          userQuestion.Solved,
		Attempted:       userQuestion.Attempted,
		LastAttemptAt:   &userQuestion.LastAttemptAt,
		LastSolvedAt:    userQuestion.LastSolvedAt,
		TimeTaken:       userQuestion.TimeTaken,
		DifficultyLevel: userQuestion.DifficultyLevel,
		QuestionType:    userQuestion.QuestionType,
		XPGained:        userQuestion.XPGained,
	}, nil
}

// GetBulkQuestionProgress retrieves progress for multiple questions
func (s *QuestionService) GetBulkQuestionProgress(userID string, questionIDs []string) (map[string]*models.UserQuestionResponse, error) {
	userQuestions, err := s.userQuestionRepo.FindByUserAndQuestions(userID, questionIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*models.UserQuestionResponse)
	for _, qID := range questionIDs {
		if uq, ok := userQuestions[qID]; ok {
			result[qID] = &models.UserQuestionResponse{
				UserID:          uq.UserID,
				QuestionID:      qID,
				Solved:          uq.Solved,
				Attempted:       uq.Attempted,
				LastAttemptAt:   &uq.LastAttemptAt,
				LastSolvedAt:    uq.LastSolvedAt,
				TimeTaken:       uq.TimeTaken,
				DifficultyLevel: uq.DifficultyLevel,
				QuestionType:    uq.QuestionType,
				XPGained:        uq.XPGained,
			}
		} else {
			result[qID] = &models.UserQuestionResponse{
				UserID:     userID,
				QuestionID: qID,
				Solved:     false,
				Attempted:  false,
				TimeTaken:  0,
				XPGained:   0,
			}
		}
	}

	return result, nil
}

// SubmitQuestionAttempt submits a question attempt
func (s *QuestionService) SubmitQuestionAttempt(userID, questionID string, req *models.SubmitQuestionAttemptRequest) (*models.UserQuestion, error) {
	// Get question to fetch difficulty and type
	question, err := s.questionRepo.FindByID(questionID)
	if err != nil {
		return nil, errors.New("question not found")
	}

	// Dereference the solved pointer
	solved := false
	if req.Solved != nil {
		solved = *req.Solved
	}

	// Calculate XP based on difficulty and whether it's solved
	xpGained := 0
	if solved {
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
	}

	// Upsert user question
	userQuestion, err := s.userQuestionRepo.Upsert(
		userID,
		questionID,
		solved,
		req.TimeTaken,
		question.DifficultyLevel,
		question.QuestionType,
		xpGained,
	)
	if err != nil {
		return nil, err
	}

	// Update user stats and log activity
	if solved {
		// Increment total_solved, total_xp, and total_attempts
		if err := s.userRepo.IncrementSolved(userID, xpGained); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to increment user stats: %v\n", err)
		}
	} else {
		// Just increment attempts for wrong answers
		if err := s.userRepo.IncrementAttempts(userID); err != nil {
			fmt.Printf("Failed to increment user attempts: %v\n", err)
		}
	}

	// Log question activity (works for both attempted and solved)
	questionText := question.QuestionText
	if len(questionText) > 150 {
		questionText = questionText[:150] + "..."
	}

	activityMeta := models.QuestionActivityMetadata{
		QuestionID:      questionID,
		QuestionText:    questionText,
		Solved:          solved,
		DifficultyLevel: question.DifficultyLevel,
		QuestionType:    question.QuestionType,
		TimeTaken:       req.TimeTaken,
		XPGained:        xpGained,
	}
	if err := s.activityService.LogQuestionActivity(userID, activityMeta); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to log question activity: %v\n", err)
	}

	return userQuestion, nil
}
