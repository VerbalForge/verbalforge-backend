package models

import "time"

// PracticeItemType represents the type of practice item
type PracticeItemType string

const (
	PracticeItemTypeQuestion PracticeItemType = "question"
	PracticeItemTypePassage  PracticeItemType = "passage"
)

// PracticeItem represents a normalized practice item that can be either a question or passage
type PracticeItem struct {
	ID         string           `json:"id"`
	Type       PracticeItemType `json:"type"`
	Title      string           `json:"title"`
	Difficulty string           `json:"difficulty"`
	CreatedAt  time.Time        `json:"created_at"`

	// Question-specific fields (nil if type is passage)
	QuestionText *string `json:"question_text,omitempty"`
	QuestionType *string `json:"question_type,omitempty"`
	Topic        *string `json:"topic,omitempty"`

	// Passage-specific fields (nil if type is question)
	PassagePreview *string  `json:"passage_preview,omitempty"`
	QuestionIDs    []string `json:"question_ids,omitempty"`
	Source         *string  `json:"source,omitempty"`
}

// PracticeResponse represents the paginated response for practice items
type PracticeResponse struct {
	Items      []PracticeItem `json:"items"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// NewPracticeItemFromQuestion creates a PracticeItem from a PartialQuestion
func NewPracticeItemFromQuestion(q PartialQuestion) PracticeItem {
	createdAt, err := time.Parse(time.RFC3339, q.Metadata.CreatedAt)
	if err != nil {
		// If parsing fails, use zero time (will appear at end when sorted descending)
		createdAt = time.Time{}
	}

	// Create a title from question text (first 100 chars)
	title := q.QuestionText
	if len(title) > 100 {
		title = title[:97] + "..."
	}

	return PracticeItem{
		ID:           q.ID,
		Type:         PracticeItemTypeQuestion,
		Title:        title,
		Difficulty:   q.DifficultyLevel,
		CreatedAt:    createdAt,
		QuestionText: &q.QuestionText,
		QuestionType: &q.QuestionType,
		Topic:        &q.Topic,
	}
}

// NewPracticeItemFromPassage creates a PracticeItem from a PartialPassage
func NewPracticeItemFromPassage(p PartialPassage) PracticeItem {
	createdAt, err := time.Parse(time.RFC3339, p.Metadata.CreatedAt)
	if err != nil {
		// If parsing fails, use zero time (will appear at end when sorted descending)
		createdAt = time.Time{}
	}

	// Create a preview of the passage (first 150 chars)
	preview := p.Passage
	if len(preview) > 150 {
		preview = preview[:147] + "..."
	}

	title := p.Title
	if title == "" {
		// If no title, use the preview
		title = preview
		if len(title) > 100 {
			title = title[:97] + "..."
		}
	}

	return PracticeItem{
		ID:             p.ID,
		Type:           PracticeItemTypePassage,
		Title:          title,
		Difficulty:     p.Difficulty,
		CreatedAt:      createdAt,
		PassagePreview: &preview,
		QuestionIDs:    p.QuestionIDs,
	}
}
