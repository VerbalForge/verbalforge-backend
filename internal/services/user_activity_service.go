package services

import (
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// UserActivityService handles user activity business logic
type UserActivityService struct {
	activityRepo   *repository.UserActivityRepository
	summaryService *UserActivitySummaryService
}

// NewUserActivityService creates a new user activity service
func NewUserActivityService(activityRepo *repository.UserActivityRepository, summaryService *UserActivitySummaryService) *UserActivityService {
	return &UserActivityService{
		activityRepo:   activityRepo,
		summaryService: summaryService,
	}
}

// LogActivity records a generic user activity event
func (s *UserActivityService) LogActivity(userID string, activityType models.ActivityType, metadata map[string]interface{}) error {
	return s.activityRepo.LogActivity(userID, activityType, metadata)
}

// LogQuestionActivity logs a question activity (attempted or solved)
func (s *UserActivityService) LogQuestionActivity(userID string, meta models.QuestionActivityMetadata) error {
	metadataMap, err := models.ToMetadataMap(meta)
	if err != nil {
		return err
	}

	// Log the activity
	if err := s.activityRepo.LogActivity(userID, models.ActivityQuestion, metadataMap); err != nil {
		return err
	}

	// Update daily summary (if summary service is available)
	if s.summaryService != nil {
		_ = s.summaryService.UpdateDailySummary(userID, meta.Solved, meta.XPGained, meta.QuestionID)
		// Don't fail the activity log if summary update fails
	}

	return nil
}

// LogDiscussionActivity logs a discussion activity (created, updated, comment added, etc.)
func (s *UserActivityService) LogDiscussionActivity(userID string, meta models.DiscussionActivityMetadata) error {
	metadataMap, err := models.ToMetadataMap(meta)
	if err != nil {
		return err
	}

	// All discussion activities use the same ActivityDiscussion type
	// The metadata.ActionType field distinguishes between different actions
	return s.activityRepo.LogActivity(userID, models.ActivityDiscussion, metadataMap)
}

// LogProfileViewed logs a profile viewed activity
func (s *UserActivityService) LogProfileViewed(userID string, meta models.ProfileViewedMetadata) error {
	metadataMap, err := models.ToMetadataMap(meta)
	if err != nil {
		return err
	}
	return s.activityRepo.LogActivity(userID, models.ActivityProfileViewed, metadataMap)
}

// LogStreakUpdated logs a streak updated activity
func (s *UserActivityService) LogStreakUpdated(userID string, meta models.StreakUpdatedMetadata) error {
	metadataMap, err := models.ToMetadataMap(meta)
	if err != nil {
		return err
	}
	return s.activityRepo.LogActivity(userID, models.ActivityStreakUpdated, metadataMap)
}

// GetActivityCalendar retrieves user activity calendar data
func (s *UserActivityService) GetActivityCalendar(userID string, days int, timezone string) ([]models.ActivityCalendarResponse, error) {
	if s.summaryService == nil {
		return []models.ActivityCalendarResponse{}, nil
	}
	return s.summaryService.GetActivityCalendar(userID, days, timezone)
}

// GetRecentActivities retrieves recent activities for a user
func (s *UserActivityService) GetRecentActivities(userID string, limit int) ([]models.UserActivity, error) {
	return s.activityRepo.GetRecentActivities(userID, limit)
}

// GetActivitiesByType retrieves activities of a specific type
func (s *UserActivityService) GetActivitiesByType(userID string, activityType models.ActivityType, limit int) ([]models.UserActivity, error) {
	return s.activityRepo.GetActivitiesByType(userID, activityType, limit)
}

// GetActivitiesByDateRange retrieves activities within a date range
func (s *UserActivityService) GetActivitiesByDateRange(userID string, startDate, endDate string) ([]models.UserActivity, error) {
	return s.activityRepo.GetActivitiesByDateRange(userID, startDate, endDate)
}

// GetDailySummary retrieves the daily activity for a specific date
func (s *UserActivityService) GetDailySummary(userID string, date string) (*models.DailyActivity, error) {
	if s.summaryService == nil {
		return nil, nil
	}
	return s.summaryService.GetDailySummary(userID, date)
}
