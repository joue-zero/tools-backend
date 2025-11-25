package controllers

import (
	"context"
	"regexp"
	"tools-backend/database"
	"tools-backend/models"
	"tools-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SearchController struct{}

// SearchRequest represents a request for searching and filtering events
type SearchRequest struct {
	Keyword string `json:"keyword"`          // Search in title and description
	StartDate string `json:"start_date"`     // ISO 8601 format: YYYY-MM-DD
	EndDate   string `json:"end_date"`       // ISO 8601 format: YYYY-MM-DD
	UserRole  string `json:"user_role"`      // organizer or attendee
	Location  string `json:"location"`       // Search by location
}

// SearchEvents performs advanced search and filtering on events
func (sc *SearchController) SearchEvents(c *gin.Context) {
	var req SearchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "Invalid request data")
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		utils.ErrorResponse(c, 401, "Invalid user ID format")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	// Build filter
	filter := buildEventSearchFilter(userObjectID, req)

	collection := database.GetCollection("events")
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to search events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.EventResponse
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	if len(events) == 0 {
		events = []models.EventResponse{}
	}

	utils.SuccessResponse(c, 200, "Search completed successfully", gin.H{
		"total_results": len(events),
		"events":        events,
	})
}

// GetAllUserEvents returns all events for the user (organized + invited)
func (sc *SearchController) GetAllUserEvents(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		utils.ErrorResponse(c, 401, "Invalid user ID format")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	collection := database.GetCollection("events")
	
	// Find all events where user is a participant
	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
			},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.EventResponse
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	if len(events) == 0 {
		events = []models.EventResponse{}
	}

	utils.SuccessResponse(c, 200, "All user events retrieved successfully", events)
}

// FilterEventsByDate returns events within a date range
func (sc *SearchController) FilterEventsByDate(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		utils.ErrorResponse(c, 401, "Invalid user ID format")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	// Build filter
	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
			},
		},
	}

	// Add date range filter if provided
	if startDate != "" || endDate != "" {
		dateFilter := bson.M{}
		if startDate != "" {
			dateFilter["$gte"] = startDate
		}
		if endDate != "" {
			dateFilter["$lte"] = endDate
		}
		if len(dateFilter) > 0 {
			filter["date"] = dateFilter
		}
	}

	collection := database.GetCollection("events")
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to filter events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.EventResponse
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	if len(events) == 0 {
		events = []models.EventResponse{}
	}

	utils.SuccessResponse(c, 200, "Events filtered by date successfully", events)
}

// FilterEventsByKeyword returns events matching a keyword in title or description
func (sc *SearchController) FilterEventsByKeyword(c *gin.Context) {
	keyword := c.Query("q")

	if keyword == "" {
		utils.ErrorResponse(c, 400, "Search keyword 'q' is required")
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		utils.ErrorResponse(c, 401, "Invalid user ID format")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	// Create case-insensitive regex pattern
	regexPattern := bson.M{"$regex": keyword, "$options": "i"}

	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
			},
		},
		"$or": bson.A{
			bson.M{"title": regexPattern},
			bson.M{"description": regexPattern},
		},
	}

	collection := database.GetCollection("events")
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to search events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.EventResponse
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	if len(events) == 0 {
		events = []models.EventResponse{}
	}

	utils.SuccessResponse(c, 200, "Events filtered by keyword successfully", gin.H{
		"keyword": keyword,
		"results": events,
	})
}

// FilterEventsByRole returns events where user has a specific role
func (sc *SearchController) FilterEventsByRole(c *gin.Context) {
	role := c.Query("role")

	// Validate role
	validRoles := map[string]bool{
		"organizer": true,
		"attendee":  true,
	}

	if role == "" {
		utils.ErrorResponse(c, 400, "Role parameter is required: organizer or attendee")
		return
	}

	if !validRoles[role] {
		utils.ErrorResponse(c, 400, "Invalid role. Must be: organizer or attendee")
		return
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		utils.ErrorResponse(c, 401, "Invalid user ID format")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
				"role":    role,
			},
		},
	}

	collection := database.GetCollection("events")
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to filter events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.EventResponse
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	if len(events) == 0 {
		events = []models.EventResponse{}
	}

	utils.SuccessResponse(c, 200, "Events filtered by role successfully", events)
}

