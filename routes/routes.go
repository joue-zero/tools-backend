package routes

import (
	"tools-backend/controllers"
	"tools-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all routes (similar to Laravel's routes/web.php)
func SetupRoutes() *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())

	// Initialize controllers
	authController := &controllers.AuthController{}

	// API version 1
	v1 := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/")
		{
			// Auth routes
			public.POST("/register", authController.Register)
			public.POST("/login", authController.Login)
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(middleware.Auth())
		{
			// Add your protected routes here
			// Example: user profile, protected resources, etc.
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	return router
}
