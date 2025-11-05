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

// UserActivitySummaryRepository handles user activity summary database operations
type UserActivitySummaryRepository struct {
	collection *mongo.Collection
}

// NewUserActivitySummaryRepository creates a new user activity summary repository
func NewUserActivitySummaryRepository(db *mongo.Database) *UserActivitySummaryRepository {
	return &UserActivitySummaryRepository{
		collection: db.Collection("user_activity_summary"),
	}
}

// UpsertDailyActivity updates or adds a daily activity entry in the user's summary
func (r *UserActivitySummaryRepository) UpsertDailyActivity(userID, date string, questionSolved bool, totalXP int, questionID string) error {
	ctx := context.Background()
	now := time.Now()

	// Parse the date string to get start of day timestamp
	dayStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}

	filter := bson.M{
		"user_id": userID,
	}

	// Check if activity for this date already exists (match by day timestamp)
	arrayFilter := bson.M{
		"user_id": userID,
		"daily_activities.timestamp": bson.M{
			"$gte": dayStart,
			"$lt":  dayStart.AddDate(0, 0, 1),
		},
	}

	var existing models.UserActivitySummary
	err = r.collection.FindOne(ctx, arrayFilter).Decode(&existing)

	if err == mongo.ErrNoDocuments || len(existing.DailyActivities) == 0 {
		// Date doesn't exist, push new daily activity
		questionsSolvedValue := 0
		questionIDsValue := []string{}
		if questionSolved {
			questionsSolvedValue = 1
			if questionID != "" {
				questionIDsValue = []string{questionID}
			}
		}

		update := bson.M{
			"$push": bson.M{
				"daily_activities": bson.M{
					"timestamp":           dayStart,
					"questions_solved":    questionsSolvedValue,
					"questions_attempted": 1,
					"question_ids":        questionIDsValue,
					"total_xp":            totalXP,
					"total_activities":    1,
					"created_at":          now,
					"updated_at":          now,
				},
			},
			"$setOnInsert": bson.M{
				"_id":        uuid.New().String(),
				"user_id":    userID,
				"created_at": now,
			},
			"$set": bson.M{
				"updated_at": now,
			},
		}

		opts := options.Update().SetUpsert(true)
		_, err = r.collection.UpdateOne(ctx, filter, update, opts)
		return err
	}

	// Date exists, update the existing daily activity
	update := bson.M{
		"$inc": bson.M{
			"daily_activities.$[elem].total_activities":    1,
			"daily_activities.$[elem].questions_attempted": 1,
			"daily_activities.$[elem].total_xp":            totalXP,
		},
		"$set": bson.M{
			"updated_at":                          now,
			"daily_activities.$[elem].updated_at": now,
		},
	}

	if questionSolved && questionID != "" {
		update["$inc"].(bson.M)["daily_activities.$[elem].questions_solved"] = 1
		update["$addToSet"] = bson.M{
			"daily_activities.$[elem].question_ids": questionID,
		}
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.timestamp": bson.M{
					"$gte": dayStart,
					"$lt":  dayStart.AddDate(0, 0, 1),
				},
			},
		},
	})

	_, err = r.collection.UpdateOne(ctx, filter, update, arrayFilters)
	return err
}

