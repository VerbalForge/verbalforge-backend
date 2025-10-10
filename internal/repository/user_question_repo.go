package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
)

// UserQuestionRepository handles user question attempt database operations
type UserQuestionRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

// NewUserQuestionRepository creates a new user question repository
func NewUserQuestionRepository(db *mongo.Database) *UserQuestionRepository {
	return &UserQuestionRepository{
		collection: db.Collection("user_question"),
		db:         db,
	}
}

// Upsert creates or updates a user's question attempt
func (r *UserQuestionRepository) Upsert(userID, questionID string, solved bool, timeTaken int, difficulty, questionType string, xpGained int) (*models.UserQuestion, error) {
	filter := bson.M{
		"user_id":     userID,
		"question_id": questionID,
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"solved":          solved,
			"attempted":       true,
			"last_attempt_at": now,
			"time_taken":      timeTaken,
			"difficulty":      difficulty,
			"question_type":   questionType,
			"xp_gained":       xpGained,
			"updated_at":      now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}

	// Only update last_solved_at if solved is true
	if solved {
		update["$set"].(bson.M)["last_solved_at"] = now
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return nil, err
	}

	// Retrieve and return the updated document
	var userQuestion models.UserQuestion
	err = r.collection.FindOne(context.Background(), filter).Decode(&userQuestion)
	if err != nil {
		return nil, err
	}

	return &userQuestion, nil
}

// FindByUserAndQuestion retrieves a user's progress on a specific question
func (r *UserQuestionRepository) FindByUserAndQuestion(userID, questionID string) (*models.UserQuestion, error) {
	filter := bson.M{
		"user_id":     userID,
		"question_id": questionID,
	}

	var userQuestion models.UserQuestion
	err := r.collection.FindOne(context.Background(), filter).Decode(&userQuestion)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userQuestion, nil
}

// FindByUserAndQuestions retrieves user's progress for multiple questions
func (r *UserQuestionRepository) FindByUserAndQuestions(userID string, questionIDs []string) (map[string]*models.UserQuestion, error) {
	filter := bson.M{
		"user_id":     userID,
		"question_id": bson.M{"$in": questionIDs},
	}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	result := make(map[string]*models.UserQuestion)
	for cursor.Next(context.Background()) {
		var userQuestion models.UserQuestion
		if err := cursor.Decode(&userQuestion); err != nil {
			continue
		}
		result[userQuestion.QuestionID] = &userQuestion
	}

	return result, nil
}

// GetRecentSolved retrieves the most recent solved questions for a user
func (r *UserQuestionRepository) GetRecentSolved(userID string, limit int) ([]models.UserQuestion, error) {
	filter := bson.M{
		"user_id": userID,
		"solved":  true,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "last_solved_at", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var questions []models.UserQuestion
	if err = cursor.All(context.Background(), &questions); err != nil {
		return nil, err
	}

	return questions, nil
}

// CalculateUserStats aggregates stats from user_questions collection
func (r *UserQuestionRepository) CalculateUserStats(userID string) (*models.UserStats, error) {
	pipeline := mongo.Pipeline{
		// Match all solved questions for this user
		{{Key: "$match", Value: bson.M{
			"user_id": userID,
			"solved":  true,
		}}},
		// Group and calculate stats
		{{Key: "$group", Value: bson.M{
			"_id":          "$user_id",
			"total_solved": bson.M{"$sum": 1},
			"total_xp":     bson.M{"$sum": "$xp_gained"},
			"easy_solved": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$difficulty", "easy"}},
						1,
						0,
					},
				},
			},
			"medium_solved": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$difficulty", "medium"}},
						1,
						0,
					},
				},
			},
			"hard_solved": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$difficulty", "hard"}},
						1,
						0,
					},
				},
			},
			"tc_solved": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$or": []interface{}{
							bson.M{"$eq": []interface{}{"$question_type", "text_completion_single"}},
							bson.M{"$eq": []interface{}{"$question_type", "text_completion_double"}},
							bson.M{"$eq": []interface{}{"$question_type", "text_completion_triple"}},
						}},
						1,
						0,
					},
				},
			},
			"se_solved": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$eq": []interface{}{"$question_type", "sentence_equivalence"}},
						1,
						0,
					},
				},
			},
			"rc_solved": bson.M{
				"$sum": bson.M{
					"$cond": []interface{}{
						bson.M{"$or": []interface{}{
							bson.M{"$eq": []interface{}{"$question_type", "reading_comprehension"}},
							bson.M{"$eq": []interface{}{"$question_type", "reading_comprehension_single"}},
							bson.M{"$eq": []interface{}{"$question_type", "reading_comprehension_multiple"}},
							bson.M{"$eq": []interface{}{"$question_type", "reading_comprehension_highlight"}},
						}},
						1,
						0,
					},
				},
			},
		}}},
	}

	cursor, err := r.collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		return nil, err
	}

	// If no results, return zero stats
	if len(results) == 0 {
		return &models.UserStats{
			UserID:        userID,
			TotalSolved:   0,
			EasySolved:    0,
			MediumSolved:  0,
			HardSolved:    0,
			TCSolved:      0,
			SESolved:      0,
			RCSolved:      0,
			TotalXP:       0,
			Rank:          0,
			CurrentStreak: 0,
			LongestStreak: 0,
			LastLogin:     time.Time{},
			TotalAttempts: 0,
			ProfileViews:  0,
		}, nil
	}

	result := results[0]

	stats := &models.UserStats{
		UserID:       userID,
		TotalSolved:  getInt(result, "total_solved"),
		EasySolved:   getInt(result, "easy_solved"),
		MediumSolved: getInt(result, "medium_solved"),
		HardSolved:   getInt(result, "hard_solved"),
		TCSolved:     getInt(result, "tc_solved"),
		SESolved:     getInt(result, "se_solved"),
		RCSolved:     getInt(result, "rc_solved"),
		TotalXP:      getInt(result, "total_xp"),
	}

	return stats, nil
}

// Helper function to safely get int from bson.M
func getInt(m bson.M, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}
