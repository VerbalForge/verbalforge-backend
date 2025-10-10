package services

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
	"verbalforge-backend/internal/utils"
)

// DiscussionService handles discussion business logic
type DiscussionService struct {
	discussionRepo *repository.DiscussionRepository
	userRepo       *repository.UserRepository
}

// NewDiscussionService creates a new discussion service
func NewDiscussionService(discussionRepo *repository.DiscussionRepository, userRepo *repository.UserRepository) *DiscussionService {
	return &DiscussionService{
		discussionRepo: discussionRepo,
		userRepo:       userRepo,
	}
}

// GetDiscussions retrieves all discussions with pagination
func (s *DiscussionService) GetDiscussions(page, limit int, sortBy string, tags []string) ([]models.Discussion, int64, error) {
	return s.discussionRepo.FindAll(page, limit, sortBy, tags)
}

// GetDiscussionByID retrieves a discussion by ID
func (s *DiscussionService) GetDiscussionByID(id string) (*models.Discussion, error) {
	discussion, err := s.discussionRepo.FindByID(id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("discussion not found")
		}
		return nil, err
	}
	return discussion, nil
}

// CreateDiscussion creates a new discussion
func (s *DiscussionService) CreateDiscussion(userID, username string, req *models.CreateDiscussionRequest) (*models.Discussion, error) {
	// Get user name
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	discussion := &models.Discussion{
		Title:         req.Title,
		Description:   req.Description,
		QuestionIDs:   req.QuestionIDs,
		Tags:          req.Tags,
		CreatedBy:     userID,
		CreatedByName: user.Name,
	}

	err = s.discussionRepo.Create(discussion)
	if err != nil {
		return nil, err
	}

	return discussion, nil
}

// UpdateDiscussion updates a discussion
func (s *DiscussionService) UpdateDiscussion(discussionID, userID string, req *models.UpdateDiscussionRequest) (*models.Discussion, error) {
	// Get discussion to verify ownership
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return nil, errors.New("discussion not found")
	}

	if discussion.CreatedBy != userID {
		return nil, errors.New("unauthorized to update this discussion")
	}

	updates := make(bson.M)
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.QuestionIDs != nil {
		updates["questionIds"] = req.QuestionIDs
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	err = s.discussionRepo.Update(discussionID, updates)
	if err != nil {
		return nil, err
	}

	return s.discussionRepo.FindByID(discussionID)
}

// DeleteDiscussion deletes a discussion
func (s *DiscussionService) DeleteDiscussion(discussionID, userID string) error {
	// Get discussion to verify ownership
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return errors.New("discussion not found")
	}

	if discussion.CreatedBy != userID {
		return errors.New("unauthorized to delete this discussion")
	}

	return s.discussionRepo.Delete(discussionID)
}

// IncrementView increments view count
func (s *DiscussionService) IncrementView(discussionID, userID string) error {
	return s.discussionRepo.IncrementView(discussionID, userID)
}

// ToggleLike toggles like on a discussion
func (s *DiscussionService) ToggleLike(discussionID, userID string) (bool, error) {
	return s.discussionRepo.ToggleLike(discussionID, userID)
}

// AddComment adds a comment to a discussion
func (s *DiscussionService) AddComment(discussionID, userID, username string, req *models.CreateCommentRequest) error {
	comment := &models.Comment{
		UserID:   userID,
		Username: username,
		Text:     req.Text,
	}

	return s.discussionRepo.AddComment(discussionID, comment)
}

// UpdateComment updates a comment
func (s *DiscussionService) UpdateComment(discussionID, commentID, userID string, req *models.UpdateCommentRequest) error {
	// Get discussion to verify comment ownership
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return errors.New("discussion not found")
	}

	// Find comment and verify ownership
	found := false
	for _, comment := range discussion.Comments {
		if comment.ID == commentID {
			if comment.UserID != userID {
				return errors.New("unauthorized to update this comment")
			}
			found = true
			break
		}
	}

	if !found {
		return errors.New("comment not found")
	}

	return s.discussionRepo.UpdateComment(discussionID, commentID, req.Text)
}

// DeleteComment deletes a comment
func (s *DiscussionService) DeleteComment(discussionID, commentID, userID string) error {
	// Get discussion to verify comment ownership
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return errors.New("discussion not found")
	}

	// Find comment and verify ownership
	found := false
	for _, comment := range discussion.Comments {
		if comment.ID == commentID {
			if comment.UserID != userID {
				return errors.New("unauthorized to delete this comment")
			}
			found = true
			break
		}
	}

	if !found {
		return errors.New("comment not found")
	}

	return s.discussionRepo.DeleteComment(discussionID, commentID)
}

