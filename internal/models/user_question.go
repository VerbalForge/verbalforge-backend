package models

import "time"

// UserQuestion represents a user's attempt on a question
type UserQuestion struct {
	ID              string     `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          string     `bson:"user_id" json:"userId"`
	QuestionID      string     `bson:"question_id" json:"questionId"`
	Solved          bool       `bson:"solved" json:"solved"`
	Attempted       bool       `bson:"attempted" json:"attempted"`
	LastAttemptAt   time.Time  `bson:"last_attempt_at" json:"lastAttemptAt"`
	LastSolvedAt    *time.Time `bson:"last_solved_at,omitempty" json:"lastSolvedAt,omitempty"`
	TimeTaken       int        `bson:"time_taken" json:"timeTaken"` // in seconds
	DifficultyLevel string     `bson:"difficulty_level" json:"difficulty_level"`
	QuestionType    string     `bson:"question_type" json:"question_type"`
	XPGained        int        `bson:"xp_gained" json:"xpGained"`
	CreatedAt       time.Time  `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `bson:"updated_at" json:"updatedAt"`
}

// PassageProgress represents derived passage progress
type PassageProgress struct {
	PassageID         string   `json:"passageId"`
	SolvedQuestionIDs []string `json:"solvedQuestionIds"`
	TotalQuestions    int      `json:"totalQuestions"`
	Solved            bool     `json:"solved"`
	Attempted         bool     `json:"attempted"`
}

// SubmitQuestionAttemptRequest represents the request to submit a question attempt
type SubmitQuestionAttemptRequest struct {
	Solved    *bool `json:"solved" binding:"required"`
	TimeTaken int   `json:"timeTaken" binding:"required"`
}

// SubmitPassageAttemptRequest represents the request to submit a passage attempt
type SubmitPassageAttemptRequest struct {
	QuestionAttempts []QuestionAttempt `json:"questionAttempts" binding:"required"`
}

// QuestionAttempt represents a single question attempt within a passage
type QuestionAttempt struct {
	QuestionID string `json:"questionId" binding:"required"`
	Solved     bool   `json:"solved" binding:"required"`
	TimeTaken  int    `json:"timeTaken" binding:"required"`
}

// UserQuestionResponse represents a question progress response
type UserQuestionResponse struct {
	UserID          string     `json:"userId"`
	QuestionID      string     `json:"questionId"`
	Solved          bool       `json:"solved"`
	Attempted       bool       `json:"attempted"`
	LastAttemptAt   *time.Time `json:"lastAttemptAt,omitempty"`
	LastSolvedAt    *time.Time `json:"lastSolvedAt,omitempty"`
	TimeTaken       int        `json:"timeTaken"`
	DifficultyLevel string     `json:"difficulty_level"`
	QuestionType    string     `json:"question_type"`
	XPGained        int        `json:"xpGained"`
}

// BulkProgressRequest represents the request to get bulk progress
type BulkProgressRequest struct {
	IDs []string `json:"ids" binding:"required"`
}
