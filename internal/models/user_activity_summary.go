package models

import "time"

// UserActivitySummary represents a user's activity summary with daily entries
type UserActivitySummary struct {
	ID              string          `bson:"_id,omitempty" json:"id,omitempty"`
	UserID          string          `bson:"user_id" json:"userId"`
	CreatedAt       time.Time       `bson:"created_at" json:"createdAt"`
	UpdatedAt       time.Time       `bson:"updated_at" json:"updatedAt"`
	DailyActivities []DailyActivity `bson:"daily_activities" json:"dailyActivities"`
}

// DailyActivity represents activity data for a single day
type DailyActivity struct {
	Timestamp          time.Time `bson:"timestamp" json:"timestamp"` // Day start timestamp
	QuestionsSolved    int       `bson:"questions_solved" json:"questionsSolved"`
	QuestionsAttempted int       `bson:"questions_attempted" json:"questionsAttempted"` // Total attempts (solved + failed)
	QuestionIDs        []string  `bson:"question_ids" json:"questionIds"`               // Unique question IDs solved on this day
	TotalXP            int       `bson:"total_xp" json:"totalXP"`
	TotalActivities    int       `bson:"total_activities" json:"totalActivities"` // All activities (questions, discussions, etc.)

	// Word-related activities
	WordsMarkedKnown    int      `bson:"words_marked_known" json:"wordsMarkedKnown"`       // Words marked as known today
	WordsMarkedPractice int      `bson:"words_marked_practice" json:"wordsMarkedPractice"` // Words marked for practice today
	WordsViewed         int      `bson:"words_viewed" json:"wordsViewed"`                  // Words viewed today
	WordIDs             []string `bson:"word_ids" json:"wordIds"`                          // Unique word IDs interacted with today

	CreatedAt time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time `bson:"updated_at" json:"updatedAt"`
}

// ActivityCalendarResponse represents the activity calendar response for frontend
type ActivityCalendarResponse struct {
	Date  string `json:"date"`
	Count int    `json:"count"` // Number of questions solved
	Level int    `json:"level"` // 0-4 for GitHub-style activity visualization
}

// WordStatistics represents aggregated word learning statistics
type WordStatistics struct {
	TotalWordsViewed         int              `json:"totalWordsViewed"`
	TotalWordsMarkedKnown    int              `json:"totalWordsMarkedKnown"`
	TotalWordsMarkedPractice int              `json:"totalWordsMarkedPractice"`
	UniqueWordsCount         int              `json:"uniqueWordsCount"`
	UniqueWordsViewed        map[string]bool  `json:"-"` // Internal use, not serialized
	RecentDays               int              `json:"recentDays"`
	DailyStats               []DailyWordStats `json:"dailyStats"`
}

// DailyWordStats represents word activity stats for a single day
type DailyWordStats struct {
	Date                string `json:"date"`
	WordsViewed         int    `json:"wordsViewed"`
	WordsMarkedKnown    int    `json:"wordsMarkedKnown"`
	WordsMarkedPractice int    `json:"wordsMarkedPractice"`
}
