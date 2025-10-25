package middleware

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"verbalforge-backend/internal/config"
)

// CORSMiddleware configures CORS settings
func CORSMiddleware() gin.HandlerFunc {
	cfg := config.GetConfig()

	// Log the allowed origins for debugging
	log.Printf("[CORS] Allowed origins: %v", cfg.FrontendURLs)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = cfg.FrontendURLs
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	corsConfig.AllowCredentials = true

	return cors.New(corsConfig)
}

// LogCORSRequest logs incoming CORS-related requests for debugging
func LogCORSRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		method := c.Request.Method

		if origin != "" {
			log.Printf("[CORS] Request from origin: %s, method: %s, path: %s", origin, method, c.Request.URL.Path)
		}

		c.Next()
	}
}
