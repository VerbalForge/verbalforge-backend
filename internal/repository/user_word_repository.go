package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

type UserWordRepository struct {
	collection *mongo.Collection
}

func NewUserWordRepository(db *mongo.Database) *UserWordRepository {
	return &UserWordRepository{
		collection: db.Collection("user_words"),
	}
}

func (r *UserWordRepository) GetUserWords(ctx context.Context, userID string) (*models.UserWord, error) {
	var userWords models.UserWord

	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&userWords)
	if err == mongo.ErrNoDocuments {
		// Create new user word document if not exists
		userWords = models.UserWord{
			ID:        uuid.New().String(),
			UserID:    userID,
			Known:     []string{},
			Practice:  []string{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = r.collection.InsertOne(ctx, userWords)
		if err != nil {
			return nil, err
		}
		return &userWords, nil
	}

	return &userWords, err
}

func (r *UserWordRepository) UpdateUserWords(ctx context.Context, userWords *models.UserWord) error {
	userWords.UpdatedAt = time.Now()

	_, err := r.collection.ReplaceOne(
		ctx,
		bson.M{"user_id": userWords.UserID},
		userWords,
		options.Replace().SetUpsert(true),
	)

	return err
}

func (r *UserWordRepository) AddToKnown(ctx context.Context, userID, wordID string) error {
	userWords, err := r.GetUserWords(ctx, userID)
	if err != nil {
		return err
	}

	// Remove from practice if exists
	userWords.Practice = removeFromSlice(userWords.Practice, wordID)

	// Add to known if not already there
	if !contains(userWords.Known, wordID) {
		userWords.Known = append(userWords.Known, wordID)
	}

	return r.UpdateUserWords(ctx, userWords)
}

func (r *UserWordRepository) AddToPractice(ctx context.Context, userID, wordID string) error {
	userWords, err := r.GetUserWords(ctx, userID)
	if err != nil {
		return err
	}

	// Remove from known if exists
	userWords.Known = removeFromSlice(userWords.Known, wordID)

	// Add to practice if not already there
	if !contains(userWords.Practice, wordID) {
		userWords.Practice = append(userWords.Practice, wordID)
	}

	return r.UpdateUserWords(ctx, userWords)
}

func (r *UserWordRepository) RemoveWordStatus(ctx context.Context, userID, wordID string) error {
	userWords, err := r.GetUserWords(ctx, userID)
	if err != nil {
		return err
	}

	userWords.Known = removeFromSlice(userWords.Known, wordID)
	userWords.Practice = removeFromSlice(userWords.Practice, wordID)

	return r.UpdateUserWords(ctx, userWords)
}

func (r *UserWordRepository) ResetUserProgress(ctx context.Context, userID string) error {
	userWords, err := r.GetUserWords(ctx, userID)
	if err != nil {
		return err
	}

	userWords.Known = []string{}
	userWords.Practice = []string{}

	return r.UpdateUserWords(ctx, userWords)
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
