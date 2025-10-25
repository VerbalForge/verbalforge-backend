package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"verbalforge-backend/internal/config"
	"verbalforge-backend/internal/controllers"
	"verbalforge-backend/internal/handlers"
	"verbalforge-backend/internal/middleware"
	"verbalforge-backend/internal/repository"
	"verbalforge-backend/internal/services"
)

// SetupRouter initializes and configures all routes
func SetupRouter(db *mongo.Database) *gin.Engine {
	cfg := config.GetConfig()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	questionRepo := repository.NewQuestionRepository(db)
	passageRepo := repository.NewPassageRepository(db)
	discussionRepo := repository.NewDiscussionRepository(db)
	userQuestionRepo := repository.NewUserQuestionRepository(db)
	userActivityRepo := repository.NewUserActivityRepository(db)
	userActivitySummaryRepo := repository.NewUserActivitySummaryRepository(db)
	wordRepo := repository.NewWordRepository(db)
	userWordRepo := repository.NewUserWordRepository(db)
	faqRepo := repository.NewFAQRepository(db)
	supportRepo := repository.NewSupportTicketRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	activitySummaryService := services.NewUserActivitySummaryService(userActivitySummaryRepo)
	activityService := services.NewUserActivityService(userActivityRepo, activitySummaryService, userRepo)
	userService := services.NewUserService(userRepo, userQuestionRepo, activityService)
	questionService := services.NewQuestionService(questionRepo, userQuestionRepo, passageRepo, userRepo, activityService)
	passageService := services.NewPassageService(passageRepo, questionRepo, userQuestionRepo, userRepo, activityService)
	discussionService := services.NewDiscussionService(discussionRepo, userRepo, questionRepo, passageRepo, activityService)
	fileService := services.NewFileService(cfg.UploadDir)
	wordService := services.NewWordService(wordRepo)
	userWordService := services.NewUserWordService(userWordRepo)
	faqService := services.NewFAQService(faqRepo)
	supportService := services.NewSupportTicketService(supportRepo)
	feedbackService := services.NewFeedbackService(feedbackRepo)
	cacheService := services.NewCacheService()
	practiceService := services.NewPracticeService(questionRepo, passageRepo, cacheService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService, activityService)
	questionHandler := handlers.NewQuestionHandler(questionService)
	passageHandler := handlers.NewPassageHandler(passageService)
	discussionHandler := handlers.NewDiscussionHandler(discussionService)
	fileHandler := handlers.NewFileHandler(fileService)
	faqHandler := handlers.NewFAQHandler(faqService)
	supportHandler := handlers.NewSupportTicketHandler(supportService)
	feedbackHandler := handlers.NewFeedbackHandler(feedbackService)
	practiceHandler := handlers.NewPracticeHandler(practiceService)

	// Initialize controllers
	wordController := controllers.NewWordController(wordService, userWordService)

	// Create router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LogCORSRequest())
	router.Use(middleware.LoggingMiddleware())

	// Inject database into context for compatibility (if needed)
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.GetCurrentUser)
		auth.POST("/refresh", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.RefreshToken)
		auth.POST("/change-password", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.ChangePassword)
		auth.DELETE("/delete-account", middleware.AuthMiddleware(cfg.JWTSecret), authHandler.DeleteAccount)
	}

	// User routes
	user := router.Group("/user")
	user.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		user.GET("/profile", userHandler.GetProfile)
		user.PUT("/profile", userHandler.UpdateProfile)
		user.GET("/preferences", userHandler.GetPreferences)
		user.PATCH("/preferences", userHandler.UpdatePreference)
		user.PATCH("/theme", userHandler.UpdateTheme)

		// User question progress routes
		user.POST("/questions/:question_id/attempt", questionHandler.SubmitQuestionAttempt)
		user.POST("/passages/:passage_id/attempt", passageHandler.SubmitPassageAttempt)
		user.GET("/questions/:question_id/progress", questionHandler.GetQuestionProgress)
		user.GET("/passages/:passage_id/progress", passageHandler.GetPassageProgress)
		user.POST("/questions/progress/bulk", questionHandler.GetBulkQuestionProgress)
		user.POST("/passages/progress/bulk", passageHandler.GetBulkPassageProgress)
	}

	// Profile routes (public access with privacy controls)
	profile := router.Group("/profile")
	profile.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		profile.GET("/:username", userHandler.GetUserProfile)
		profile.POST("/:username/view", userHandler.IncrementProfileView)
		profile.GET("/:username/stats", userHandler.GetUserStatsByUsername)
		profile.GET("/:username/activity", userHandler.GetUserRecentActivity)
		profile.GET("/:username/activity/summary", userHandler.GetActivityCalendar)
	}

	// Leaderboard routes
	leaderboard := router.Group("/leaderboard")
	leaderboard.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		leaderboard.GET("", userHandler.GetLeaderboard)
		leaderboard.POST("/ranks/update", userHandler.UpdateRanks)
	}

	// Practice routes (unified endpoint for questions + passages)
	practice := router.Group("/practice")
	practice.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		practice.GET("", practiceHandler.GetPracticeItems)
		practice.POST("/invalidate", practiceHandler.InvalidatePracticeCache)
	}

	// Question routes
	questions := router.Group("/questions")
	questions.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		questions.GET("/:id", questionHandler.GetQuestionByID)
	}

	// Passage routes
	passages := router.Group("/passages")
	passages.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		passages.GET("/:id", passageHandler.GetPassageByID)
	}

	// File upload routes
	router.POST("/upload", middleware.AuthMiddleware(cfg.JWTSecret), fileHandler.UploadFile)

	// Discussion routes
	discussions := router.Group("/discussions")
	discussions.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		discussions.GET("", discussionHandler.GetDiscussions)
		discussions.POST("", discussionHandler.CreateDiscussion)
		discussions.GET("/search", discussionHandler.SearchDiscussions)
		discussions.GET("/:id", discussionHandler.GetDiscussionByID)
		discussions.PUT("/:id", discussionHandler.UpdateDiscussion)
		discussions.DELETE("/:id", discussionHandler.DeleteDiscussion)
		discussions.POST("/:id/view", discussionHandler.IncrementDiscussionView)
		discussions.POST("/:id/like", discussionHandler.ToggleDiscussionLike)

		// Comment routes
		discussions.POST("/:id/comments", discussionHandler.AddComment)
		discussions.PUT("/:id/comments/:commentId", discussionHandler.UpdateComment)
		discussions.DELETE("/:id/comments/:commentId", discussionHandler.DeleteComment)
		discussions.POST("/:id/comments/:commentId/like", discussionHandler.ToggleCommentLike)
	}

	// User's discussions route
	router.GET("/profile/:username/discussions", middleware.AuthMiddleware(cfg.JWTSecret), discussionHandler.GetUserDiscussions)

	// Word routes (for Learn feature)
	words := router.Group("/words")
	words.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		words.GET("", wordController.GetWords)
		words.GET("/sources", wordController.GetSources)
		words.GET("/search", wordController.SearchWords)
		words.GET("/progress", wordController.GetUserProgress)
		words.POST("/progress/reset", wordController.ResetUserProgress)
		words.GET("/:id", wordController.GetWordByID)
		words.POST("/:id/status", wordController.UpdateWordStatus)
	}

	// Serve uploaded files statically
	router.Static("/uploads", cfg.UploadDir)

	// FAQ routes (Public POST, Admin GET)
	router.POST("/faq", faqHandler.CreateFAQ)

	// Support routes (Authenticated POST, Admin GET)
	support := router.Group("/support")
	{
		support.POST("", middleware.AuthMiddleware(cfg.JWTSecret), supportHandler.CreateSupportTicket)
		support.GET("/my-tickets", middleware.AuthMiddleware(cfg.JWTSecret), supportHandler.GetUserSupportTickets)
	}

	// Feedback routes (Public POST, Admin GET)
	router.POST("/feedback", feedbackHandler.CreateFeedback)

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	admin.Use(middleware.AdminMiddleware(userRepo))
	{
		// FAQ admin routes
		admin.GET("/faq", faqHandler.GetAllFAQs)
		admin.PATCH("/faq/:id/status", faqHandler.UpdateFAQStatus)
		admin.DELETE("/faq/:id", faqHandler.DeleteFAQ)

		// Support admin routes
		admin.GET("/support", supportHandler.GetAllSupportTickets)
		admin.PATCH("/support/:id", supportHandler.UpdateSupportTicket)
		admin.DELETE("/support/:id", supportHandler.DeleteSupportTicket)

		// Feedback admin routes
		admin.GET("/feedback", feedbackHandler.GetAllFeedback)
		admin.PATCH("/feedback/:id/status", feedbackHandler.UpdateFeedbackStatus)
		admin.DELETE("/feedback/:id", feedbackHandler.DeleteFeedback)
	}

	return router
}
