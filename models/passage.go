package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PassageMetadata struct {
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	PublishedAt time.Time `json:"published_at" bson:"published_at"`
	BatchID     string    `json:"batch_id" bson:"batch_id"`
}

type Passage struct {
	ID          string          `json:"id" bson:"_id,omitempty"`
	Passage     string          `json:"passage" bson:"passage"`
	Source      string          `json:"source" bson:"source"`
	Title       string          `json:"title" bson:"title"`
	Difficulty  string          `json:"difficulty" bson:"difficulty"`
	Type        string          `json:"type" bson:"type"`
	QuestionIDs []string        `json:"question_ids" bson:"question_ids"`
	Metadata    PassageMetadata `json:"metadata" bson:"metadata"`
}

// PartialPassage contains minimal passage information for listing
type PartialPassage struct {
	ID          string   `json:"id" bson:"_id,omitempty"`
	Passage     string   `json:"passage" bson:"passage"`
	Title       string   `json:"title" bson:"title"`
	Difficulty  string   `json:"difficulty" bson:"difficulty"`
	QuestionIDs []string `json:"question_ids" bson:"question_ids"`
	CreatedAt   string   `json:"created_at" bson:"metadata.created_at"`
}

type PassageModel struct {
	collection *mongo.Collection
}

func NewPassageModel(db *mongo.Database) *PassageModel {
	return &PassageModel{
		collection: db.Collection("passages"),
	}
}

// GetPassageByID returns a single passage by ID
func (p *PassageModel) GetPassageByID(id string) (*Passage, error) {
	ctx := context.Background()
	var passage Passage

	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&passage)
	if err != nil {
		return nil, err
	}

	return &passage, nil
}

// GetPassages returns a list of passages with optional filters
func (p *PassageModel) GetPassages(filter bson.M, limit int64, skip int64) ([]Passage, error) {
	ctx := context.Background()

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	opts.SetSort(bson.D{{Key: "metadata.created_at", Value: -1}})

	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var passages []Passage
	if err = cursor.All(ctx, &passages); err != nil {
		return nil, err
	}

	return passages, nil
}

// GetPartialPassages returns a list of partial passages with only essential fields
func (p *PassageModel) GetPartialPassages(filter bson.M, limit int64, skip int64) ([]PartialPassage, error) {
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
		"passage":             1,
		"title":               1,
		"difficulty":          1,
		"question_ids":        1,
		"metadata.created_at": 1,
	})

	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var passages []PartialPassage
	if err = cursor.All(ctx, &passages); err != nil {
		return nil, err
	}

	return passages, nil
}

// CountPassages returns the total count of passages matching the filter
func (p *PassageModel) CountPassages(filter bson.M) (int64, error) {
	ctx := context.Background()
	count, err := p.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}
