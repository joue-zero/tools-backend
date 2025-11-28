package controllers

import (
	"context"
	"time"
	"tools-backend/database"
	"tools-backend/models"
	"tools-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventController struct{}

// CreateEvent creates a new event (creator becomes organizer)
func (ec *EventController) CreateEvent(c *gin.Context) {
	var req models.CreateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, errors)
		return
	}

	// Validate date is not in the past
	eventDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Compare with today (start of day)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if eventDate.Before(today) {
		utils.ErrorResponse(c, 400, "Event date cannot be in the past")
		return
	}

	// Get user ID from context (set by auth middleware)
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

	// Convert user ID string to ObjectID
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	// Create event with creator as organizer
	event := models.Event{
		Title:       req.Title,
		Description: req.Description,
		Date:        req.Date,
		Time:        req.Time,
		Location:    req.Location,
		Participants: []models.EventParticipant{
			{
				UserID: userObjectID,
				Role:   models.RoleOrganizer,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert event
	collection := database.GetCollection("events")
	result, err := collection.InsertOne(context.TODO(), event)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to create event")
		return
	}

	event.ID = result.InsertedID.(primitive.ObjectID)
	utils.SuccessResponse(c, 201, "Event created successfully", event.ToResponse())
}

// GetOrganizedEvents returns all events organized by the user
func (ec *EventController) GetOrganizedEvents(c *gin.Context) {
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

	// Find events where user is organizer
	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
				"role":    models.RoleOrganizer,
			},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.Event
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	eventResponses := make([]models.EventResponse, 0, len(events))
	for _, event := range events {
		eventResponses = append(eventResponses, event.ToResponse())
	}

	utils.SuccessResponse(c, 200, "Organized events retrieved successfully", eventResponses)
}

// GetInvitedEvents returns all events the user is invited to
func (ec *EventController) GetInvitedEvents(c *gin.Context) {
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

	// Find events where user is attendee
	filter := bson.M{
		"participants": bson.M{
			"$elemMatch": bson.M{
				"user_id": userObjectID,
				"role":    models.RoleAttendee,
			},
		},
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch events")
		return
	}
	defer cursor.Close(context.TODO())

	var events []models.Event
	if err = cursor.All(context.TODO(), &events); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process events")
		return
	}

	// Fetch statuses for these events for the current user
	eventIDs := make([]primitive.ObjectID, len(events))
	for i, e := range events {
		eventIDs[i] = e.ID
	}

	statusCollection := database.GetCollection("event_statuses")
	statusFilter := bson.M{
		"event_id": bson.M{"$in": eventIDs},
		"user_id":  userObjectID,
	}

	statusCursor, err := statusCollection.Find(context.TODO(), statusFilter)
	if err != nil {
		// Log error but continue, statuses will just be empty
	}

	statusMap := make(map[primitive.ObjectID]models.EventStatusValue)
	if err == nil {
		var statuses []models.EventStatus
		if err = statusCursor.All(context.TODO(), &statuses); err == nil {
			for _, s := range statuses {
				statusMap[s.EventID] = s.Status
			}
		}
		statusCursor.Close(context.TODO())
	}

	eventResponses := make([]models.EventResponse, 0, len(events))
	for _, event := range events {
		resp := event.ToResponse()
		if status, exists := statusMap[event.ID]; exists {
			resp.MyStatus = status
		} else {
			resp.MyStatus = models.StatusNoResponse
		}
		eventResponses = append(eventResponses, resp)
	}

	utils.SuccessResponse(c, 200, "Invited events retrieved successfully", eventResponses)
}

// GetEventByID returns a specific event by ID
func (ec *EventController) GetEventByID(c *gin.Context) {
	eventID := c.Param("id")

	eventObjectID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid event ID")
		return
	}

	collection := database.GetCollection("events")
	var event models.Event

	err = collection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, 404, "Event not found")
		} else {
			utils.ErrorResponse(c, 500, "Failed to fetch event")
		}
		return
	}

	utils.SuccessResponse(c, 200, "Event retrieved successfully", event.ToResponse())
}

