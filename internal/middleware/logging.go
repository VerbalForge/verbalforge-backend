package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"verbalforge-backend/pkg/logger"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Build log message
		if raw != "" {
			path = path + "?" + raw
		}

		logger.Infof("[%s] %s %s - Status: %d - Latency: %v",
			c.Request.Method,
			path,
			c.ClientIP(),
			statusCode,
			latency,
		)
	}
}
