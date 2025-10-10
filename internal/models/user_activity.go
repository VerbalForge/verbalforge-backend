package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// ActivityType represents the type of user activity
type ActivityType string

const (
	// Question activities
	ActivityQuestion ActivityType = "question" // Covers both attempted and solved

	// Discussion activities
	ActivityDiscussion ActivityType = "discussion" // Covers all discussion/comment activities

	// Other activities
	ActivityProfileViewed ActivityType = "profile_viewed"
	ActivityStreakUpdated ActivityType = "streak_updated"
)

// DiscussionActionType represents the specific action type for discussion activities
type DiscussionActionType string

const (
	// Discussion actions
	ActionDiscussionCreated DiscussionActionType = "created"
	ActionDiscussionUpdated DiscussionActionType = "updated"

	// Comment actions
	ActionCommentAdded   DiscussionActionType = "comment_added"
	ActionCommentUpdated DiscussionActionType = "comment_updated"

	// Like actions
	ActionDiscussionLiked DiscussionActionType = "discussion_liked"
	ActionCommentLiked    DiscussionActionType = "comment_liked"
)

// QuestionActivityMetadata represents metadata for question activity (attempted or solved)
type QuestionActivityMetadata struct {
	QuestionID      string `bson:"question_id" json:"questionId"`
	Solved          bool   `bson:"solved" json:"solved"` // false = attempted, true = solved
	PassageID       string `bson:"passage_id,omitempty" json:"passageId,omitempty"`
	DifficultyLevel string `bson:"difficulty_level" json:"difficulty_level"`
	QuestionType    string `bson:"question_type" json:"question_type"`
	TimeTaken       int    `bson:"time_taken" json:"timeTaken"`
	XPGained        int    `bson:"xp_gained,omitempty" json:"xpGained,omitempty"` // 0 if not solved
}

// DiscussionActivityMetadata represents metadata for all discussion-related activities
type DiscussionActivityMetadata struct {
	DiscussionID string               `bson:"discussion_id" json:"discussionId"`
	CommentID    string               `bson:"comment_id,omitempty" json:"commentId,omitempty"` // For comment activities
	ParentID     string               `bson:"parent_id,omitempty" json:"parentId,omitempty"`   // For nested comments
	Title        string               `bson:"title,omitempty" json:"title,omitempty"`          // For created/updated
	Tags         []string             `bson:"tags,omitempty" json:"tags,omitempty"`            // For created
	ActionType   DiscussionActionType `bson:"action_type" json:"actionType"`                   // Type of discussion action
	IsComment    bool                 `bson:"is_comment" json:"isComment"`                     // true for comment actions, false for discussion actions
}

// LikedMetadata represents metadata for liked activity (deprecated - use DiscussionActivityMetadata)
type LikedMetadata struct {
	TargetID   string `bson:"target_id" json:"targetId"`     // ID of post/comment/etc
	TargetType string `bson:"target_type" json:"targetType"` // "post", "comment", etc
}

// ProfileViewedMetadata represents metadata for profile viewed activity
type ProfileViewedMetadata struct {
	ViewedUserID string `bson:"viewed_user_id" json:"viewedUserId"`
	ViewerID     string `bson:"viewer_id,omitempty" json:"viewerId,omitempty"`
}

// StreakUpdatedMetadata represents metadata for streak updated activity
type StreakUpdatedMetadata struct {
	OldStreak int `bson:"old_streak" json:"oldStreak"`
	NewStreak int `bson:"new_streak" json:"newStreak"`
}

// UserActivity represents a single user activity event
type UserActivity struct {
	ID           string                 `bson:"_id,omitempty" json:"id,omitempty"`
	UserID       string                 `bson:"user_id" json:"userId"`
	ActivityType ActivityType           `bson:"activity_type" json:"activityType"`
	Date         string                 `bson:"date" json:"date"`
	Timestamp    time.Time              `bson:"timestamp" json:"timestamp"`
	Metadata     map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt    time.Time              `bson:"created_at" json:"createdAt"`
}

// Helper methods to convert typed metadata to/from map

// GetQuestionActivityMetadata extracts question activity metadata
func (a *UserActivity) GetQuestionActivityMetadata() (*QuestionActivityMetadata, error) {
	var meta QuestionActivityMetadata
	if err := unmarshalMetadata(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetDiscussionActivityMetadata extracts discussion activity metadata
func (a *UserActivity) GetDiscussionActivityMetadata() (*DiscussionActivityMetadata, error) {
	var meta DiscussionActivityMetadata
	if err := unmarshalMetadata(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetLikedMetadata extracts liked metadata (deprecated)
func (a *UserActivity) GetLikedMetadata() (*LikedMetadata, error) {
	var meta LikedMetadata
	if err := unmarshalMetadata(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetProfileViewedMetadata extracts profile viewed metadata
func (a *UserActivity) GetProfileViewedMetadata() (*ProfileViewedMetadata, error) {
	var meta ProfileViewedMetadata
	if err := unmarshalMetadata(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// GetStreakUpdatedMetadata extracts streak updated metadata
func (a *UserActivity) GetStreakUpdatedMetadata() (*StreakUpdatedMetadata, error) {
	var meta StreakUpdatedMetadata
	if err := unmarshalMetadata(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// Helper function to unmarshal metadata map to typed struct
func unmarshalMetadata(metadata map[string]interface{}, target interface{}) error {
	bsonBytes, err := bson.Marshal(metadata)
	if err != nil {
		return err
	}
	return bson.Unmarshal(bsonBytes, target)
}

// ToMetadataMap converts typed metadata to map
func ToMetadataMap(metadata interface{}) (map[string]interface{}, error) {
	bsonBytes, err := bson.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	var metadataMap map[string]interface{}
	if err := bson.Unmarshal(bsonBytes, &metadataMap); err != nil {
		return nil, err
	}
	return metadataMap, nil
}
