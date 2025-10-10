package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

// PassageRepository handles passage database operations
type PassageRepository struct {
	collection *mongo.Collection
}

// NewPassageRepository creates a new passage repository
func NewPassageRepository(db *mongo.Database) *PassageRepository {
	return &PassageRepository{
		collection: db.Collection("passages"),
	}
}

// FindByID finds a passage by ID
func (r *PassageRepository) FindByID(id string) (*models.Passage, error) {
	ctx := context.Background()
	var passage models.Passage

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&passage)
	if err != nil {
		return nil, err
	}

	return &passage, nil
}

// FindAll returns a list of passages with optional filters
func (r *PassageRepository) FindAll(filter bson.M, limit int64, skip int64) ([]models.Passage, error) {
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

	var passages []models.Passage
	if err = cursor.All(ctx, &passages); err != nil {
		return nil, err
	}

	return passages, nil
}

// FindPartial returns a list of partial passages with only essential fields
func (r *PassageRepository) FindPartial(filter bson.M, limit int64, skip int64) ([]models.PartialPassage, error) {
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

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var passages []models.PartialPassage
	if err = cursor.All(ctx, &passages); err != nil {
		return nil, err
	}

	return passages, nil
}

// Count returns the total count of passages matching the filter
func (r *PassageRepository) Count(filter bson.M) (int64, error) {
	ctx := context.Background()
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindPartialWithCursor returns partial passages using cursor-based pagination
func (r *PassageRepository) FindPartialWithCursor(filter bson.M, limit int64, lastID, lastCreatedAt string) ([]models.PartialPassage, error) {
	ctx := context.Background()

	// Build cursor filter
	cursorFilter := bson.M{}
	for k, v := range filter {
		cursorFilter[k] = v
	}

	// Add cursor conditions for pagination
	if lastID != "" && lastCreatedAt != "" {
		cursorFilter["$or"] = []bson.M{
			{"metadata.created_at": bson.M{"$lt": lastCreatedAt}},
			{
				"metadata.created_at": lastCreatedAt,
				"_id":                 bson.M{"$gt": lastID},
			},
		}
	}

	// Use aggregation pipeline to flatten metadata.created_at
	pipeline := []bson.M{
		{"$match": cursorFilter},
		{"$sort": bson.D{
			{Key: "metadata.created_at", Value: -1},
			{Key: "_id", Value: 1},
		}},
	}

	if limit > 0 {
		pipeline = append(pipeline, bson.M{"$limit": limit + 1})
	}

	// Project and flatten the created_at field
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"_id":          "$_id",
			"passage":      1,
			"title":        1,
			"difficulty":   1,
			"question_ids": 1,
			"created_at":   "$metadata.created_at",
		},
	})

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var passages []models.PartialPassage
	if err = cursor.All(ctx, &passages); err != nil {
		return nil, err
	}

	return passages, nil
}
