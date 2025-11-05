package repository

import (
	"context"
	"fmt"
	"time"

	"verbalforge-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type WordRepository struct {
	collection *mongo.Collection
}

func NewWordRepository(db *mongo.Database) *WordRepository {
	return &WordRepository{
		collection: db.Collection("words"),
	}
}

// GetWords retrieves words with filters and pagination
func (r *WordRepository) GetWords(ctx context.Context, filters models.WordFilters, userWords *models.UserWord) (*models.WordsResponse, error) {
	// Build filter query
	filter := bson.M{}

	// Source filter
	if len(filters.Sources) > 0 {
		filter["sources"] = bson.M{"$in": filters.Sources}
	}

	// Search filter (case-insensitive word search)
	if filters.Search != "" {
		filter["word"] = bson.M{"$regex": primitive.Regex{
			Pattern: filters.Search,
			Options: "i",
		}}
	}

	// Progress filter based on user's known/practice words
	if filters.Known != nil {
		if *filters.Known {
			// Show only known words
			if userWords != nil && len(userWords.Known) > 0 {
				filter["_id"] = bson.M{"$in": userWords.Known}
			} else {
				// No known words, return empty result
				filter["_id"] = bson.M{"$in": []string{}}
			}
		} else {
			// Exclude known words
			if userWords != nil && len(userWords.Known) > 0 {
				filter["_id"] = bson.M{"$nin": userWords.Known}
			}
		}
	}

	if filters.Practice != nil {
		if *filters.Practice {
			// Show only practice words
			if userWords != nil && len(userWords.Practice) > 0 {
				if existingFilter, exists := filter["_id"]; exists {
					// Combine with existing filter (intersection)
					if knownFilter, ok := existingFilter.(bson.M); ok {
						if inFilter, ok := knownFilter["$in"]; ok {
							filter["_id"] = bson.M{"$in": intersectSlices(inFilter.([]string), userWords.Practice)}
						}
					}
				} else {
					filter["_id"] = bson.M{"$in": userWords.Practice}
				}
			} else {
				// No practice words, return empty result
				filter["_id"] = bson.M{"$in": []string{}}
			}
		} else {
			// Exclude practice words
			if userWords != nil && len(userWords.Practice) > 0 {
				if existingFilter, exists := filter["_id"]; exists {
					if ninFilter, ok := existingFilter.(bson.M); ok {
						if ninSlice, ok := ninFilter["$nin"]; ok {
							// Combine nin filters
							combined := append(ninSlice.([]string), userWords.Practice...)
							filter["_id"] = bson.M{"$nin": combined}
						}
					}
				} else {
					filter["_id"] = bson.M{"$nin": userWords.Practice}
				}
			}
		}
	}

	// Set default pagination
	if filters.Limit == 0 {
		filters.Limit = 50
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count words: %w", err)
	}

	// Find options with pagination and sorting
	opts := options.Find().
		SetSkip(int64(filters.Offset)).
		SetLimit(int64(filters.Limit)).
		SetSort(bson.D{{Key: "word", Value: 1}}) // Sort alphabetically

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find words: %w", err)
	}
	defer cursor.Close(ctx)

	var words []models.Word
	if err = cursor.All(ctx, &words); err != nil {
		return nil, fmt.Errorf("failed to decode words: %w", err)
	}

	// Get available sources for filtering
	sources, err := r.GetAvailableSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available sources: %w", err)
	}

	hasMore := int64(filters.Offset+len(words)) < total

	return &models.WordsResponse{
		Words:   words,
		Total:   total,
		Sources: sources,
		HasMore: hasMore,
		Offset:  filters.Offset,
		Limit:   filters.Limit,
	}, nil
}

// GetWordByID retrieves a single word by its ID
func (r *WordRepository) GetWordByID(ctx context.Context, id string) (*models.Word, error) {
	var word models.Word
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&word)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("word not found")
		}
		return nil, fmt.Errorf("failed to find word: %w", err)
	}
	return &word, nil
}

// GetAvailableSources returns all unique sources in the collection
func (r *WordRepository) GetAvailableSources(ctx context.Context) ([]string, error) {
	sources, err := r.collection.Distinct(ctx, "sources", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct sources: %w", err)
	}

	var result []string
	for _, source := range sources {
		if str, ok := source.(string); ok {
			result = append(result, str)
		}
	}

	return result, nil
}

// GetWordsByIDs retrieves multiple words by their IDs (for navigation)
func (r *WordRepository) GetWordsByIDs(ctx context.Context, ids []string) ([]models.Word, error) {
	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find words: %w", err)
	}
	defer cursor.Close(ctx)

	var words []models.Word
	if err = cursor.All(ctx, &words); err != nil {
		return nil, fmt.Errorf("failed to decode words: %w", err)
	}

	return words, nil
}

// CreateWord creates a new word (for future admin functionality)
func (r *WordRepository) CreateWord(ctx context.Context, word *models.Word) error {
	word.CreatedAt = time.Now()
	word.UpdatedAt = time.Now()

	if word.ID == "" {
		word.ID = primitive.NewObjectID().Hex()
	}

	_, err := r.collection.InsertOne(ctx, word)
	if err != nil {
		return fmt.Errorf("failed to create word: %w", err)
	}

	return nil
}

// UpdateWord updates an existing word
func (r *WordRepository) UpdateWord(ctx context.Context, id string, word *models.Word) error {
	word.UpdatedAt = time.Now()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": word}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update word: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("word not found")
	}

	return nil
}

// DeleteWord deletes a word by ID
func (r *WordRepository) DeleteWord(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete word: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("word not found")
	}

	return nil
}

// Helper function to find intersection of two string slices
func intersectSlices(a, b []string) []string {
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}

	var result []string
	for _, item := range b {
		if m[item] {
			result = append(result, item)
		}
	}

	return result
}
