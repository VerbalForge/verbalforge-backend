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
