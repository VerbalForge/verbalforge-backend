package initializers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"verbalforge-backend/controllers"
	"verbalforge-backend/middleware"
)

// SetupRouter initializes the Gin router with all routes
func SetupRouter() *gin.Engine {
	// Initialize controllers with dependencies
	controllers.InitializeAuthController(DB, AppConfig.JWTSecret)
	controllers.InitializeUserController(DB)
	controllers.InitializeFileController(AppConfig.UploadDir)
	controllers.InitializeQuestionController(DB)
	controllers.InitializePassageController(DB)

	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{AppConfig.FrontendURL}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.GET("/me", middleware.AuthRequired(AppConfig.JWTSecret), controllers.GetCurrentUser)
		auth.POST("/change-password", middleware.AuthRequired(AppConfig.JWTSecret), controllers.ChangePassword)
		auth.DELETE("/delete-account", middleware.AuthRequired(AppConfig.JWTSecret), controllers.DeleteAccount)
	}

	// User routes
	user := router.Group("/user")
	user.Use(middleware.AuthRequired(AppConfig.JWTSecret))
	{
		user.GET("/profile", controllers.GetProfile)
		user.PUT("/profile", controllers.UpdateProfile)
		user.GET("/stats", controllers.GetUserStats)
		user.GET("/preferences", controllers.GetPreferences)
		user.PATCH("/preferences", controllers.UpdatePreference)
		user.PATCH("/theme", controllers.UpdateTheme)
	}

	// Question routes
	questions := router.Group("/questions")
	questions.Use(middleware.AuthRequired(AppConfig.JWTSecret))
	{
		questions.GET("", controllers.GetQuestions)
		questions.GET("/:id", controllers.GetQuestionByID)
	}

	// Passage routes
	passages := router.Group("/passages")
	passages.Use(middleware.AuthRequired(AppConfig.JWTSecret))
	{
		passages.GET("", controllers.GetPassages)
		passages.GET("/:id", controllers.GetPassageByID)
	}

	// File upload routes
	router.POST("/upload", middleware.AuthRequired(AppConfig.JWTSecret), controllers.UploadFile)

	// Serve uploaded files statically
	router.Static("/uploads", AppConfig.UploadDir)

	return router
}
