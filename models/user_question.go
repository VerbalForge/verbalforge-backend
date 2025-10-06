package models

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserQuestion represents a user's attempt on a question
type UserQuestion struct {
	ID            string     `bson:"_id,omitempty" json:"id,omitempty"`
	UserID        string     `bson:"user_id" json:"userId"`
	QuestionID    string     `bson:"question_id" json:"questionId"`
	Solved        bool       `bson:"solved" json:"solved"`
	Attempted     bool       `bson:"attempted" json:"attempted"`
	LastAttemptAt time.Time  `bson:"last_attempt_at" json:"lastAttemptAt"`
	LastSolvedAt  *time.Time `bson:"last_solved_at,omitempty" json:"lastSolvedAt,omitempty"`
	TimeTaken     int        `bson:"time_taken" json:"timeTaken"` // in seconds
}

// PassageProgress represents derived passage progress.
type PassageProgress struct {
	PassageID         string   `json:"passageId"`
	SolvedQuestionIDs []string `json:"solvedQuestionIds"`
	TotalQuestions    int      `json:"totalQuestions"`
	Solved            bool     `json:"solved"`
	Attempted         bool     `json:"attempted"`
}

type UserQuestionModel struct {
	collection *mongo.Collection
	db         *mongo.Database
}

func NewUserQuestionModel(db *mongo.Database) *UserQuestionModel {
	return &UserQuestionModel{
		collection: db.Collection("user_question"),
		db:         db,
	}
}

// UpsertUserQuestion creates or updates a user's question attempt
func (m *UserQuestionModel) UpsertUserQuestion(userID, questionID string, solved bool, timeTaken int) (*UserQuestion, error) {
	filter := bson.M{
		"user_id":     userID,
		"question_id": questionID,
	}

	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"solved":          solved,
			"attempted":       true,
			"last_attempt_at": now,
			"time_taken":      timeTaken,
		},
	}

	// Only update last_solved_at if solved is true
	if solved {
		update["$set"].(bson.M)["last_solved_at"] = now
	}

	opts := options.Update().SetUpsert(true)
	_, err := m.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return nil, err
	}

	// Retrieve and return the updated document
	var userQuestion UserQuestion
	err = m.collection.FindOne(context.Background(), filter).Decode(&userQuestion)
	if err != nil {
		return nil, err
	}

	return &userQuestion, nil
}

// GetUserQuestion retrieves a user's progress on a specific question
func (m *UserQuestionModel) GetUserQuestion(userID, questionID string) (*UserQuestion, error) {
	filter := bson.M{
		"user_id":     userID,
		"question_id": questionID,
	}

	var userQuestion UserQuestion
	err := m.collection.FindOne(context.Background(), filter).Decode(&userQuestion)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userQuestion, nil
}

// GetUserQuestionsByIDs retrieves user's progress for multiple questions
func (m *UserQuestionModel) GetUserQuestionsByIDs(userID string, questionIDs []string) (map[string]*UserQuestion, error) {
	filter := bson.M{
		"user_id":     userID,
		"question_id": bson.M{"$in": questionIDs},
	}

	cursor, err := m.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	result := make(map[string]*UserQuestion)
	for cursor.Next(context.Background()) {
		var userQuestion UserQuestion
		if err := cursor.Decode(&userQuestion); err != nil {
			continue
		}
		result[userQuestion.QuestionID] = &userQuestion
	}

	return result, nil
}

// GetPassageProgress derives passage progress by aggregating user's question attempts
func (m *UserQuestionModel) GetPassageProgress(userID, passageID string) (*PassageProgress, error) {
	// First get all question IDs for this passage
	var passage struct {
		QuestionIDs []string `bson:"question_ids"`
	}

	err := m.db.Collection("passages").FindOne(context.Background(), bson.M{"_id": passageID}).Decode(&passage)
	if err != nil {
		return nil, err
	}

	// Get user's attempts for these questions
	filter := bson.M{
		"user_id":     userID,
		"question_id": bson.M{"$in": passage.QuestionIDs},
	}

	cursor, err := m.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	solvedQuestionIDs := []string{}
	attempted := false

	for cursor.Next(context.Background()) {
		var userQuestion UserQuestion
		if err := cursor.Decode(&userQuestion); err != nil {
			continue
		}
		attempted = true
		if userQuestion.Solved {
			solvedQuestionIDs = append(solvedQuestionIDs, userQuestion.QuestionID)
		}
	}

	progress := &PassageProgress{
		PassageID:         passageID,
		SolvedQuestionIDs: solvedQuestionIDs,
		TotalQuestions:    len(passage.QuestionIDs),
		Solved:            len(solvedQuestionIDs) == len(passage.QuestionIDs) && len(passage.QuestionIDs) > 0,
		Attempted:         attempted,
	}

	return progress, nil
}

// GetBulkPassageProgress retrieves progress for multiple passages
func (m *UserQuestionModel) GetBulkPassageProgress(userID string, passageIDs []string) (map[string]*PassageProgress, error) {
	result := make(map[string]*PassageProgress)

	for _, passageID := range passageIDs {
		progress, err := m.GetPassageProgress(userID, passageID)
		if err != nil {
			continue
		}
		result[passageID] = progress
	}

	return result, nil
}
