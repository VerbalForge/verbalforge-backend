package services

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	userRepo         *repository.UserRepository
	userQuestionRepo *repository.UserQuestionRepository
	activityService  *UserActivityService
	statsService     *UserStatsService
}

// NewUserService creates a new user service
func NewUserService(
	userRepo *repository.UserRepository,
	userQuestionRepo *repository.UserQuestionRepository,
	activityService *UserActivityService,
) *UserService {
	// Create stats service internally
	statsService := NewUserStatsService(userRepo, userQuestionRepo)

	return &UserService{
		userRepo:         userRepo,
		userQuestionRepo: userQuestionRepo,
		activityService:  activityService,
		statsService:     statsService,
	}
}

// GetProfile retrieves user profile by ID
func (s *UserService) GetProfile(userID string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// GetProfileByUsername retrieves user profile by username
func (s *UserService) GetProfileByUsername(username string) (*models.UserProfileResponse, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &models.UserProfileResponse{
		User: user,
	}, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(userID string, req *models.UpdateProfileRequest) (*models.User, error) {
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	err := s.userRepo.Update(userID, updates)
	if err != nil {
		return nil, err
	}

	return s.userRepo.FindByID(userID)
}

// UpdatePreference updates a specific user preference
func (s *UserService) UpdatePreference(userID string, req *models.UpdatePreferenceRequest) error {
	updates := make(map[string]interface{})

	// Map preference keys to their correct nested path
	switch req.Key {
	case "visibility", "progress", "leaderboard":
		updates["preferences.profile."+req.Key] = req.Value
	default:
		updates["preferences."+req.Key] = req.Value
	}

	return s.userRepo.Update(userID, updates)
}

// UpdateTheme updates user theme preference
func (s *UserService) UpdateTheme(userID string, req *models.UpdateThemeRequest) error {
	updates := make(map[string]interface{})
	updates["preferences.theme"] = req.Theme

	return s.userRepo.Update(userID, updates)
}

// IncrementProfileView increments profile view count
func (s *UserService) IncrementProfileView(userID string) error {
	return s.userRepo.IncrementProfileViews(userID)
}

// GetUserStats retrieves comprehensive user statistics
func (s *UserService) GetUserStats(userID string) (*models.UserStats, error) {
	return s.statsService.GetUserStats(userID)
}

// GetUserStatsByUsername retrieves user stats by username
func (s *UserService) GetUserStatsByUsername(username string) (*models.UserStats, error) {
	return s.statsService.GetUserStatsByUsername(username)
}

// GetLeaderboard retrieves the leaderboard
func (s *UserService) GetLeaderboard(limit int) ([]models.LeaderboardEntry, error) {
	return s.statsService.GetLeaderboard(limit)
}

// UpdateRanks recalculates and updates all user ranks
func (s *UserService) UpdateRanks() error {
	return s.statsService.UpdateRanks()
}

// GetActivityCalendar retrieves user activity calendar data
func (s *UserService) GetActivityCalendar(userID string, days int, timezone string) ([]models.ActivityCalendarResponse, error) {
	return s.activityService.GetActivityCalendar(userID, days, timezone)
}

// GetRecentActivity retrieves recent activity for a user by username
func (s *UserService) GetRecentActivity(username string, limit int) ([]models.RecentActivityItem, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.getRecentActivity(user.ID, limit), nil
}

// getRecentActivity retrieves recent solved questions for user profile
func (s *UserService) getRecentActivity(userID string, limit int) []models.RecentActivityItem {
	// Get recent activities from user_activity collection
	activities, err := s.activityService.GetRecentActivities(userID, limit)
	if err != nil {
		return []models.RecentActivityItem{}
	}

	recentActivity := make([]models.RecentActivityItem, 0)
	for _, activity := range activities {
		// Only include question activities where the question was solved
		if activity.ActivityType != models.ActivityQuestion {
			continue
		}

		// Check if the question was solved (not just attempted)
		solved, ok := activity.Metadata["solved"].(bool)
		if !ok || !solved {
			continue
		}

		// Extract metadata for question solved
		questionID, _ := activity.Metadata["question_id"].(string)
		title, _ := activity.Metadata["title"].(string)
		difficultyLevel, _ := activity.Metadata["difficulty_level"].(string)
		questionType, _ := activity.Metadata["question_type"].(string)
		passageID, _ := activity.Metadata["passage_id"].(string)

		// Extract XP from metadata
		xpGained := 0
		if xp, ok := activity.Metadata["xp_gained"]; ok {
			if xpInt, ok := xp.(int); ok {
				xpGained = xpInt
			} else if xpFloat, ok := xp.(float64); ok {
				xpGained = int(xpFloat)
			}
		}

		recentActivity = append(recentActivity, models.RecentActivityItem{
			ID:              questionID,
			Title:           title,
			DifficultyLevel: difficultyLevel,
			QuestionType:    questionType,
			SolvedAt:        activity.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			XPGained:        xpGained,
			PassageID:       passageID,
		})

		// Stop once we have enough solved questions
		if len(recentActivity) >= limit {
			break
		}
	}

	return recentActivity
}
