package services

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// UserActivityService handles user activity business logic
type UserActivityService struct {
	activityRepo   *repository.UserActivityRepository
	summaryService *UserActivitySummaryService
	userRepo       *repository.UserRepository
}

// NewUserActivityService creates a new user activity service
func NewUserActivityService(activityRepo *repository.UserActivityRepository, summaryService *UserActivitySummaryService, userRepo *repository.UserRepository) *UserActivityService {
	return &UserActivityService{
		activityRepo:   activityRepo,
		summaryService: summaryService,
		userRepo:       userRepo,
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
		// Get user's timezone preference
		timezone := "UTC" // Default
		if s.userRepo != nil {
			user, err := s.userRepo.FindByID(userID)
			if err == nil && user != nil && user.Preferences.Timezone != "" {
				timezone = user.Preferences.Timezone
			}
		}

		_ = s.summaryService.UpdateDailySummary(userID, meta.Solved, meta.XPGained, meta.QuestionID, timezone)
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

// LogWordActivity logs a word learning activity (marked known, practice, viewed, etc.)
func (s *UserActivityService) LogWordActivity(userID string, meta models.WordActivityMetadata) error {
	metadataMap, err := models.ToMetadataMap(meta)
	if err != nil {
		return err
	}

	// Log the activity
	if err := s.activityRepo.LogActivity(userID, models.ActivityWord, metadataMap); err != nil {
		return err
	}

	// Update daily summary for word activities
	if s.summaryService != nil {
		// Get user's timezone preference
		timezone := "UTC" // Default
		if s.userRepo != nil {
			user, err := s.userRepo.FindByID(userID)
			if err == nil && user != nil && user.Preferences.Timezone != "" {
				timezone = user.Preferences.Timezone
			}
		}

		_ = s.summaryService.UpdateDailySummaryForWord(userID, meta.ActionType, meta.WordID, timezone)
		// Don't fail the activity log if summary update fails
	}

	return nil
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

// GetUserActivityLogs returns detailed activity logs with filters
func (s *UserActivityService) GetUserActivityLogs(userID, questionType, startDate, endDate string, limit, skip int64) ([]models.UserActivity, int64, error) {
	filter := bson.M{"user_id": userID}

	if questionType != "" {
		filter["metadata.question_type"] = questionType
	}

	// Date range filter
	if startDate != "" || endDate != "" {
		dateFilter := bson.M{}
		if startDate != "" {
			start, err := time.Parse("2006-01-02", startDate)
			if err == nil {
				dateFilter["$gte"] = start
			}
		}
		if endDate != "" {
			end, err := time.Parse("2006-01-02", endDate)
			if err == nil {
				dateFilter["$lte"] = end.Add(24 * time.Hour) // Include full day
			}
		}
		if len(dateFilter) > 0 {
			filter["timestamp"] = dateFilter
		}
	}

	activities, err := s.activityRepo.FindAll(filter, limit, skip)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.activityRepo.Count(filter)
	if err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// GetUserAggregatedStats returns aggregated stats for a user
func (s *UserActivityService) GetUserAggregatedStats(userID string) (map[string]interface{}, error) {
	// Get user info
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"user_id":  userID,
		"username": user.Username,
		"email":    user.Email,
	}

	// Get activity summary by question type
	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID, "activity_type": models.ActivityQuestion}},
		{"$group": bson.M{
			"_id":             "$metadata.question_type",
			"total_attempted": bson.M{"$sum": 1},
			"total_correct": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$metadata.solved", true}},
					1,
					0,
				},
			}},
		}},
	}

	cursor, err := s.activityRepo.Aggregate(pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(nil)

	var typeSummary []bson.M
	if err := cursor.All(nil, &typeSummary); err != nil {
		return nil, err
	}

	// Calculate overall stats
	totalAttempted := 0
	totalCorrect := 0
	typeBreakdown := make(map[string]map[string]int)
	mostAttemptedType := ""
	maxAttempts := 0

	for _, summary := range typeSummary {
		questionType := summary["_id"].(string)
		attempted := int(summary["total_attempted"].(int32))
		correct := int(summary["total_correct"].(int32))

		totalAttempted += attempted
		totalCorrect += correct

		// Map backend types to frontend types
		var frontendType string
		switch questionType {
		case "text_completion_single", "text_completion_double", "text_completion_triple":
			frontendType = "TC"
		case "sentence_equivalence":
			frontendType = "SE"
		case "reading_comprehension_single", "reading_comprehension_multiple", "reading_comprehension_highlight":
			frontendType = "RC"
		default:
			frontendType = questionType
		}

		if typeBreakdown[frontendType] == nil {
			typeBreakdown[frontendType] = map[string]int{"attempted": 0, "correct": 0}
		}
		typeBreakdown[frontendType]["attempted"] += attempted
		typeBreakdown[frontendType]["correct"] += correct

		if attempted > maxAttempts {
			maxAttempts = attempted
			mostAttemptedType = frontendType
		}
	}

	// Calculate accuracy rate
	accuracyRate := 0.0
	if totalAttempted > 0 {
		accuracyRate = float64(totalCorrect) / float64(totalAttempted)
	}

	stats["total_attempted"] = totalAttempted
	stats["accuracy_rate"] = accuracyRate
	stats["most_attempted_type"] = mostAttemptedType
	stats["type_breakdown"] = typeBreakdown

	return stats, nil
}

// GetWordStatistics retrieves word learning statistics for a user
func (s *UserActivityService) GetWordStatistics(userID string, days int) (*models.WordStatistics, error) {
	if s.summaryService != nil {
		return s.summaryService.GetWordStatistics(userID, days)
	}
	return nil, nil
}
