package services

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
	"verbalforge-backend/internal/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// Register registers a new user
func (s *AuthService) Register(req *models.CreateUserRequest) (*models.User, string, error) {
	// Create user
	user, err := s.userRepo.Create(req)
	if err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Username, s.jwtSecret)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return user, token, nil
}

// Login authenticates a user
func (s *AuthService) Login(req *models.LoginRequest) (*models.User, string, error) {
	var user *models.User
	var err error

	// Check if identifier is an email or username
	if utils.IsEmail(req.Identifier) {
		user, err = s.userRepo.FindByEmail(req.Identifier)
	} else {
		user, err = s.userRepo.FindByUsername(req.Identifier)
	}

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", errors.New("invalid credentials")
		}
		return nil, "", errors.New("database error")
	}

	// Validate password
	if !s.userRepo.ValidatePassword(user, req.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Username, s.jwtSecret)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	// Update login streak
	s.userRepo.UpdateStreak(user.ID)

	return user, token, nil
}

// GetCurrentUser retrieves the current authenticated user
func (s *AuthService) GetCurrentUser(userID string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(userID string, req *models.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Validate current password
	if !s.userRepo.ValidatePassword(user, req.OldPassword) {
		return errors.New("current password is incorrect")
	}

	// Check if new password is same as current
	if req.OldPassword == req.NewPassword {
		return errors.New("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := s.userRepo.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password
	updates := map[string]interface{}{
		"password": hashedPassword,
	}
	return s.userRepo.Update(userID, updates)
}

// DeleteAccount deletes a user account
func (s *AuthService) DeleteAccount(userID, password string) error {
	// Get user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Validate password
	if !s.userRepo.ValidatePassword(user, password) {
		return errors.New("incorrect password")
	}

	// Delete user
	return s.userRepo.Delete(userID)
}
