package main

import (
	"log"
	"tools-backend/config"
	"tools-backend/database"
	"tools-backend/routes"
)

func main() {
	// Load environment variables (similar to Laravel's .env)
	config.LoadEnv()

	// Connect to MongoDB (similar to Laravel's database connection)
	database.Connect()

	// Setup routes (similar to Laravel's routes/web.php)
	router := routes.SetupRoutes()

	// Start server (similar to Laravel's php artisan serve)
	port := config.GetEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
