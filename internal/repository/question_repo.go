package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

// QuestionRepository handles question database operations
type QuestionRepository struct {
	collection *mongo.Collection
}

// NewQuestionRepository creates a new question repository
func NewQuestionRepository(db *mongo.Database) *QuestionRepository {
	return &QuestionRepository{
		collection: db.Collection("questions"),
	}
}

// FindByID finds a question by ID
func (r *QuestionRepository) FindByID(id string) (*models.Question, error) {
	ctx := context.Background()
	var question models.Question

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&question)
	if err != nil {
		return nil, err
	}

	return &question, nil
}

// FindAll returns a list of questions with optional filters
func (r *QuestionRepository) FindAll(filter bson.M, limit int64, skip int64) ([]models.Question, error) {
	ctx := context.Background()

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	opts.SetSort(bson.D{{Key: "metadata.created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// FindPartial returns a list of partial questions with only essential fields
func (r *QuestionRepository) FindPartial(filter bson.M, limit int64, skip int64) ([]models.PartialQuestion, error) {
	ctx := context.Background()

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	opts.SetSort(bson.D{{Key: "metadata.created_at", Value: -1}})

	// Only select the fields we need
	opts.SetProjection(bson.M{
		"_id":                 1,
		"question_type":       1,
		"difficulty_level":    1,
		"topic":               1,
		"question_text":       1,
		"metadata.created_at": 1,
	})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.PartialQuestion
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// Count returns the total count of questions matching the filter
func (r *QuestionRepository) Count(filter bson.M) (int64, error) {
	ctx := context.Background()
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}
