package main

import (
"context"
"encoding/json"
"fmt"
"log"
"os"

"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/mongo"
"go.mongodb.org/mongo-driver/mongo/options"
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
	ID          string          `json:"id" bson:"_id,omitempty"`
	Username    string          `json:"username" bson:"username"`
	Email       string          `json:"email" bson:"email"`
	Preferences UserPreferences `json:"preferences" bson:"preferences"`
}

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection := client.Database("verbalforge").Collection("users")

	// Find a sample user
	var user User
	err = collection.FindOne(context.Background(), bson.M{}).Decode(&user)
	if err != nil {
		log.Fatal("No users found:", err)
	}

	// Pretty print the user
	jsonData, _ := json.MarshalIndent(user, "", "  ")
	fmt.Println("Sample user preferences:")
	fmt.Println(string(jsonData))

	// Count users without preferences
	count, err := collection.CountDocuments(context.Background(), bson.M{
		"$or": []bson.M{
			{"preferences": bson.M{"$exists": false}},
			{"preferences.profile": bson.M{"$exists": false}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nUsers without proper preferences: %d\n", count)
}