// Helper function to build search filter
func buildEventSearchFilter(userObjectID primitive.ObjectID, req SearchRequest) bson.M {
	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
			},
		},
	}

	// Apply keyword filter
	if req.Keyword != "" {
		regexPattern := bson.M{"$regex": req.Keyword, "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"title": regexPattern},
			bson.M{"description": regexPattern},
		}
	}

	// Apply date range filter
	if req.StartDate != "" || req.EndDate != "" {
		dateFilter := bson.M{}
		if req.StartDate != "" {
			dateFilter["$gte"] = req.StartDate
		}
		if req.EndDate != "" {
			dateFilter["$lte"] = req.EndDate
		}
		if len(dateFilter) > 0 {
			filter["date"] = dateFilter
		}
	}

	// Apply location filter
	if req.Location != "" {
		regexPattern := bson.M{"$regex": req.Location, "$options": "i"}
		filter["location"] = regexPattern
	}

	// Apply user role filter
	if req.UserRole != "" {
		// Update the $elemMatch to include role filter if specified
		if elemMatch, exists := filter["participants"].(bson.M)["$elemMatch"]; exists {
			if elemMatchObj, ok := elemMatch.(bson.M); ok {
				elemMatchObj["role"] = req.UserRole
			}
		}
	}

	return filter
}

// AdvancedSearch performs complex search with multiple filters
func (sc *SearchController) AdvancedSearch(c *gin.Context) {
	var req SearchRequest

	// Try to bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body, try to get from query parameters
		req.Keyword = c.Query("keyword")
		req.StartDate = c.Query("start_date")
		req.EndDate = c.Query("end_date")
		req.UserRole = c.Query("user_role")
		req.Location = c.Query("location")
	}

	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, 401, "User ID not found in token")
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		utils.ErrorResponse(c, 401, "Invalid user ID format")
		return
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	// Validate user role if provided
	if req.UserRole != "" {
		validRoles := map[string]bool{
			"organizer": true,
			"attendee":  true,
		}
		if !validRoles[req.UserRole] {
			utils.ErrorResponse(c, 400, "Invalid user_role. Must be: organizer or attendee")
			return
		}
	}

	// Build filter using helper function
	filter := buildComplexSearchFilter(userObjectID, req)

	collection := database.GetCollection("events")
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to search events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.EventResponse
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	if len(events) == 0 {
		events = []models.EventResponse{}
	}

	utils.SuccessResponse(c, 200, "Advanced search completed successfully", gin.H{
		"filters":       req,
		"total_results": len(events),
		"events":        events,
	})
}

// Helper function for complex search filter
func buildComplexSearchFilter(userObjectID primitive.ObjectID, req SearchRequest) bson.M {
	// Base filter - user must be participant
	elemMatch := bson.M{
		"user_id": userObjectID,
	}

	// Add role to elemMatch if specified
	if req.UserRole != "" {
		elemMatch["role"] = req.UserRole
	}

	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": elemMatch,
		},
	}

	// Apply keyword filter on title and description
	if req.Keyword != "" {
		regexPattern := bson.M{"$regex": regexp.QuoteMeta(req.Keyword), "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"title": regexPattern},
			bson.M{"description": regexPattern},
		}
	}

	// Apply date range filter
	if req.StartDate != "" || req.EndDate != "" {
		dateFilter := bson.M{}
		if req.StartDate != "" {
			dateFilter["$gte"] = req.StartDate
		}
		if req.EndDate != "" {
			dateFilter["$lte"] = req.EndDate
		}
		if len(dateFilter) > 0 {
			filter["date"] = dateFilter
		}
	}

	// Apply location filter
	if req.Location != "" {
		regexPattern := bson.M{"$regex": regexp.QuoteMeta(req.Location), "$options": "i"}
		filter["location"] = regexPattern
	}

	return filter
}
