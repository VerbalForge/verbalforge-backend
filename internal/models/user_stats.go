package models

import "time"

// UserStats represents aggregated user statistics
type UserStats struct {
	UserID        string    `json:"userId"`
	TotalSolved   int       `json:"totalSolved"`
	EasySolved    int       `json:"easySolved"`
	MediumSolved  int       `json:"mediumSolved"`
	HardSolved    int       `json:"hardSolved"`
	TCSolved      int       `json:"tcSolved"`
	SESolved      int       `json:"seSolved"`
	RCSolved      int       `json:"rcSolved"`
	TotalXP       int       `json:"totalXp"`
	Rank          int       `json:"rank"`
	CurrentStreak int       `json:"currentStreak"`
	LongestStreak int       `json:"longestStreak"`
	LastLogin     time.Time `json:"lastLogin"`
	TotalAttempts int       `json:"totalAttempts"`
	ProfileViews  int       `json:"profileViews"`
}
