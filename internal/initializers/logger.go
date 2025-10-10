package initializers

import "verbalforge-backend/pkg/logger"

// InitLogger initializes the application logger
func InitLogger() {
	logger.Init()
	logger.Info("Logger initialized successfully")
}
