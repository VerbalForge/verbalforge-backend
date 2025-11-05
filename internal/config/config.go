package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	MongoURI           string
	JWTSecret          string
	FrontendURLs       []string
	UploadDir          string
	Port               string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

var AppConfig *Config

// Load initializes the application configuration from environment variables
func Load() *Config {
	// Load .env file - try multiple locations
	envPaths := []string{
		".env",       // Current directory
		"../.env",    // Parent directory (if running from tmp/)
		"../../.env", // Two levels up
	}

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

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

	// Load Mongo URI
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/verbalforge"
	}

	AppConfig = &Config{
		MongoURI:           mongoURI,
		JWTSecret:          getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		FrontendURLs:       frontendURLs,
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		Port:               getEnv("PORT", "8080"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/auth/google/callback"),
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
