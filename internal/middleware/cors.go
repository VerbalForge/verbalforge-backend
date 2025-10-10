package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/config"
)

// CORSMiddleware configures CORS settings
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{cfg.FrontendURL}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}

	return cors.New(corsConfig)
}
