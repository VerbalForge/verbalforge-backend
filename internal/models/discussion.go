package models

import "time"

// Comment represents a discussion comment
type Comment struct {
	ID        string    `json:"id" bson:"_id"`
	UserID    string    `json:"userId" bson:"userId"`
	Username  string    `json:"username" bson:"username"`
	Text      string    `json:"text" bson:"text"`
	Likes     int       `json:"likes" bson:"likes"`
	LikedBy   []string  `json:"likedBy" bson:"likedBy"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// LinkedQuestion represents a question linked to a discussion
type LinkedQuestion struct {
	QuestionID      string `json:"questionId" bson:"questionId"`
	PassageID       string `json:"passageId,omitempty" bson:"passageId,omitempty"`
	QuestionText    string `json:"questionText" bson:"questionText"`
	QuestionType    string `json:"questionType" bson:"questionType"`
	DifficultyLevel string `json:"difficultyLevel" bson:"difficultyLevel"`
	PassageTitle    string `json:"passageTitle,omitempty" bson:"passageTitle,omitempty"`
}

// Discussion represents a discussion topic
type Discussion struct {
	ID             string          `json:"id" bson:"_id,omitempty"`
	Title          string          `json:"title" bson:"title"`
	Description    string          `json:"description" bson:"description"`       // Rich text HTML
	QuestionIDs    []string        `json:"questionIds" bson:"questionIds"`       // Deprecated, kept for backward compatibility
	DiscussionType string          `json:"discussionType" bson:"discussionType"` // "general" or "question_linked"
	LinkedQuestion *LinkedQuestion `json:"linkedQuestion,omitempty" bson:"linkedQuestion,omitempty"`
	CreatedBy      string          `json:"createdBy" bson:"createdBy"`
	CreatedByName  string          `json:"createdByName" bson:"createdByName"`
	CreatedAt      time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt" bson:"updatedAt"`
	Comments       []Comment       `json:"comments" bson:"comments"`
	Views          int             `json:"views" bson:"views"`
	ViewedBy       []string        `json:"viewedBy" bson:"viewedBy"`
	Likes          int             `json:"likes" bson:"likes"`
	LikedBy        []string        `json:"likedBy" bson:"likedBy"`
	Tags           []string        `json:"tags" bson:"tags"`
	IsPinned       bool            `json:"isPinned" bson:"isPinned"`
	IsLocked       bool            `json:"isLocked" bson:"isLocked"`
	CommentCount   int             `json:"commentCount" bson:"commentCount"`
}

// CreateDiscussionRequest represents the request to create a discussion
type CreateDiscussionRequest struct {
	Title       string   `json:"title" binding:"required,min=5,max=200"`
	Description string   `json:"description" binding:"required,min=10"`
	QuestionIDs []string `json:"questionIds"` // Deprecated
	Tags        []string `json:"tags"`
	QuestionID  string   `json:"questionId,omitempty"` // Optional: if provided, creates a question-linked discussion
}

// UpdateDiscussionRequest represents the request to update a discussion
type UpdateDiscussionRequest struct {
	Title       string   `json:"title" binding:"min=5,max=200"`
	Description string   `json:"description" binding:"min=10"`
	QuestionIDs []string `json:"questionIds"`
	Tags        []string `json:"tags"`
}

// CreateCommentRequest represents the request to create a comment
type CreateCommentRequest struct {
	Text string `json:"text" binding:"required,min=1"`
}

// UpdateCommentRequest represents the request to update a comment
type UpdateCommentRequest struct {
	Text string `json:"text" binding:"required,min=1"`
}

// DiscussionsResponse represents paginated discussions response with cursor-based pagination
type DiscussionsResponse struct {
	Discussions    []Discussion `json:"discussions"`
	Total          int64        `json:"total"`
	Limit          int          `json:"limit"`
	NextCursor     string       `json:"nextCursor,omitempty"`     // Continuation token for next page
	PreviousCursor string       `json:"previousCursor,omitempty"` // Continuation token for previous page
	HasMore        bool         `json:"hasMore"`                  // Indicates if more results exist
}
