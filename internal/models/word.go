package models

import "time"

// WordMeaning represents a single meaning/definition of a word
type WordMeaning struct {
	Definition string   `json:"definition" bson:"definition"`
	Examples   []string `json:"examples" bson:"examples"`
}

// Word represents a vocabulary word with all its details
type Word struct {
	ID            string        `json:"id" bson:"_id,omitempty"`
	Word          string        `json:"word" bson:"word"`
	Pronunciation string        `json:"pronunciation" bson:"pronunciation"`
	Meanings      []WordMeaning `json:"meanings" bson:"meanings"`
	Sources       []string      `json:"sources" bson:"sources"`
	Synonyms      []string      `json:"synonyms" bson:"synonyms"`
	Antonyms      []string      `json:"antonyms" bson:"antonyms"`
	CreatedAt     time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt" bson:"updatedAt"`
}

// WordFilters represents filters for word queries
type WordFilters struct {
	Sources  []string `json:"sources"`
	Search   string   `json:"search"`
	Known    *bool    `json:"known,omitempty"`    // Filter by known status
	Practice *bool    `json:"practice,omitempty"` // Filter by practice status
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
	Page     int      `json:"page"` // For pagination
}

// WordsResponse represents paginated word response
type WordsResponse struct {
	Words      []Word    `json:"words"`
	Total      int64     `json:"total"`
	Sources    []string  `json:"sources"` // Available sources for filtering
	HasMore    bool      `json:"hasMore"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	Page       int       `json:"page"`
	TotalPages int       `json:"totalPages"`
	UserWords  *UserWord `json:"userWords,omitempty"` // User's progress
}
