package initializers

import (
	"verbalforge-backend/internal/config"
	"verbalforge-backend/pkg/logger"
)

// InitEnv initializes environment configuration
func InitEnv() {
	logger.Info("Loading environment configuration...")
	config.Load()
	logger.Infof("Configuration loaded. Port: %s", config.GetConfig().Port)
}
