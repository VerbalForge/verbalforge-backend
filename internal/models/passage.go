package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// PassageMetadata represents passage metadata
type PassageMetadata struct {
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	PublishedAt time.Time `json:"published_at" bson:"published_at"`
	BatchID     string    `json:"batch_id" bson:"batch_id"`
}

// MarshalJSON customizes JSON encoding to ensure RFC3339 format with timezone
func (m PassageMetadata) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		PublishedAt string `json:"published_at"`
		BatchID     string `json:"batch_id"`
	}{
		CreatedAt:   m.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339),
		PublishedAt: m.PublishedAt.Format(time.RFC3339),
		BatchID:     m.BatchID,
	})
}

// UnmarshalBSON customizes BSON decoding to handle timestamps without timezone info
func (m *PassageMetadata) UnmarshalBSON(data []byte) error {
	type Alias struct {
		CreatedAt   interface{} `bson:"created_at"`
		UpdatedAt   interface{} `bson:"updated_at"`
		PublishedAt interface{} `bson:"published_at"`
		BatchID     string      `bson:"batch_id"`
	}

	var aux Alias
	if err := bson.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Helper function to parse timestamp from various formats
	parseTime := func(v interface{}) (time.Time, error) {
		switch val := v.(type) {
		case string:
			// Try multiple formats
			formats := []string{
				time.RFC3339,
				time.RFC3339Nano,
				"2006-01-02T15:04:05.999999", // Without timezone
				"2006-01-02T15:04:05",        // Without milliseconds and timezone
			}
			for _, format := range formats {
				if t, err := time.Parse(format, val); err == nil {
					return t, nil
				}
			}
			// If all parsing fails, return error from the first format
			return time.Parse(time.RFC3339, val)
		case time.Time:
			return val, nil
		default:
			return time.Time{}, nil
		}
	}

	var err error
	if m.CreatedAt, err = parseTime(aux.CreatedAt); err != nil {
		return err
	}
	if m.UpdatedAt, err = parseTime(aux.UpdatedAt); err != nil {
		return err
	}
	if m.PublishedAt, err = parseTime(aux.PublishedAt); err != nil {
		return err
	}
	m.BatchID = aux.BatchID

	return nil
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
	Metadata    struct {
		CreatedAt string `json:"created_at" bson:"created_at"`
	} `json:"metadata" bson:"metadata"`
	CreatedAt string `json:"-" bson:"-"` // Computed field for backward compatibility
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