// UpsertWordActivity updates or adds word activity data to the daily activity entry
func (r *UserActivitySummaryRepository) UpsertWordActivity(userID, date string, actionType models.WordActionType, wordID string) error {
	ctx := context.Background()
	now := time.Now()

	// Parse the date string to get start of day timestamp
	dayStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		return err
	}

	filter := bson.M{
		"user_id": userID,
	}

	// Check if activity for this date already exists
	arrayFilter := bson.M{
		"user_id": userID,
		"daily_activities.timestamp": bson.M{
			"$gte": dayStart,
			"$lt":  dayStart.AddDate(0, 0, 1),
		},
	}

	var existing models.UserActivitySummary
	err = r.collection.FindOne(ctx, arrayFilter).Decode(&existing)

	// Determine which field to increment based on action type
	var fieldToIncrement string
	switch actionType {
	case models.ActionWordMarkedKnown:
		fieldToIncrement = "words_marked_known"
	case models.ActionWordMarkedPractice:
		fieldToIncrement = "words_marked_practice"
	case models.ActionWordViewed:
		fieldToIncrement = "words_viewed"
	default:
		// For unmarked actions, don't increment counters
		fieldToIncrement = ""
	}

	if err == mongo.ErrNoDocuments || len(existing.DailyActivities) == 0 {
		// Date doesn't exist, push new daily activity
		newActivity := bson.M{
			"timestamp":             dayStart,
			"questions_solved":      0,
			"questions_attempted":   0,
			"question_ids":          []string{},
			"total_xp":              0,
			"total_activities":      1,
			"words_marked_known":    0,
			"words_marked_practice": 0,
			"words_viewed":          0,
			"word_ids":              []string{},
			"created_at":            now,
			"updated_at":            now,
		}

		if fieldToIncrement != "" {
			newActivity[fieldToIncrement] = 1
		}
		if wordID != "" {
			newActivity["word_ids"] = []string{wordID}
		}

		update := bson.M{
			"$push": bson.M{
				"daily_activities": newActivity,
			},
			"$setOnInsert": bson.M{
				"_id":        uuid.New().String(),
				"user_id":    userID,
				"created_at": now,
			},
			"$set": bson.M{
				"updated_at": now,
			},
		}

		opts := options.Update().SetUpsert(true)
		_, err = r.collection.UpdateOne(ctx, filter, update, opts)
		return err
	}

	// Date exists, update the existing daily activity
	update := bson.M{
		"$inc": bson.M{
			"daily_activities.$[elem].total_activities": 1,
		},
		"$set": bson.M{
			"updated_at":                          now,
			"daily_activities.$[elem].updated_at": now,
		},
	}

	if fieldToIncrement != "" {
		update["$inc"].(bson.M)["daily_activities.$[elem]."+fieldToIncrement] = 1
	}

	if wordID != "" {
		update["$addToSet"] = bson.M{
			"daily_activities.$[elem].word_ids": wordID,
		}
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{
			bson.M{
				"elem.timestamp": bson.M{
					"$gte": dayStart,
					"$lt":  dayStart.AddDate(0, 0, 1),
				},
			},
		},
	})

	_, err = r.collection.UpdateOne(ctx, filter, update, arrayFilters)
	return err
}

// GetUserSummary retrieves the activity summary for a user
func (r *UserActivitySummaryRepository) GetUserSummary(userID string) (*models.UserActivitySummary, error) {
	ctx := context.Background()
	filter := bson.M{"user_id": userID}

	var summary models.UserActivitySummary
	err := r.collection.FindOne(ctx, filter).Decode(&summary)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// GetDailyActivity retrieves activity for a specific date
func (r *UserActivitySummaryRepository) GetDailyActivity(userID, date string) (*models.DailyActivity, error) {
	ctx := context.Background()

	// Parse the date string to get day boundaries
	dayStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}
	dayEnd := dayStart.AddDate(0, 0, 1)

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$unwind", Value: "$daily_activities"}},
		{{Key: "$match", Value: bson.M{
			"daily_activities.timestamp": bson.M{
				"$gte": dayStart,
				"$lt":  dayEnd,
			},
		}}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$daily_activities"}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.DailyActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	if len(activities) == 0 {
		return nil, nil
	}

	return &activities[0], nil
}

// GetActivitiesByDateRange retrieves activities within a date range
func (r *UserActivitySummaryRepository) GetActivitiesByDateRange(userID, startDate, endDate string) ([]models.DailyActivity, error) {
	ctx := context.Background()

	// Parse date strings
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	end = end.AddDate(0, 0, 1) // Include the entire end date

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$unwind", Value: "$daily_activities"}},
		{{Key: "$match", Value: bson.M{
			"daily_activities.timestamp": bson.M{
				"$gte": start,
				"$lt":  end,
			},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "daily_activities.timestamp", Value: -1}}}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$daily_activities"}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.DailyActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// GetActivityCalendar retrieves daily activities for the past N days in the specified timezone
func (r *UserActivitySummaryRepository) GetActivityCalendar(userID string, days int, timezone string) ([]models.DailyActivity, error) {
	// Load the timezone location
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
	}

	// Get current time in the specified timezone
	now := time.Now().In(loc)
	startDate := now.AddDate(0, 0, -days).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	return r.GetActivitiesByDateRange(userID, startDate, endDate)
}

// DeleteOldActivities removes activities older than the specified number of days
func (r *UserActivitySummaryRepository) DeleteOldActivities(userID string, days int) error {
	ctx := context.Background()
	cutoffDate := time.Now().AddDate(0, 0, -days)

	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$pull": bson.M{
			"daily_activities": bson.M{
				"timestamp": bson.M{"$lt": cutoffDate},
			},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetRecentActivities retrieves the most recent N daily activities
func (r *UserActivitySummaryRepository) GetRecentActivities(userID string, limit int) ([]models.DailyActivity, error) {
	ctx := context.Background()

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$unwind", Value: "$daily_activities"}},
		{{Key: "$sort", Value: bson.D{{Key: "daily_activities.timestamp", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$replaceRoot", Value: bson.M{"newRoot": "$daily_activities"}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.DailyActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}
