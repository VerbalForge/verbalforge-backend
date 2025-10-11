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

// DiscussionRepository handles discussion database operations
type DiscussionRepository struct {
	collection *mongo.Collection
}

// NewDiscussionRepository creates a new discussion repository
func NewDiscussionRepository(db *mongo.Database) *DiscussionRepository {
	return &DiscussionRepository{
		collection: db.Collection("discussions"),
	}
}

// FindAll retrieves all discussions with pagination and sorting
func (r *DiscussionRepository) FindAll(page, limit int, sortBy string, tags []string) ([]models.Discussion, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter
	filter := bson.M{}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}

	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Build sort options
	sortOrder := -1 // Default: newest first
	sortField := "createdAt"
	switch sortBy {
	case "oldest":
		sortOrder = 1
	case "popular":
		sortField = "likes"
	case "views":
		sortField = "views"
	case "updated":
		sortField = "updatedAt"
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "isPinned", Value: -1}, {Key: sortField, Value: sortOrder}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var discussions []models.Discussion
	if err = cursor.All(ctx, &discussions); err != nil {
		return nil, 0, err
	}

	return discussions, total, nil
}

// FindByID retrieves a single discussion by ID
func (r *DiscussionRepository) FindByID(id string) (*models.Discussion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var discussion models.Discussion
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&discussion)
	if err != nil {
		return nil, err
	}

	return &discussion, nil
}

// Create creates a new discussion
func (r *DiscussionRepository) Create(discussion *models.Discussion) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discussion.ID = uuid.New().String()
	discussion.CreatedAt = time.Now()
	discussion.UpdatedAt = time.Now()
	discussion.Comments = []models.Comment{}
	discussion.Views = 0
	discussion.ViewedBy = []string{}
	discussion.Likes = 0
	discussion.LikedBy = []string{}
	discussion.CommentCount = 0
	discussion.IsPinned = false
	discussion.IsLocked = false

	// Default to general if not specified (for backward compatibility)
	if discussion.DiscussionType == "" {
		discussion.DiscussionType = "general"
	}

	if discussion.QuestionIDs == nil {
		discussion.QuestionIDs = []string{}
	}
	if discussion.Tags == nil {
		discussion.Tags = []string{}
	}

	_, err := r.collection.InsertOne(ctx, discussion)
	return err
}

// Update updates an existing discussion
func (r *DiscussionRepository) Update(id string, updates bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updates["updatedAt"] = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updates},
	)
	return err
}

// Delete deletes a discussion
func (r *DiscussionRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// IncrementView increments view count for a discussion
func (r *DiscussionRepository) IncrementView(discussionID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user already viewed
	var discussion models.Discussion
	err := r.collection.FindOne(ctx, bson.M{"_id": discussionID}).Decode(&discussion)
	if err != nil {
		return err
	}

	// Only increment if user hasn't viewed before
	for _, id := range discussion.ViewedBy {
		if id == userID {
			return nil // Already viewed
		}
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": discussionID},
		bson.M{
			"$inc":  bson.M{"views": 1},
			"$push": bson.M{"viewedBy": userID},
		},
	)
	return err
}

// ToggleLike toggles like status for a discussion
func (r *DiscussionRepository) ToggleLike(discussionID, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var discussion models.Discussion
	err := r.collection.FindOne(ctx, bson.M{"_id": discussionID}).Decode(&discussion)
	if err != nil {
		return false, err
	}

	// Check if user already liked
	liked := false
	for _, id := range discussion.LikedBy {
		if id == userID {
			liked = true
			break
		}
	}

	var update bson.M
	if liked {
		// Unlike
		update = bson.M{
			"$inc":  bson.M{"likes": -1},
			"$pull": bson.M{"likedBy": userID},
		}
	} else {
		// Like
		update = bson.M{
			"$inc":  bson.M{"likes": 1},
			"$push": bson.M{"likedBy": userID},
		}
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": discussionID}, update)
	return !liked, err
}

// AddComment adds a comment to a discussion
func (r *DiscussionRepository) AddComment(discussionID string, comment *models.Comment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()
	comment.Likes = 0
	comment.LikedBy = []string{}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": discussionID},
		bson.M{
			"$push": bson.M{"comments": comment},
			"$inc":  bson.M{"commentCount": 1},
			"$set":  bson.M{"updatedAt": time.Now()},
		},
	)
	return err
}

// UpdateComment updates a comment in a discussion
func (r *DiscussionRepository) UpdateComment(discussionID, commentID, newText string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": discussionID, "comments._id": commentID},
		bson.M{
			"$set": bson.M{
				"comments.$.text":      newText,
				"comments.$.updatedAt": time.Now(),
			},
		},
	)
	return err
}

// DeleteComment deletes a comment from a discussion
func (r *DiscussionRepository) DeleteComment(discussionID, commentID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": discussionID},
		bson.M{
			"$pull": bson.M{"comments": bson.M{"_id": commentID}},
			"$inc":  bson.M{"commentCount": -1},
		},
	)
	return err
}

// ToggleCommentLike toggles like status for a comment
func (r *DiscussionRepository) ToggleCommentLike(discussionID, commentID, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var discussion models.Discussion
	err := r.collection.FindOne(ctx, bson.M{"_id": discussionID}).Decode(&discussion)
	if err != nil {
		return false, err
	}

	// Find the comment and check if user already liked
	liked := false
	for _, comment := range discussion.Comments {
		if comment.ID == commentID {
			for _, id := range comment.LikedBy {
				if id == userID {
					liked = true
					break
				}
			}
			break
		}
	}

	var update bson.M
	if liked {
		// Unlike
		update = bson.M{
			"$inc":  bson.M{"comments.$[elem].likes": -1},
			"$pull": bson.M{"comments.$[elem].likedBy": userID},
		}
	} else {
		// Like
		update = bson.M{
			"$inc":  bson.M{"comments.$[elem].likes": 1},
			"$push": bson.M{"comments.$[elem].likedBy": userID},
		}
	}

	arrayFilters := options.Update().SetArrayFilters(options.ArrayFilters{
		Filters: []interface{}{bson.M{"elem._id": commentID}},
	})

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": discussionID},
		update,
		arrayFilters,
	)
	return !liked, err
}

// FindByUser retrieves all discussions created by a specific user
func (r *DiscussionRepository) FindByUser(userID string, page, limit int) ([]models.Discussion, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"createdBy": userID}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var discussions []models.Discussion
	if err = cursor.All(ctx, &discussions); err != nil {
		return nil, 0, err
	}

	return discussions, total, nil
}

// Search searches discussions by title, description, or tags
func (r *DiscussionRepository) Search(query string, page, limit int) ([]models.Discussion, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "createdAt", Value: -1}}).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var discussions []models.Discussion
	if err = cursor.All(ctx, &discussions); err != nil {
		return nil, 0, err
	}

	return discussions, total, nil
}
