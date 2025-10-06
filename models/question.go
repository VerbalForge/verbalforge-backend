package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Choice struct {
	Option    string `json:"option" bson:"option"`
	Blank     int    `json:"blank" bson:"blank"`
	IsCorrect bool   `json:"is_correct" bson:"is_correct"`
	Reasoning string `json:"reasoning" bson:"reasoning"`
}

type QuestionMetadata struct {
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	PublishedAt time.Time `json:"published_at" bson:"published_at"`
	BatchID     string    `json:"batch_id" bson:"batch_id"`
}

type Question struct {
	ID              string           `json:"id" bson:"_id,omitempty"`
	QuestionType    string           `json:"question_type" bson:"question_type"`
	DifficultyLevel string           `json:"difficulty_level" bson:"difficulty_level"`
	Topic           string           `json:"topic" bson:"topic"`
	QuestionText    string           `json:"question_text" bson:"question_text"`
	PassageID       string           `json:"passage_id,omitempty" bson:"passage_id,omitempty"` // Optional, only for RC questions
	Choices         []Choice         `json:"choices" bson:"choices"`
	Metadata        QuestionMetadata `json:"metadata" bson:"metadata"`
}

// PartialQuestion contains minimal question information for listing
type PartialQuestion struct {
	ID              string `json:"id" bson:"_id,omitempty"`
	QuestionType    string `json:"question_type" bson:"question_type"`
	DifficultyLevel string `json:"difficulty_level" bson:"difficulty_level"`
	Topic           string `json:"topic" bson:"topic"`
	QuestionText    string `json:"question_text" bson:"question_text"`
	CreatedAt       string `json:"created_at" bson:"metadata.created_at"`
}

type QuestionModel struct {
	collection *mongo.Collection
}

func NewQuestionModel(db *mongo.Database) *QuestionModel {
	return &QuestionModel{
		collection: db.Collection("questions"),
	}
}

// GetQuestions returns a list of questions with optional filters
func (q *QuestionModel) GetQuestions(filter bson.M, limit int64, skip int64) ([]Question, error) {
	ctx := context.Background()

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	opts.SetSort(bson.D{{Key: "metadata.created_at", Value: -1}})

	cursor, err := q.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []Question
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// GetPartialQuestions returns a list of partial questions with only essential fields
func (q *QuestionModel) GetPartialQuestions(filter bson.M, limit int64, skip int64) ([]PartialQuestion, error) {
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

	cursor, err := q.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []PartialQuestion
	if err = cursor.All(ctx, &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// GetQuestionByID returns a single question by ID
func (q *QuestionModel) GetQuestionByID(id string) (*Question, error) {
	ctx := context.Background()
	var question Question

	err := q.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&question)
	if err != nil {
		return nil, err
	}

	return &question, nil
}

// CountQuestions returns the total count of questions matching the filter
func (q *QuestionModel) CountQuestions(filter bson.M) (int64, error) {
	ctx := context.Background()
	count, err := q.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}
