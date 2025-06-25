// package cmd
package main

import (
	"log"
	"time"

	"github.com/absmach/magistrala/api" // Replace 'your-module-name' with the actual module name
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize user storage before starting the server
	if err := api.InitUserStorage(); err != nil {
		log.Fatalf("Failed to initialize user storage: %v", err)
	}

	router := gin.Default()

	// CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Allow your frontend origin
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Register the devices API routes
	api.SetupRoutes(router)

	// Register the WebSocket route if applicable
	api.SetupWebSocketRoute(router)

	log.Println("Server starting on port 9000...")
	router.Run(":9000") // Start the server on port 9000
}
