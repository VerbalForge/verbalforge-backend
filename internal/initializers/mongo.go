package initializers

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"verbalforge-backend/internal/config"
	"verbalforge-backend/pkg/logger"
)

var DB *mongo.Database

// InitMongo initializes the MongoDB connection
func InitMongo() *mongo.Database {
	cfg := config.GetConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Fatal("Failed to ping MongoDB:", err)
	}

	DB = client.Database("verbalforge")
	logger.Info("Successfully connected to MongoDB")
	return DB
}

// GetDB returns the database instance
func GetDB() *mongo.Database {
	if DB == nil {
		return InitMongo()
	}
	return DB
}
