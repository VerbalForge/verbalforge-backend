package config

import (
	"os"
	"strings"
)

// Config holds all application configuration
type Config struct {
	MongoURI     string
	JWTSecret    string
	FrontendURLs []string
	UploadDir    string
	Port         string
}

var AppConfig *Config

// Load initializes the application configuration from environment variables
func Load() *Config {
	// Parse comma-separated FRONTEND_URLS
	var frontendURLs []string
	if urlsEnv := os.Getenv("FRONTEND_URLS"); urlsEnv != "" {
		for _, url := range strings.Split(urlsEnv, ",") {
			trimmed := strings.TrimSpace(url)
			if trimmed != "" {
				frontendURLs = append(frontendURLs, trimmed)
			}
		}
	} else {
		// Default to localhost
		frontendURLs = []string{"http://localhost:3000"}
	}


	// Load primary Mongo URI (no legacy fallbacks)
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/verbalforge"
	}

	AppConfig = &Config{
		MongoURI:     mongoURI,
		JWTSecret:    getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		FrontendURLs: frontendURLs,
		UploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
		Port:         getEnv("PORT", "8080"),
	}
	return AppConfig
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	if AppConfig == nil {
		return Load()
	}
	return AppConfig
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
