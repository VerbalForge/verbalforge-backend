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

// UserActivityRepository handles user activity database operations
type UserActivityRepository struct {
	activityCollection *mongo.Collection
}

// NewUserActivityRepository creates a new user activity repository
func NewUserActivityRepository(db *mongo.Database) *UserActivityRepository {
	return &UserActivityRepository{
		activityCollection: db.Collection("user_activity"),
	}
}

// LogActivity records a user activity event (generic version)
func (r *UserActivityRepository) LogActivity(userID string, activityType models.ActivityType, metadata map[string]interface{}) error {
	now := time.Now()
	activity := &models.UserActivity{
		ID:           uuid.New().String(),
		UserID:       userID,
		ActivityType: activityType,
		Date:         now.Format("2006-01-02"),
		Timestamp:    now,
		Metadata:     metadata,
		CreatedAt:    now,
	}

	ctx := context.Background()
	_, err := r.activityCollection.InsertOne(ctx, activity)
	return err
}

// GetRecentActivities retrieves recent activities for a user
func (r *UserActivityRepository) GetRecentActivities(userID string, limit int) ([]models.UserActivity, error) {
	ctx := context.Background()
	filter := bson.M{"user_id": userID}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.activityCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.UserActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// GetActivitiesByType retrieves activities of a specific type for a user
func (r *UserActivityRepository) GetActivitiesByType(userID string, activityType models.ActivityType, limit int) ([]models.UserActivity, error) {
	ctx := context.Background()
	filter := bson.M{
		"user_id":       userID,
		"activity_type": activityType,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.activityCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.UserActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// GetActivitiesByDateRange retrieves activities within a date range
func (r *UserActivityRepository) GetActivitiesByDateRange(userID string, startDate, endDate string) ([]models.UserActivity, error) {
	ctx := context.Background()
	filter := bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := r.activityCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.UserActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// FindAll returns activities with filters and pagination
func (r *UserActivityRepository) FindAll(filter bson.M, limit int64, skip int64) ([]models.UserActivity, error) {
	ctx := context.Background()

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.activityCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []models.UserActivity
	if err = cursor.All(ctx, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}

// Count returns the total count of activities matching the filter
func (r *UserActivityRepository) Count(filter bson.M) (int64, error) {
	ctx := context.Background()
	count, err := r.activityCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Aggregate performs aggregation pipeline on activities
func (r *UserActivityRepository) Aggregate(pipeline []bson.M) (*mongo.Cursor, error) {
	ctx := context.Background()

	// Convert []bson.M to []interface{} for the pipeline
	interfacePipeline := make([]interface{}, len(pipeline))
	for i, stage := range pipeline {
		interfacePipeline[i] = stage
	}

	cursor, err := r.activityCollection.Aggregate(ctx, interfacePipeline)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}
