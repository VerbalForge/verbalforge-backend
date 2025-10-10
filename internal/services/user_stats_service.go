package services

import (
	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// UserStatsService handles user statistics business logic
type UserStatsService struct {
	userRepo         *repository.UserRepository
	userQuestionRepo *repository.UserQuestionRepository
}

// NewUserStatsService creates a new user stats service
func NewUserStatsService(
	userRepo *repository.UserRepository,
	userQuestionRepo *repository.UserQuestionRepository,
) *UserStatsService {
	return &UserStatsService{
		userRepo:         userRepo,
		userQuestionRepo: userQuestionRepo,
	}
}

// GetUserStats retrieves comprehensive user statistics
func (s *UserStatsService) GetUserStats(userID string) (*models.UserStats, error) {
	// Get stats from user_questions (for breakdown by type and difficulty)
	stats, err := s.userQuestionRepo.CalculateUserStats(userID)
	if err != nil {
		return nil, err
	}

	// Get user data for additional fields
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Merge data - use users table as authoritative source for totals
	stats.TotalSolved = user.TotalSolved
	stats.TotalXP = user.TotalXP
	stats.TotalAttempts = user.TotalAttempts
	stats.Rank = user.Rank
	stats.CurrentStreak = user.CurrentStreak
	stats.LongestStreak = user.LongestStreak
	stats.LastLogin = user.LastLogin
	stats.ProfileViews = user.ProfileViews

	return stats, nil
}

// GetUserStatsByUsername retrieves user stats by username
func (s *UserStatsService) GetUserStatsByUsername(username string) (*models.UserStats, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err
	}

	return s.GetUserStats(user.ID)
}

// GetLeaderboard retrieves the leaderboard
func (s *UserStatsService) GetLeaderboard(limit int) ([]models.LeaderboardEntry, error) {
	return s.userRepo.GetLeaderboard(limit)
}

// UpdateRanks recalculates and updates all user ranks
func (s *UserStatsService) UpdateRanks() error {
	return s.userRepo.UpdateAllRanks()
}
