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
func (s *UserActivitySummaryService) UpdateDailySummary(userID string, questionSolved bool, xpGained int, questionID string) error {
	today := time.Now().UTC().Format("2006-01-02")
	return s.summaryRepo.UpsertDailyActivity(userID, today, questionSolved, xpGained, questionID)
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
