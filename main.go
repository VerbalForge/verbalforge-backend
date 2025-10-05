package main

import (
	"log"

	"verbalforge-backend/initializers"
)

func init() {
	initializers.LoadConfig()
	initializers.ConnectDatabase()
}

func main() {
	router := initializers.SetupRouter()

	log.Printf("Server running on :%s", initializers.AppConfig.Port)

	if err := router.Run(":" + initializers.AppConfig.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
