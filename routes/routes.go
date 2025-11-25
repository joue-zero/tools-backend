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
	eventController := &controllers.EventController{}
	rsvpController := &controllers.RSVPController{}
	searchController := &controllers.SearchController{}

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
			// Event Management routes
			protected.POST("/events", eventController.CreateEvent)
			protected.GET("/events/:id", eventController.GetEventByID)
			protected.GET("/events/organized", eventController.GetOrganizedEvents)
			protected.GET("/events/invited", eventController.GetInvitedEvents)
			protected.PUT("/events/:id", eventController.UpdateEvent)
			protected.DELETE("/events/:id", eventController.DeleteEvent)
			protected.POST("/events/:id/invite", eventController.InviteToEvent)

			// RSVP Management routes
			protected.POST("/events/:id/rsvp", rsvpController.CreateOrUpdateRSVP)
			protected.GET("/events/:id/rsvp/status", rsvpController.GetUserRSVPStatus)
			protected.GET("/events/:id/attendees", rsvpController.GetEventAttendees)
			protected.GET("/events/:id/attendees/status", rsvpController.GetAttendeesByStatus)

			// Search and Filtering routes
			protected.POST("/search", searchController.SearchEvents)
			protected.GET("/search/advanced", searchController.AdvancedSearch)
			protected.POST("/search/advanced", searchController.AdvancedSearch)
			protected.GET("/search/keyword", searchController.FilterEventsByKeyword)
			protected.GET("/search/date", searchController.FilterEventsByDate)
			protected.GET("/search/role", searchController.FilterEventsByRole)
			protected.GET("/all-events", searchController.GetAllUserEvents)
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
