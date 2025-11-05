package repository

import (
	"context"
	"log"

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

	filter := bson.M{"_id": id}
	log.Printf("DEBUG: Looking for passage with filter: %+v", filter)

	err := r.collection.FindOne(ctx, filter).Decode(&passage)
	if err != nil {
		log.Printf("DEBUG: FindByID error for id '%s': %v", id, err)
		return nil, err
	}

	log.Printf("DEBUG: Found passage: %s", passage.ID)
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

// Update updates a passage
func (r *PassageRepository) Update(id string, passage *models.Passage) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}
	update := bson.M{"$set": passage}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete deletes a passage by ID
func (r *PassageRepository) Delete(id string) error {
	ctx := context.Background()
	filter := bson.M{"_id": id}

	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}
