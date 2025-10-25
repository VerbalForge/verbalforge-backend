package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

// SupportTicketRepository handles support ticket database operations
type SupportTicketRepository struct {
	collection *mongo.Collection
}

// NewSupportTicketRepository creates a new support ticket repository
func NewSupportTicketRepository(db *mongo.Database) *SupportTicketRepository {
	return &SupportTicketRepository{
		collection: db.Collection("support_tickets"),
	}
}

// Create creates a new support ticket
func (r *SupportTicketRepository) Create(ctx context.Context, ticket *models.SupportTicket) error {
	_, err := r.collection.InsertOne(ctx, ticket)
	return err
}

// GetAll retrieves all support tickets with optional status filter
func (r *SupportTicketRepository) GetAll(ctx context.Context, status string) ([]*models.SupportTicket, error) {
	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tickets []*models.SupportTicket
	if err = cursor.All(ctx, &tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

// GetByUserID retrieves all support tickets for a specific user
func (r *SupportTicketRepository) GetByUserID(ctx context.Context, userID string) ([]*models.SupportTicket, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tickets []*models.SupportTicket
	if err = cursor.All(ctx, &tickets); err != nil {
		return nil, err
	}

	return tickets, nil
}

// GetByID retrieves a support ticket by ID
func (r *SupportTicketRepository) GetByID(ctx context.Context, id string) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&ticket)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// Update updates a support ticket
func (r *SupportTicketRepository) Update(ctx context.Context, id string, status string, priority string) error {
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if status != "" {
		update["$set"].(bson.M)["status"] = status
	}
	if priority != "" {
		update["$set"].(bson.M)["priority"] = priority
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete deletes a support ticket by ID
func (r *SupportTicketRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
