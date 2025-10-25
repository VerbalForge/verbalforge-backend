package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// FAQService handles FAQ business logic
type FAQService struct {
	repo *repository.FAQRepository
}

// NewFAQService creates a new FAQ service
func NewFAQService(repo *repository.FAQRepository) *FAQService {
	return &FAQService{
		repo: repo,
	}
}

// CreateFAQ creates a new FAQ submission
func (s *FAQService) CreateFAQ(ctx context.Context, req *models.CreateFAQRequest) (*models.FAQ, error) {
	faq := &models.FAQ{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Question:  req.Question,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, faq); err != nil {
		return nil, err
	}

	return faq, nil
}

// GetAllFAQs retrieves all FAQs with optional status filter
func (s *FAQService) GetAllFAQs(ctx context.Context, status string) ([]*models.FAQ, error) {
	return s.repo.GetAll(ctx, status)
}

// GetFAQByID retrieves an FAQ by ID
func (s *FAQService) GetFAQByID(ctx context.Context, id string) (*models.FAQ, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateFAQStatus updates the status and optionally answer of an FAQ
func (s *FAQService) UpdateFAQStatus(ctx context.Context, id string, req *models.UpdateFAQStatusRequest) error {
	return s.repo.UpdateStatus(ctx, id, req.Status, req.Answer)
}

// DeleteFAQ deletes an FAQ by ID
func (s *FAQService) DeleteFAQ(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
