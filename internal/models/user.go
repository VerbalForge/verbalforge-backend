package models

import "time"

// ProfilePreferences represents user profile privacy settings
type ProfilePreferences struct {
	Visibility  bool `json:"visibility" bson:"visibility"`
	Progress    bool `json:"progress" bson:"progress"`
	Leaderboard bool `json:"leaderboard" bson:"leaderboard"`
}

// UserPreferences represents user application preferences
type UserPreferences struct {
	Profile ProfilePreferences `json:"profile" bson:"profile"`
	Theme   string             `json:"theme" bson:"theme"`
}

// User represents a user in the system
type User struct {
	ID            string          `json:"id" bson:"_id,omitempty"`
	Name          string          `json:"name" bson:"name"`
	Username      string          `json:"username" bson:"username"`
	Email         string          `json:"email" bson:"email"`
	Phone         string          `json:"phone" bson:"phone"`
	Bio           string          `json:"bio" bson:"bio"`
	Password      string          `json:"-" bson:"password"`
	Preferences   UserPreferences `json:"preferences" bson:"preferences"`
	TotalXP       int             `json:"totalXP" bson:"total_xp"`
	TotalSolved   int             `json:"totalSolved" bson:"total_solved"`
	Rank          int             `json:"rank" bson:"rank"`
	CurrentStreak int             `json:"currentStreak" bson:"current_streak"`
	LongestStreak int             `json:"longestStreak" bson:"longest_streak"`
	LastLogin     time.Time       `json:"lastLogin" bson:"last_login"`
	TotalAttempts int             `json:"totalAttempts" bson:"total_attempts"`
	ProfileViews  int             `json:"profileViews" bson:"profile_views"`
	CreatedAt     time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt" bson:"updatedAt"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,min=10"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents the login request
type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // Email or Username
	Password   string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents the request to update user profile
type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2"`
	Bio   string `json:"bio"`
	Phone string `json:"phone" binding:"omitempty,min=10"`
}

// ChangePasswordRequest represents the request to change password
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// UpdatePreferenceRequest represents the request to update a preference
type UpdatePreferenceRequest struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value" binding:"required"`
}

// UpdateThemeRequest represents the request to update theme
type UpdateThemeRequest struct {
	Theme string `json:"theme" binding:"required,oneof=light dark"`
}

// LeaderboardEntry represents a user entry in the leaderboard
type LeaderboardEntry struct {
	UserID        string `json:"userId" bson:"_id"`
	Username      string `json:"username" bson:"username"`
	Name          string `json:"name" bson:"name"`
	TotalXP       int    `json:"totalXP" bson:"total_xp"`
	TotalSolved   int    `json:"totalSolved" bson:"total_solved"`
	Rank          int    `json:"rank" bson:"rank"`
	CurrentStreak int    `json:"currentStreak" bson:"current_streak"`
}

// RecentActivityItem represents a recent activity entry for user profile
type RecentActivityItem struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	DifficultyLevel string `json:"difficulty_level"`
	QuestionType    string `json:"question_type"`
	SolvedAt        string `json:"solvedAt"`
	XPGained        int    `json:"xpGained"`
	PassageID       string `json:"passageId,omitempty"`
}

// UserProfileResponse represents the complete user profile response
type UserProfileResponse struct {
	User *User `json:"user"`
}
