package services

import (
	"time"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// UserActivitySummaryService handles user activity summary business logic
type UserActivitySummaryService struct {
	summaryRepo *repository.UserActivitySummaryRepository
}

// NewUserActivitySummaryService creates a new user activity summary service
func NewUserActivitySummaryService(summaryRepo *repository.UserActivitySummaryRepository) *UserActivitySummaryService {
	return &UserActivitySummaryService{
		summaryRepo: summaryRepo,
	}
}

// UpdateDailySummary updates the daily summary for question activities
// Uses timezone to bucket activities by the user's local date
func (s *UserActivitySummaryService) UpdateDailySummary(userID string, questionSolved bool, xpGained int, questionID string, timezone string) error {
	// Get current time in user's timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
	}

	// Get today's date in user's timezone
	now := time.Now().In(loc)
	todayInUserTZ := now.Format("2006-01-02")

	return s.summaryRepo.UpsertDailyActivity(userID, todayInUserTZ, questionSolved, xpGained, questionID)
}

// UpdateDailySummaryForWord updates the daily summary for word activities
// Uses timezone to bucket activities by the user's local date
func (s *UserActivitySummaryService) UpdateDailySummaryForWord(userID string, actionType models.WordActionType, wordID string, timezone string) error {
	// Get current time in user's timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
	}

	// Get today's date in user's timezone
	now := time.Now().In(loc)
	todayInUserTZ := now.Format("2006-01-02")

	return s.summaryRepo.UpsertWordActivity(userID, todayInUserTZ, actionType, wordID)
}

// GetUserSummary retrieves the complete activity summary for a user
func (s *UserActivitySummaryService) GetUserSummary(userID string) (*models.UserActivitySummary, error) {
	return s.summaryRepo.GetUserSummary(userID)
}

// GetDailySummary retrieves the daily activity for a specific date
func (s *UserActivitySummaryService) GetDailySummary(userID, date string) (*models.DailyActivity, error) {
	return s.summaryRepo.GetDailyActivity(userID, date)
}

// GetSummariesByDateRange retrieves daily activities within a date range
func (s *UserActivitySummaryService) GetSummariesByDateRange(userID, startDate, endDate string) ([]models.DailyActivity, error) {
	return s.summaryRepo.GetActivitiesByDateRange(userID, startDate, endDate)
}

// GetActivityCalendar retrieves activity calendar data for visualization
func (s *UserActivitySummaryService) GetActivityCalendar(userID string, days int, timezone string) ([]models.ActivityCalendarResponse, error) {
	activities, err := s.summaryRepo.GetActivityCalendar(userID, days, timezone)
	if err != nil {
		return nil, err
	}

	// Convert to calendar response format
	calendar := make([]models.ActivityCalendarResponse, 0, len(activities))
	for _, activity := range activities {
		level := s.calculateActivityLevel(activity.QuestionsSolved)
		calendar = append(calendar, models.ActivityCalendarResponse{
			Date:  activity.Timestamp.Format("2006-01-02"),
			Count: activity.QuestionsSolved,
			Level: level,
		})
	}

	return calendar, nil
}

// calculateActivityLevel calculates the activity level (0-4) based on questions solved
func (s *UserActivitySummaryService) calculateActivityLevel(questionsSolved int) int {
	switch {
	case questionsSolved == 0:
		return 0
	case questionsSolved <= 2:
		return 1
	case questionsSolved <= 4:
		return 2
	case questionsSolved <= 6:
		return 3
	default:
		return 4
	}
}

// GetRecentActivities retrieves the most recent daily activities
func (s *UserActivitySummaryService) GetRecentActivities(userID string, limit int) ([]models.DailyActivity, error) {
	return s.summaryRepo.GetRecentActivities(userID, limit)
}

// CleanupOldActivities deletes activities older than the specified number of days for a user
func (s *UserActivitySummaryService) CleanupOldActivities(userID string, days int) error {
	return s.summaryRepo.DeleteOldActivities(userID, days)
}

// GetWordStatistics retrieves aggregated word learning statistics for a user
func (s *UserActivitySummaryService) GetWordStatistics(userID string, days int) (*models.WordStatistics, error) {
	summary, err := s.summaryRepo.GetUserSummary(userID)
	if err != nil {
		return nil, err
	}

	// Calculate statistics from daily activities
	stats := &models.WordStatistics{
		TotalWordsViewed:         0,
		TotalWordsMarkedKnown:    0,
		TotalWordsMarkedPractice: 0,
		UniqueWordsViewed:        make(map[string]bool),
		RecentDays:               days,
		DailyStats:               []models.DailyWordStats{},
	}

	// Limit to recent days if specified
	cutoffDate := time.Now().AddDate(0, 0, -days)

	for _, activity := range summary.DailyActivities {
		// Skip if before cutoff date and days is specified
		if days > 0 && activity.Timestamp.Before(cutoffDate) {
			continue
		}

		stats.TotalWordsViewed += activity.WordsViewed
		stats.TotalWordsMarkedKnown += activity.WordsMarkedKnown
		stats.TotalWordsMarkedPractice += activity.WordsMarkedPractice

		// Track unique words
		for _, wordID := range activity.WordIDs {
			stats.UniqueWordsViewed[wordID] = true
		}

		// Add to daily stats
		stats.DailyStats = append(stats.DailyStats, models.DailyWordStats{
			Date:                activity.Timestamp.Format("2006-01-02"),
			WordsViewed:         activity.WordsViewed,
			WordsMarkedKnown:    activity.WordsMarkedKnown,
			WordsMarkedPractice: activity.WordsMarkedPractice,
		})
	}

	stats.UniqueWordsCount = len(stats.UniqueWordsViewed)

	return stats, nil
}
