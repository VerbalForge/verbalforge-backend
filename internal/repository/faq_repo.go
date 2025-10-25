package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

// FAQRepository handles FAQ database operations
type FAQRepository struct {
	collection *mongo.Collection
}

// NewFAQRepository creates a new FAQ repository
func NewFAQRepository(db *mongo.Database) *FAQRepository {
	return &FAQRepository{
		collection: db.Collection("faqs"),
	}
}

// Create creates a new FAQ
func (r *FAQRepository) Create(ctx context.Context, faq *models.FAQ) error {
	_, err := r.collection.InsertOne(ctx, faq)
	return err
}

// GetAll retrieves all FAQs with optional status filter
func (r *FAQRepository) GetAll(ctx context.Context, status string) ([]*models.FAQ, error) {
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

	var faqs []*models.FAQ
	if err = cursor.All(ctx, &faqs); err != nil {
		return nil, err
	}

	return faqs, nil
}

// GetByID retrieves an FAQ by ID
func (r *FAQRepository) GetByID(ctx context.Context, id string) (*models.FAQ, error) {
	var faq models.FAQ
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&faq)
	if err != nil {
		return nil, err
	}
	return &faq, nil
}

// UpdateStatus updates the status and optionally answer of an FAQ
func (r *FAQRepository) UpdateStatus(ctx context.Context, id string, status string, answer string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if answer != "" {
		update["$set"].(bson.M)["answer"] = answer
	}

	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

// Delete deletes an FAQ by ID
func (r *FAQRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
