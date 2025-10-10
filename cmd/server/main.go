package main

import (
	"verbalforge-backend/internal/config"
	"verbalforge-backend/internal/initializers"
	"verbalforge-backend/internal/routes"
	"verbalforge-backend/pkg/logger"
)

func main() {
	// Initialize logger
	initializers.InitLogger()
	logger.Info("Starting VerbalForge Backend Server...")

	// Load environment configuration
	initializers.InitEnv()

	// Initialize MongoDB
	db := initializers.InitMongo()

	// Setup router with all dependencies
	router := routes.SetupRouter(db)

	// Get configuration
	cfg := config.GetConfig()

	// Start server
	logger.Infof("Server running on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		logger.Fatal("Failed to start server:", err)
	}
}
