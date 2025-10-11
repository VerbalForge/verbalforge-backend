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
	discussionRepo  *repository.DiscussionRepository
	userRepo        *repository.UserRepository
	questionRepo    *repository.QuestionRepository
	passageRepo     *repository.PassageRepository
	activityService *UserActivityService
}

// NewDiscussionService creates a new discussion service
func NewDiscussionService(
	discussionRepo *repository.DiscussionRepository,
	userRepo *repository.UserRepository,
	questionRepo *repository.QuestionRepository,
	passageRepo *repository.PassageRepository,
	activityService *UserActivityService,
) *DiscussionService {
	return &DiscussionService{
		discussionRepo:  discussionRepo,
		userRepo:        userRepo,
		questionRepo:    questionRepo,
		passageRepo:     passageRepo,
		activityService: activityService,
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

	// Determine discussion type based on whether questionId is provided
	discussionType := "general"
	var linkedQuestion *models.LinkedQuestion

	// If questionId is provided, make it a question-linked discussion
	if req.QuestionID != "" {
		discussionType = "question_linked"

		// Fetch question details
		question, err := s.questionRepo.FindByID(req.QuestionID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.New("question not found")
			}
			return nil, fmt.Errorf("failed to fetch question: %w", err)
		}

		// Prepare linked question data
		linkedQuestion = &models.LinkedQuestion{
			QuestionID:      question.ID,
			QuestionText:    question.QuestionText,
			QuestionType:    question.QuestionType,
			DifficultyLevel: question.DifficultyLevel,
		}

		// If passage-based question, fetch passage title
		if question.PassageID != "" {
			passage, err := s.passageRepo.FindByID(question.PassageID)
			if err == nil && passage != nil {
				linkedQuestion.PassageID = passage.ID
				linkedQuestion.PassageTitle = passage.Title
			}
		}
	}

	discussion := &models.Discussion{
		Title:          req.Title,
		Description:    req.Description,
		QuestionIDs:    req.QuestionIDs,
		Tags:           req.Tags,
		CreatedBy:      userID,
		CreatedByName:  user.Name,
		DiscussionType: discussionType,
		LinkedQuestion: linkedQuestion,
	}

	err = s.discussionRepo.Create(discussion)
	if err != nil {
		return nil, err
	}

	// Log discussion created activity
	if s.activityService != nil {
		discussionTitle := req.Title
		if len(discussionTitle) > 150 {
			discussionTitle = discussionTitle[:150] + "..."
		}

		activityMeta := models.DiscussionActivityMetadata{
			DiscussionID:    discussion.ID,
			DiscussionTitle: discussionTitle,
			Tags:            req.Tags,
			ActionType:      models.ActionDiscussionCreated,
			IsComment:       false,
		}
		if err := s.activityService.LogDiscussionActivity(userID, activityMeta); err != nil {
			fmt.Printf("Failed to log discussion activity: %v\n", err)
		}
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

	// Get the updated discussion for the title
	updatedDiscussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return nil, err
	}

	// Log discussion updated activity
	if s.activityService != nil {
		discussionTitle := updatedDiscussion.Title
		if len(discussionTitle) > 150 {
			discussionTitle = discussionTitle[:150] + "..."
		}

		activityMeta := models.DiscussionActivityMetadata{
			DiscussionID:    discussionID,
			DiscussionTitle: discussionTitle,
			ActionType:      models.ActionDiscussionUpdated,
			IsComment:       false,
		}
		if err := s.activityService.LogDiscussionActivity(userID, activityMeta); err != nil {
			fmt.Printf("Failed to log discussion update activity: %v\n", err)
		}
	}

	return updatedDiscussion, nil
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
	// Get discussion for the title
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return false, errors.New("discussion not found")
	}

	liked, err := s.discussionRepo.ToggleLike(discussionID, userID)
	if err != nil {
		return false, err
	}

	// Only log activity if the user liked (not unliked)
	if liked && s.activityService != nil {
		discussionTitle := discussion.Title
		if len(discussionTitle) > 150 {
			discussionTitle = discussionTitle[:150] + "..."
		}

		activityMeta := models.DiscussionActivityMetadata{
			DiscussionID:    discussionID,
			DiscussionTitle: discussionTitle,
			ActionType:      models.ActionDiscussionLiked,
			IsComment:       false,
		}
		if err := s.activityService.LogDiscussionActivity(userID, activityMeta); err != nil {
			fmt.Printf("Failed to log discussion like activity: %v\n", err)
		}
	}

	return liked, nil
}

// AddComment adds a comment to a discussion
func (s *DiscussionService) AddComment(discussionID, userID, username string, req *models.CreateCommentRequest) error {
	// Get discussion for the title
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return errors.New("discussion not found")
	}

	comment := &models.Comment{
		UserID:   userID,
		Username: username,
		Text:     req.Text,
	}

	err = s.discussionRepo.AddComment(discussionID, comment)
	if err != nil {
		return err
	}

	// Log comment added activity
	if s.activityService != nil {
		discussionTitle := discussion.Title
		if len(discussionTitle) > 150 {
			discussionTitle = discussionTitle[:150] + "..."
		}

		activityMeta := models.DiscussionActivityMetadata{
			DiscussionID:    discussionID,
			DiscussionTitle: discussionTitle,
			CommentID:       comment.ID,
			ActionType:      models.ActionCommentAdded,
			IsComment:       true,
		}
		if err := s.activityService.LogDiscussionActivity(userID, activityMeta); err != nil {
			fmt.Printf("Failed to log comment activity: %v\n", err)
		}
	}

	return nil
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

	err = s.discussionRepo.UpdateComment(discussionID, commentID, req.Text)
	if err != nil {
		return err
	}

	// Log comment updated activity
	if s.activityService != nil {
		discussionTitle := discussion.Title
		if len(discussionTitle) > 150 {
			discussionTitle = discussionTitle[:150] + "..."
		}

		activityMeta := models.DiscussionActivityMetadata{
			DiscussionID:    discussionID,
			DiscussionTitle: discussionTitle,
			CommentID:       commentID,
			ActionType:      models.ActionCommentUpdated,
			IsComment:       true,
		}
		if err := s.activityService.LogDiscussionActivity(userID, activityMeta); err != nil {
			fmt.Printf("Failed to log comment update activity: %v\n", err)
		}
	}

	return nil
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
	// Get discussion to find the comment text
	discussion, err := s.discussionRepo.FindByID(discussionID)
	if err != nil {
		return false, errors.New("discussion not found")
	}

	liked, err := s.discussionRepo.ToggleCommentLike(discussionID, commentID, userID)
	if err != nil {
		return false, err
	}

	// Only log activity if the user liked (not unliked)
	if liked && s.activityService != nil {
		// Find the comment text
		var commentText string
		for _, comment := range discussion.Comments {
			if comment.ID == commentID {
				commentText = comment.Text
				break
			}
		}

		// Truncate comment text to 150 chars
		if len(commentText) > 150 {
			commentText = commentText[:150] + "..."
		}

		activityMeta := models.DiscussionActivityMetadata{
			DiscussionID:    discussionID,
			DiscussionTitle: commentText, // Using DiscussionTitle field for comment text
			CommentID:       commentID,
			ActionType:      models.ActionCommentLiked,
			IsComment:       true,
		}
		if err := s.activityService.LogDiscussionActivity(userID, activityMeta); err != nil {
			fmt.Printf("Failed to log comment like activity: %v\n", err)
		}
	}

	return liked, nil
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
