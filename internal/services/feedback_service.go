package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// FeedbackService handles feedback business logic
type FeedbackService struct {
	repo *repository.FeedbackRepository
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(repo *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{
		repo: repo,
	}
}

// CreateFeedback creates a new feedback submission
func (s *FeedbackService) CreateFeedback(ctx context.Context, req *models.CreateFeedbackRequest) (*models.Feedback, error) {
	feedback := &models.Feedback{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Type:      req.Type,
		Message:   req.Message,
		Status:    "new",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, feedback); err != nil {
		return nil, err
	}

	return feedback, nil
}

// GetAllFeedback retrieves all feedback with optional filters
func (s *FeedbackService) GetAllFeedback(ctx context.Context, feedbackType string, status string) ([]*models.Feedback, error) {
	return s.repo.GetAll(ctx, feedbackType, status)
}

// GetFeedbackByID retrieves feedback by ID
func (s *FeedbackService) GetFeedbackByID(ctx context.Context, id string) (*models.Feedback, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateFeedbackStatus updates the status of feedback
func (s *FeedbackService) UpdateFeedbackStatus(ctx context.Context, id string, req *models.UpdateFeedbackStatusRequest) error {
	return s.repo.UpdateStatus(ctx, id, req.Status)
}

// DeleteFeedback deletes feedback by ID
func (s *FeedbackService) DeleteFeedback(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