// InviteToEvent invites users to an event (only organizer can invite)
func (ec *EventController) InviteToEvent(c *gin.Context) {
	eventID := c.Param("id")
	var req models.InviteToEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, errors)
		return
	}

	eventObjectID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid event ID")
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

	collection := database.GetCollection("events")
	var event models.Event

	// Find event and check if user is organizer
	err = collection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, 404, "Event not found")
		} else {
			utils.ErrorResponse(c, 500, "Failed to fetch event")
		}
		return
	}

	// Check if user is organizer
	isOrganizer := false
	for _, p := range event.Participants {
		if p.UserID == userObjectID && p.Role == models.RoleOrganizer {
			isOrganizer = true
			break
		}
	}

	if !isOrganizer {
		utils.ErrorResponse(c, 403, "Only event organizers can invite users")
		return
	}

	// Add new participants as attendees
	newParticipants := []models.EventParticipant{}
	for _, inviteUserID := range req.UserIDs {
		// Check if user is already invited
		alreadyInvited := false
		for _, p := range event.Participants {
			if p.UserID == inviteUserID {
				alreadyInvited = true
				break
			}
		}

		if !alreadyInvited {
			newParticipants = append(newParticipants, models.EventParticipant{
				UserID: inviteUserID,
				Role:   models.RoleAttendee,
			})
		}
	}

	if len(newParticipants) == 0 {
		utils.ErrorResponse(c, 400, "All users are already invited to this event")
		return
	}

	// Update event with new participants
	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": eventObjectID},
		bson.M{"$push": bson.M{"participants": bson.M{"$each": newParticipants}}, "$set": bson.M{"updated_at": time.Now()}},
	)

	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to invite users")
		return
	}

	utils.SuccessResponse(c, 200, "Users invited successfully", gin.H{
		"invited_count": len(newParticipants),
	})
}

// UpdateEvent updates event details (only organizer can update)
func (ec *EventController) UpdateEvent(c *gin.Context) {
	eventID := c.Param("id")
	var req models.UpdateEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Validate request
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, errors)
		return
	}

	eventObjectID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid event ID")
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

	collection := database.GetCollection("events")
	var event models.Event

	// Find event and check if user is organizer
	err = collection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, 404, "Event not found")
		} else {
			utils.ErrorResponse(c, 500, "Failed to fetch event")
		}
		return
	}

	// Check if user is organizer
	isOrganizer := false
	for _, p := range event.Participants {
		if p.UserID == userObjectID && p.Role == models.RoleOrganizer {
			isOrganizer = true
			break
		}
	}

	if !isOrganizer {
		utils.ErrorResponse(c, 403, "Only event organizers can update events")
		return
	}

	// Build update document with only provided fields
	updateDoc := bson.M{}
	if req.Title != "" {
		updateDoc["title"] = req.Title
	}
	if req.Description != "" {
		updateDoc["description"] = req.Description
	}
	if req.Date != "" {
		// Validate date is not in the past
		eventDate, err := time.Parse("2006-01-02", req.Date)
		if err != nil {
			utils.ErrorResponse(c, 400, "Invalid date format. Use YYYY-MM-DD")
			return
		}

		// Compare with today (start of day)
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		if eventDate.Before(today) {
			utils.ErrorResponse(c, 400, "Event date cannot be in the past")
			return
		}
		updateDoc["date"] = req.Date
	}
	if req.Time != "" {
		updateDoc["time"] = req.Time
	}
	if req.Location != "" {
		updateDoc["location"] = req.Location
	}
	updateDoc["updated_at"] = time.Now()

	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"_id": eventObjectID},
		bson.M{"$set": updateDoc},
	)

	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to update event")
		return
	}

	utils.SuccessResponse(c, 200, "Event updated successfully", nil)
}

// DeleteEvent deletes an event (only organizer can delete)
func (ec *EventController) DeleteEvent(c *gin.Context) {
	eventID := c.Param("id")

	eventObjectID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid event ID")
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

	collection := database.GetCollection("events")
	var event models.Event

	// Find event and check if user is organizer
	err = collection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, 404, "Event not found")
		} else {
			utils.ErrorResponse(c, 500, "Failed to fetch event")
		}
		return
	}

	// Check if user is organizer
	isOrganizer := false
	for _, p := range event.Participants {
		if p.UserID == userObjectID && p.Role == models.RoleOrganizer {
			isOrganizer = true
			break
		}
	}

	if !isOrganizer {
		utils.ErrorResponse(c, 403, "Only event organizers can delete events")
		return
	}

	// Delete event
	result, err := collection.DeleteOne(context.TODO(), bson.M{"_id": eventObjectID})
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to delete event")
		return
	}

	if result.DeletedCount == 0 {
		utils.ErrorResponse(c, 404, "Event not found")
		return
	}

	// Also delete related event statuses
	eventStatusCollection := database.GetCollection("event_statuses")
	eventStatusCollection.DeleteMany(context.TODO(), bson.M{"event_id": eventObjectID})

	utils.SuccessResponse(c, 200, "Event deleted successfully", nil)
}
