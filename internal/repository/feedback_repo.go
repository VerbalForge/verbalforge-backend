package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

// FeedbackRepository handles feedback database operations
type FeedbackRepository struct {
	collection *mongo.Collection
}

// NewFeedbackRepository creates a new feedback repository
func NewFeedbackRepository(db *mongo.Database) *FeedbackRepository {
	return &FeedbackRepository{
		collection: db.Collection("feedback"),
	}
}

// Create creates a new feedback
func (r *FeedbackRepository) Create(ctx context.Context, feedback *models.Feedback) error {
	_, err := r.collection.InsertOne(ctx, feedback)
	return err
}

// GetAll retrieves all feedback with optional filters
func (r *FeedbackRepository) GetAll(ctx context.Context, feedbackType string, status string) ([]*models.Feedback, error) {
	filter := bson.M{}
	if feedbackType != "" {
		filter["type"] = feedbackType
	}
	if status != "" {
		filter["status"] = status
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var feedbacks []*models.Feedback
	if err = cursor.All(ctx, &feedbacks); err != nil {
		return nil, err
	}

	return feedbacks, nil
}

// GetByID retrieves feedback by ID
func (r *FeedbackRepository) GetByID(ctx context.Context, id string) (*models.Feedback, error) {
	var feedback models.Feedback
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&feedback)
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

// UpdateStatus updates the status of feedback
func (r *FeedbackRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete deletes feedback by ID
func (r *FeedbackRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
