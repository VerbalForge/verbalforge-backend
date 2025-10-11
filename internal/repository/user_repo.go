package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"verbalforge-backend/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	collection *mongo.Collection
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection("users"),
	}
}

// Create creates a new user
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	ctx := context.Background()

	// Check if user already exists by email
	existingUser := &models.User{}
	err := r.collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(existingUser)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Check if username already exists
	err = r.collection.FindOne(ctx, bson.M{"username": req.Username}).Decode(existingUser)
	if err == nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate UUID for the user
	userID := uuid.New().String()

	// Create user
	user := &models.User{
		ID:       userID,
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Preferences: models.UserPreferences{
			Profile: models.ProfilePreferences{
				Visibility:  true,
				Progress:    true,
				Leaderboard: true,
			},
			Theme:    "light",
			Timezone: req.Timezone, // Save user's timezone from registration
		},
		TotalXP:       0,
		TotalSolved:   0,
		Rank:          0,
		CurrentStreak: 0,
		LongestStreak: 0,
		LastLogin:     time.Time{},
		TotalAttempts: 0,
		ProfileViews:  0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := r.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	err := r.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id string) (*models.User, error) {
	user := &models.User{}
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(id string, updates bson.M) error {
	updates["updatedAt"] = time.Now()
	_, err := r.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

// Delete deletes a user
func (r *UserRepository) Delete(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

// ValidatePassword validates a user's password
func (r *UserRepository) ValidatePassword(user *models.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// HashPassword hashes a password
func (r *UserRepository) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// IncrementXP increments user's total XP and total attempts
func (r *UserRepository) IncrementXP(userID string, xpGained int) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$inc": bson.M{
			"total_xp":       xpGained,
			"total_attempts": 1,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

// IncrementSolved increments user's total solved count, XP, and attempts
func (r *UserRepository) IncrementSolved(userID string, xpGained int) error {
	ctx := context.Background()
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$inc": bson.M{
			"total_xp":       xpGained,
			"total_solved":   1,
			"total_attempts": 1,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Update all users' ranks globally after XP change
	// This ensures the leaderboard is always accurate
	go func() {
		// Run in background to not block the response
		_ = r.UpdateAllRanks()
	}()

	return nil
}

// IncrementAttempts increments only the total attempts counter
func (r *UserRepository) IncrementAttempts(userID string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$inc": bson.M{
			"total_attempts": 1,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

// IncrementProfileViews increments profile view counter
func (r *UserRepository) IncrementProfileViews(userID string) error {
	filter := bson.M{"_id": userID}
	update := bson.M{
		"$inc": bson.M{
			"profile_views": 1,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

// UpdateStreak updates user's login streak
func (r *UserRepository) UpdateStreak(userID string) error {
	user, err := r.FindByID(userID)
	if err != nil {
		return err
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	lastLoginDate := time.Date(user.LastLogin.Year(), user.LastLogin.Month(), user.LastLogin.Day(), 0, 0, 0, 0, user.LastLogin.Location())

	// If already logged in today, no update needed
	if lastLoginDate.Equal(today) {
		return nil
	}

	yesterday := today.AddDate(0, 0, -1)
	newStreak := 1

	// If last login was yesterday, increment streak
	if lastLoginDate.Equal(yesterday) {
		newStreak = user.CurrentStreak + 1
	}

	// Update longest streak if necessary
	longestStreak := user.LongestStreak
	if newStreak > longestStreak {
		longestStreak = newStreak
	}

	filter := bson.M{"_id": userID}
	update := bson.M{
		"$set": bson.M{
			"current_streak": newStreak,
			"longest_streak": longestStreak,
			"last_login":     today,
			"updatedAt":      time.Now(),
		},
	}

	_, err = r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

// UpdateAllRanks recalculates ranks for all users based on total XP
func (r *UserRepository) UpdateAllRanks() error {
	ctx := context.Background()

	// Get all users sorted by total_xp (descending)
	opts := options.Find().SetSort(bson.D{{Key: "total_xp", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var allUsers []models.User
	if err = cursor.All(ctx, &allUsers); err != nil {
		return err
	}

	// Update ranks
	for i, user := range allUsers {
		rank := i + 1
		filter := bson.M{"_id": user.ID}
		update := bson.M{
			"$set": bson.M{
				"rank":      rank,
				"updatedAt": time.Now(),
			},
		}
		_, err := r.collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetLeaderboard retrieves top users by total XP
func (r *UserRepository) GetLeaderboard(limit int) ([]models.LeaderboardEntry, error) {
	ctx := context.Background()

	// Get top users by XP
	opts := options.Find().
		SetSort(bson.D{{Key: "total_xp", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{
			"_id":            1,
			"username":       1,
			"name":           1,
			"total_xp":       1,
			"total_solved":   1,
			"rank":           1,
			"current_streak": 1,
		})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var leaderboard []models.LeaderboardEntry
	if err = cursor.All(ctx, &leaderboard); err != nil {
		return nil, err
	}

	return leaderboard, nil
}
