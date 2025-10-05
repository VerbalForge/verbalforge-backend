package models

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type ProfilePreferences struct {
	Visibility  bool `json:"visibility" bson:"visibility"`
	Progress    bool `json:"progress" bson:"progress"`
	Leaderboard bool `json:"leaderboard" bson:"leaderboard"`
}

type UserPreferences struct {
	Profile ProfilePreferences `json:"profile" bson:"profile"`
	Theme   string             `json:"theme" bson:"theme"`
}

type User struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Username    string             `json:"username" bson:"username"`
	Email       string             `json:"email" bson:"email"`
	Phone       string             `json:"phone" bson:"phone"`
	Bio         string             `json:"bio" bson:"bio"`
	Password    string             `json:"-" bson:"password"`
	Preferences UserPreferences    `json:"preferences" bson:"preferences"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,min=10"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
}

type UserModel struct {
	collection *mongo.Collection
}

func NewUserModel(db *mongo.Database) *UserModel {
	return &UserModel{
		collection: db.Collection("users"),
	}
}

func (u *UserModel) CreateUser(req *CreateUserRequest) (*User, error) {
	// Check if user already exists by email
	existingUser := &User{}
	err := u.collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(existingUser)
	if err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Check if username already exists
	err = u.collection.FindOne(context.Background(), bson.M{"username": req.Username}).Decode(existingUser)
	if err == nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		Name:     req.Name,
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Preferences: UserPreferences{
			Profile: ProfilePreferences{
				Visibility:  true,
				Progress:    true,
				Leaderboard: true,
			},
			Theme: "light",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := u.collection.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (u *UserModel) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := u.collection.FindOne(context.Background(), bson.M{"email": email}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserModel) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := u.collection.FindOne(context.Background(), bson.M{"username": username}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserModel) GetUserByID(id primitive.ObjectID) (*User, error) {
	user := &User{}
	err := u.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserModel) ValidatePassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (u *UserModel) UpdateUser(id primitive.ObjectID, updates bson.M) error {
	updates["updatedAt"] = time.Now()
	_, err := u.collection.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$set": updates})
	return err
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

type DeleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

func (u *UserModel) ChangePassword(userID primitive.ObjectID, currentPassword, newPassword string) error {
	// Get user
	user, err := u.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Validate current password
	if !u.ValidatePassword(user, currentPassword) {
		return errors.New("current password is incorrect")
	}

	// Check if new password is same as current
	if currentPassword == newPassword {
		return errors.New("new password must be different from current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	return u.UpdateUser(userID, bson.M{"password": string(hashedPassword)})
}

func (u *UserModel) DeleteAccount(userID primitive.ObjectID, password string) error {
	// Get user
	user, err := u.GetUserByID(userID)
	if err != nil {
		return err
	}

	// Validate password
	if !u.ValidatePassword(user, password) {
		return errors.New("incorrect password")
	}

	// Delete user
	_, err = u.collection.DeleteOne(context.Background(), bson.M{"_id": userID})
	return err
}
