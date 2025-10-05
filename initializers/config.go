package initializers

import (
	"os"
)

type Config struct {
	MongoURI    string
	JWTSecret   string
	FrontendURL string
	UploadDir   string
	Port        string
}

var AppConfig *Config

// LoadConfig initializes the application configuration
func LoadConfig() {
	AppConfig = &Config{
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017/verbalforge"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		Port:        getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
