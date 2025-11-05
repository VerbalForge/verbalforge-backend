package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
	"verbalforge-backend/internal/utils"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo           *repository.UserRepository
	jwtSecret          string
	googleClientID     string
	googleClientSecret string
	frontendURL        string
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, jwtSecret, googleClientID, googleClientSecret, frontendURL string) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		jwtSecret:          jwtSecret,
		googleClientID:     googleClientID,
		googleClientSecret: googleClientSecret,
		frontendURL:        frontendURL,
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

// RefreshToken generates a new JWT token for an authenticated user
func (s *AuthService) RefreshToken(userID, email, username string) (string, error) {
	token, err := utils.GenerateJWT(userID, email, username, s.jwtSecret)
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return token, nil
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

// GoogleLogin handles Google OAuth login/signup
func (s *AuthService) GoogleLogin(ctx context.Context, authCode string) (*models.User, string, error) {
	idToken, err := s.exchangeCodeForIDToken(authCode)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	googleUser, err := s.verifyGoogleToken(idToken)
	if err != nil {
		return nil, "", fmt.Errorf("invalid Google token: %w", err)
	}

	user, err := s.userRepo.FindByOAuthID("google", googleUser.Sub)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, "", fmt.Errorf("database error: %w", err)
	}

	if err == mongo.ErrNoDocuments {
		existingUser, err := s.userRepo.FindByEmail(googleUser.Email)
		if err != nil && err != mongo.ErrNoDocuments {
			return nil, "", fmt.Errorf("database error: %w", err)
		}

		if existingUser != nil {
			return nil, "", errors.New("an account with this email already exists - please login with your email and password")
		}

		username := s.generateUsernameFromEmail(googleUser.Email)

		user = &models.User{
			Name:          googleUser.Name,
			Username:      username,
			Email:         googleUser.Email,
			Phone:         "",
			OAuthProvider: "google",
			OAuthID:       googleUser.Sub,
			ProfilePicURL: googleUser.Picture,
			Password:      "",
			IsAdmin:       false,
			Preferences: models.UserPreferences{
				Profile: models.ProfilePreferences{
					Visibility:  true,
					Progress:    true,
					Leaderboard: true,
				},
				Theme:    "light",
				Timezone: "UTC",
			},
			TotalXP:       0,
			TotalSolved:   0,
			Rank:          0,
			CurrentStreak: 0,
			LongestStreak: 0,
			TotalAttempts: 0,
			ProfileViews:  0,
			LastLogin:     time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		createdUser, err := s.userRepo.CreateOAuthUser(user)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create user: %w", err)
		}
		user = createdUser
	}

	s.userRepo.UpdateStreak(user.ID)

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Username, s.jwtSecret)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return user, token, nil
}

func (s *AuthService) exchangeCodeForIDToken(code string) (string, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", s.googleClientID)
	data.Set("client_secret", s.googleClientSecret)
	data.Set("redirect_uri", s.frontendURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResponse struct {
		IDToken      string `json:"id_token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResponse.IDToken, nil
}

func (s *AuthService) verifyGoogleToken(idToken string) (*models.GoogleUserInfo, error) {
	url := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid token: %s", string(body))
	}

	var userInfo models.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// generateUsernameFromEmail generates a unique username from an email address
func (s *AuthService) generateUsernameFromEmail(email string) string {
	// Extract the part before @
	parts := strings.Split(email, "@")
	if len(parts) == 0 {
		return fmt.Sprintf("user_%d", time.Now().Unix())
	}

	baseUsername := strings.ToLower(parts[0])
	// Remove any non-alphanumeric characters except underscore
	baseUsername = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, baseUsername)

	// Try the base username first
	username := baseUsername
	counter := 1

	// Keep trying until we find a unique username
	for {
		_, err := s.userRepo.FindByUsername(username)
		if err == mongo.ErrNoDocuments {
			// Username is available
			return username
		}
		// Username exists, try with a number suffix
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++
	}
}

// ForgotPassword initiates the password reset process by generating a reset token
func (s *AuthService) ForgotPassword(email string) (string, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Don't reveal if email exists or not for security
			return "", nil
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	// Check if user is OAuth-only (no password)
	if user.Password == "" && user.OAuthProvider != "" {
		return "", errors.New("this account uses Google login - password reset not available")
	}

	// Generate reset token (random secure token)
	resetToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Set token expiry (1 hour from now)
	expiry := time.Now().Add(1 * time.Hour)

	// Save reset token to database
	if err := s.userRepo.SetResetToken(user.ID, resetToken, expiry); err != nil {
		return "", fmt.Errorf("failed to save reset token: %w", err)
	}

	// Return the token (in production, this should be sent via email)
	// For now, we'll return it so the frontend can use it
	return resetToken, nil
}

// VerifyResetToken verifies if a password reset token is valid
func (s *AuthService) VerifyResetToken(token string) (*models.User, error) {
	user, err := s.userRepo.FindByResetToken(token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid or expired reset token")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return user, nil
}

// ResetPassword resets the user's password using a valid reset token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// Verify token and get user
	user, err := s.VerifyResetToken(token)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := s.userRepo.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	updates := map[string]interface{}{
		"password": hashedPassword,
	}
	if err := s.userRepo.Update(user.ID, updates); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear reset token
	if err := s.userRepo.ClearResetToken(user.ID); err != nil {
		return fmt.Errorf("failed to clear reset token: %w", err)
	}

	return nil
}
