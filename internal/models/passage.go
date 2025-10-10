package models

import "time"

// PassageMetadata represents passage metadata
type PassageMetadata struct {
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	PublishedAt time.Time `json:"published_at" bson:"published_at"`
	BatchID     string    `json:"batch_id" bson:"batch_id"`
}

// Passage represents a reading comprehension passage
type Passage struct {
	ID          string          `json:"id" bson:"_id,omitempty"`
	Passage     string          `json:"passage" bson:"passage"`
	Source      string          `json:"source" bson:"source"`
	Title       string          `json:"title" bson:"title"`
	Difficulty  string          `json:"difficulty" bson:"difficulty"`
	Type        string          `json:"type" bson:"type"`
	QuestionIDs []string        `json:"question_ids" bson:"question_ids"`
	Metadata    PassageMetadata `json:"metadata" bson:"metadata"`
}

// PartialPassage contains minimal passage information for listing
type PartialPassage struct {
	ID          string   `json:"id" bson:"_id,omitempty"`
	Passage     string   `json:"passage" bson:"passage"`
	Title       string   `json:"title" bson:"title"`
	Difficulty  string   `json:"difficulty" bson:"difficulty"`
	QuestionIDs []string `json:"question_ids" bson:"question_ids"`
	CreatedAt   string   `json:"created_at" bson:"created_at"`
}

// PassagesResponse represents paginated passages response with cursor-based pagination
type PassagesResponse struct {
	Passages       []PartialPassage `json:"passages"`
	Total          int64            `json:"total"`
	Limit          int              `json:"limit"`
	NextCursor     string           `json:"nextCursor,omitempty"`     // Continuation token for next page
	PreviousCursor string           `json:"previousCursor,omitempty"` // Continuation token for previous page
	HasMore        bool             `json:"hasMore"`                  // Indicates if more results exist
}
