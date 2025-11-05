package models

import "time"

// Choice represents a question choice
type Choice struct {
	Option    string `json:"option" bson:"option"`
	Blank     int    `json:"blank" bson:"blank"`
	IsCorrect bool   `json:"is_correct" bson:"is_correct"`
	Reasoning string `json:"reasoning" bson:"reasoning"`
}

// QuestionMetadata represents question metadata
type QuestionMetadata struct {
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty" bson:"published_at,omitempty"`
	BatchID     string     `json:"batch_id" bson:"batch_id"`
}

// Question represents a question in the system
type Question struct {
	ID              string           `json:"id" bson:"_id,omitempty"`
	QuestionType    string           `json:"question_type" bson:"question_type"`
	DifficultyLevel string           `json:"difficulty_level" bson:"difficulty_level"`
	Topic           string           `json:"topic" bson:"topic"`
	QuestionText    string           `json:"question_text" bson:"question_text"`
	PassageID       string           `json:"passage_id,omitempty" bson:"passage_id,omitempty"` // Optional, only for RC questions
	Choices         []Choice         `json:"choices" bson:"choices"`
	Metadata        QuestionMetadata `json:"metadata" bson:"metadata"`
}

// PartialQuestion contains minimal question information for listing
type PartialQuestion struct {
	ID              string `json:"id" bson:"_id,omitempty"`
	QuestionType    string `json:"question_type" bson:"question_type"`
	DifficultyLevel string `json:"difficulty_level" bson:"difficulty_level"`
	Topic           string `json:"topic" bson:"topic"`
	QuestionText    string `json:"question_text" bson:"question_text"`
	Metadata        struct {
		CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
		PublishedAt *time.Time `json:"published_at,omitempty" bson:"published_at,omitempty"`
	} `json:"metadata" bson:"metadata"`
}

// QuestionsResponse represents paginated questions response with cursor-based pagination
type QuestionsResponse struct {
	Questions      []PartialQuestion `json:"questions"`
	Total          int64             `json:"total"`
	Limit          int               `json:"limit"`
	NextCursor     string            `json:"nextCursor,omitempty"`     // Continuation token for next page
	PreviousCursor string            `json:"previousCursor,omitempty"` // Continuation token for previous page
	HasMore        bool              `json:"hasMore"`                  // Indicates if more results exist
}