// ToggleCommentLike toggles like on a comment
func (s *DiscussionService) ToggleCommentLike(discussionID, commentID, userID string) (bool, error) {
	return s.discussionRepo.ToggleCommentLike(discussionID, commentID, userID)
}

// GetUserDiscussions retrieves discussions by a user
func (s *DiscussionService) GetUserDiscussions(userID string, page, limit int) ([]models.Discussion, int64, error) {
	return s.discussionRepo.FindByUser(userID, page, limit)
}

// SearchDiscussions searches discussions
func (s *DiscussionService) SearchDiscussions(query string, page, limit int) ([]models.Discussion, int64, error) {
	return s.discussionRepo.Search(query, page, limit)
}

// GetDiscussionsWithCursor retrieves discussions with cursor-based pagination
func (s *DiscussionService) GetDiscussionsWithCursor(limit int64, cursor, sortBy string, tags []string) (*models.DiscussionsResponse, error) {
	// Fetch discussions with cursor
	discussions, err := s.discussionRepo.FindAllWithCursor(limit, cursor, sortBy, tags)
	if err != nil {
		return nil, err
	}

	// Determine if there are more results
	hasMore := int64(len(discussions)) > limit
	if hasMore {
		discussions = discussions[:limit] // Remove the extra item
	}

	// Create next cursor if there are more results
	var nextCursor string
	if hasMore && len(discussions) > 0 {
		lastDiscussion := discussions[len(discussions)-1]
		cursorData := utils.CursorData{
			LastID:        lastDiscussion.ID,
			LastCreatedAt: lastDiscussion.CreatedAt.Format(time.RFC3339),
		}
		nextCursor, err = utils.EncodeCursor(cursorData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	// Get total count for statistics
	filter := bson.M{}
	if len(tags) > 0 {
		filter["tags"] = bson.M{"$in": tags}
	}
	total, err := s.discussionRepo.Count(filter)
	if err != nil {
		return nil, err
	}

	return &models.DiscussionsResponse{
		Discussions: discussions,
		Total:       total,
		Limit:       int(limit),
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}

// SearchDiscussionsWithCursor searches discussions with cursor-based pagination
func (s *DiscussionService) SearchDiscussionsWithCursor(query string, limit int64, cursor string) (*models.DiscussionsResponse, error) {
	// Fetch discussions with cursor
	discussions, err := s.discussionRepo.SearchWithCursor(query, limit, cursor)
	if err != nil {
		return nil, err
	}

	// Determine if there are more results
	hasMore := int64(len(discussions)) > limit
	if hasMore {
		discussions = discussions[:limit] // Remove the extra item
	}

	// Create next cursor if there are more results
	var nextCursor string
	if hasMore && len(discussions) > 0 {
		lastDiscussion := discussions[len(discussions)-1]
		cursorData := utils.CursorData{
			LastID:        lastDiscussion.ID,
			LastCreatedAt: lastDiscussion.CreatedAt.Format(time.RFC3339),
		}
		nextCursor, err = utils.EncodeCursor(cursorData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	// Get total count for search results
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
		},
	}
	total, err := s.discussionRepo.Count(filter)
	if err != nil {
		return nil, err
	}

	return &models.DiscussionsResponse{
		Discussions: discussions,
		Total:       total,
		Limit:       int(limit),
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}

// GetUserDiscussionsWithCursor retrieves user's discussions with cursor-based pagination
func (s *DiscussionService) GetUserDiscussionsWithCursor(username string, limit int64, cursor string) (*models.DiscussionsResponse, error) {
	// Get user ID from username
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Fetch discussions with cursor
	discussions, err := s.discussionRepo.FindByUserWithCursor(user.ID, limit, cursor)
	if err != nil {
		return nil, err
	}

	// Determine if there are more results
	hasMore := int64(len(discussions)) > limit
	if hasMore {
		discussions = discussions[:limit] // Remove the extra item
	}

	// Create next cursor if there are more results
	var nextCursor string
	if hasMore && len(discussions) > 0 {
		lastDiscussion := discussions[len(discussions)-1]
		cursorData := utils.CursorData{
			LastID:        lastDiscussion.ID,
			LastCreatedAt: lastDiscussion.CreatedAt.Format(time.RFC3339),
		}
		nextCursor, err = utils.EncodeCursor(cursorData)
		if err != nil {
			return nil, fmt.Errorf("failed to encode cursor: %w", err)
		}
	}

	// Get total count for user's discussions
	filter := bson.M{"createdBy": user.ID}
	total, err := s.discussionRepo.Count(filter)
	if err != nil {
		return nil, err
	}

	return &models.DiscussionsResponse{
		Discussions: discussions,
		Total:       total,
		Limit:       int(limit),
		NextCursor:  nextCursor,
		HasMore:     hasMore,
	}, nil
}
