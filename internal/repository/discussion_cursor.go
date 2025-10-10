package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/utils"
)

// FindAllWithCursor retrieves discussions with cursor-based pagination
func (r *DiscussionRepository) FindAllWithCursor(limit int64, cursor, sortBy string, tags []string) ([]models.Discussion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	// Decode cursor if provided
	if cursor != "" {
		cursorData, err := utils.DecodeCursor(cursor)
		if err == nil && cursorData != nil {
			// Add cursor condition for pagination
			filter["$or"] = []bson.M{
				{"createdAt": bson.M{"$lt": cursorData.LastCreatedAt}},
				{
					"createdAt": cursorData.LastCreatedAt,
					"_id":       bson.M{"$lt": cursorData.LastID},
				},
			}
		}
	}

	// Build sort options
	sortOrder := -1 // Default: newest first
	sortField := "createdAt"
	switch sortBy {
	case "oldest":
		sortOrder = 1
		// For oldest, we need to reverse the cursor logic
		if cursor != "" {
			cursorData, _ := utils.DecodeCursor(cursor)
			if cursorData != nil {
				filter["$or"] = []bson.M{
					{"createdAt": bson.M{"$gt": cursorData.LastCreatedAt}},
					{
						"createdAt": cursorData.LastCreatedAt,
						"_id":       bson.M{"$gt": cursorData.LastID},
					},
				}
			}
		}
	case "popular":
		sortField = "likes"
	case "views":
		sortField = "views"
	case "updated":
		sortField = "updatedAt"
	}

	// Fetch one extra item to determine if there are more results
	opts := options.Find().
		SetSort(bson.D{{Key: "isPinned", Value: -1}, {Key: sortField, Value: sortOrder}, {Key: "_id", Value: sortOrder}}).
		SetLimit(limit + 1)

	findCursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer findCursor.Close(ctx)

	var discussions []models.Discussion
	if err = findCursor.All(ctx, &discussions); err != nil {
		return nil, err
	}

	return discussions, nil
}

// SearchWithCursor searches discussions with cursor-based pagination
func (r *DiscussionRepository) SearchWithCursor(query string, limit int64, cursor string) ([]models.Discussion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// Decode cursor if provided
	if cursor != "" {
		cursorData, err := utils.DecodeCursor(cursor)
		if err == nil && cursorData != nil {
			existingFilter := filter
			filter = bson.M{
				"$and": []bson.M{
					existingFilter,
					{
						"$or": []bson.M{
							{"createdAt": bson.M{"$lt": cursorData.LastCreatedAt}},
							{
								"createdAt": cursorData.LastCreatedAt,
								"_id":       bson.M{"$lt": cursorData.LastID},
							},
						},
					},
				},
			}
		}
	}

	// Fetch one extra item to determine if there are more results
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit + 1)

	findCursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer findCursor.Close(ctx)

	var discussions []models.Discussion
	if err = findCursor.All(ctx, &discussions); err != nil {
		return nil, err
	}

	return discussions, nil
}

// FindByUserWithCursor retrieves discussions by user with cursor-based pagination
func (r *DiscussionRepository) FindByUserWithCursor(userID string, limit int64, cursor string) ([]models.Discussion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"createdBy": userID}

	// Decode cursor if provided
	if cursor != "" {
		cursorData, err := utils.DecodeCursor(cursor)
		if err == nil && cursorData != nil {
			filter["$or"] = []bson.M{
				{"createdAt": bson.M{"$lt": cursorData.LastCreatedAt}},
				{
					"createdAt": cursorData.LastCreatedAt,
					"_id":       bson.M{"$lt": cursorData.LastID},
				},
			}
		}
	}

	// Fetch one extra item to determine if there are more results
	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetLimit(limit + 1)

	findCursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer findCursor.Close(ctx)

	var discussions []models.Discussion
	if err = findCursor.All(ctx, &discussions); err != nil {
		return nil, err
	}

	return discussions, nil
}

// Count returns total count of discussions (for stats)
func (r *DiscussionRepository) Count(filter bson.M) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.collection.CountDocuments(ctx, filter)
}
