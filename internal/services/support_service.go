package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// SupportTicketService handles support ticket business logic
type SupportTicketService struct {
	repo *repository.SupportTicketRepository
}

// NewSupportTicketService creates a new support ticket service
func NewSupportTicketService(repo *repository.SupportTicketRepository) *SupportTicketService {
	return &SupportTicketService{
		repo: repo,
	}
}

// CreateSupportTicket creates a new support ticket
func (s *SupportTicketService) CreateSupportTicket(ctx context.Context, userID string, req *models.CreateSupportTicketRequest) (*models.SupportTicket, error) {
	ticket := &models.SupportTicket{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Email:     req.Email,
		Subject:   req.Subject,
		Message:   req.Message,
		Status:    "open",
		Priority:  "medium",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

// GetAllSupportTickets retrieves all support tickets with optional status filter
func (s *SupportTicketService) GetAllSupportTickets(ctx context.Context, status string) ([]*models.SupportTicket, error) {
	return s.repo.GetAll(ctx, status)
}

// GetUserSupportTickets retrieves all support tickets for a specific user
func (s *SupportTicketService) GetUserSupportTickets(ctx context.Context, userID string) ([]*models.SupportTicket, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetSupportTicketByID retrieves a support ticket by ID
func (s *SupportTicketService) GetSupportTicketByID(ctx context.Context, id string) (*models.SupportTicket, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateSupportTicket updates a support ticket
func (s *SupportTicketService) UpdateSupportTicket(ctx context.Context, id string, req *models.UpdateSupportTicketRequest) error {
	return s.repo.Update(ctx, id, req.Status, req.Priority)
}

// DeleteSupportTicket deletes a support ticket by ID
func (s *SupportTicketService) DeleteSupportTicket(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
