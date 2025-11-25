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

type EventStatusController struct{}

// CreateOrUpdateEventStatus creates or updates an event status response for an event
func (esc *EventStatusController) CreateOrUpdateEventStatus(c *gin.Context) {
	eventID := c.Param("id")
	var req models.EventStatusRequest

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

	// Verify user is invited to the event
	eventCollection := database.GetCollection("events")
	var event models.Event

	err = eventCollection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, 404, "Event not found")
		} else {
			utils.ErrorResponse(c, 500, "Failed to fetch event")
		}
		return
	}

	// Check if user is a participant (organizer or attendee)
	isParticipant := false
	for _, p := range event.Participants {
		if p.UserID == userObjectID {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		utils.ErrorResponse(c, 403, "User is not invited to this event")
		return
	}

	// Check if EventStatus already exists
	eventStatusCollection := database.GetCollection("event_statuses")
	var existingEventStatus models.EventStatus

	err = eventStatusCollection.FindOne(context.TODO(), bson.M{
		"event_id": eventObjectID,
		"user_id":  userObjectID,
	}).Decode(&existingEventStatus)

	if err == nil {
		// EventStatus exists, update it
		_, err = eventStatusCollection.UpdateOne(
			context.TODO(),
			bson.M{"_id": existingEventStatus.ID},
			bson.M{"$set": bson.M{
				"status":     req.Status,
				"updated_at": time.Now(),
			}},
		)

		if err != nil {
			utils.ErrorResponse(c, 500, "Failed to update event status")
			return
		}

		existingEventStatus.Status = req.Status
		existingEventStatus.UpdatedAt = time.Now()
		utils.SuccessResponse(c, 200, "Event status updated successfully", existingEventStatus.ToResponse())
	} else if err == mongo.ErrNoDocuments {
		// Create new EventStatus
		newEventStatus := models.EventStatus{
			EventID:   eventObjectID,
			UserID:    userObjectID,
			Status:    req.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result, err := eventStatusCollection.InsertOne(context.TODO(), newEventStatus)
		if err != nil {
			utils.ErrorResponse(c, 500, "Failed to create event status")
			return
		}

		newEventStatus.ID = result.InsertedID.(primitive.ObjectID)
		utils.SuccessResponse(c, 201, "Event status created successfully", newEventStatus.ToResponse())
	} else {
		utils.ErrorResponse(c, 500, "Failed to check event status")
	}
}

// GetEventAttendees returns all attendees and their event statuses for an event (organizer only)
func (esc *EventStatusController) GetEventAttendees(c *gin.Context) {
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

	// Verify user is organizer of the event
	eventCollection := database.GetCollection("events")
	var event models.Event

	err = eventCollection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
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
		utils.ErrorResponse(c, 403, "Only event organizers can view attendees")
		return
	}

	// Get all event statuses for this event
	eventStatusCollection := database.GetCollection("event_statuses")
	cursor, err := eventStatusCollection.Find(context.TODO(), bson.M{"event_id": eventObjectID})
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch attendees")
		return
	}
	defer cursor.Close(context.TODO())

	var eventStatuses []models.EventStatusResponse
	if err = cursor.All(context.TODO(), &eventStatuses); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process attendees")
		return
	}

	if len(eventStatuses) == 0 {
		eventStatuses = []models.EventStatusResponse{}
	}

	// Group attendees by status
	attendeesSummary := gin.H{
		"total":         len(event.Participants) - 1, // Total minus organizer
		"going":         0,
		"maybe":         0,
		"not_going":     0,
		"no_response":   0,
		"attendees":     eventStatuses,
	}

	eventStatusMap := make(map[primitive.ObjectID]models.EventStatusValue)
	for _, es := range eventStatuses {
		eventStatusMap[es.UserID] = es.Status
		switch es.Status {
		case models.StatusGoing:
			attendeesSummary["going"] = attendeesSummary["going"].(int) + 1
		case models.StatusMaybe:
			attendeesSummary["maybe"] = attendeesSummary["maybe"].(int) + 1
		case models.StatusNotGoing:
			attendeesSummary["not_going"] = attendeesSummary["not_going"].(int) + 1
		}
	}

	// Count those without response
	noResponseCount := 0
	for _, participant := range event.Participants {
		if participant.Role == models.RoleAttendee {
			if _, exists := eventStatusMap[participant.UserID]; !exists {
				noResponseCount++
			}
		}
	}
	attendeesSummary["no_response"] = noResponseCount

	utils.SuccessResponse(c, 200, "Attendees retrieved successfully", attendeesSummary)
}

// GetUserEventStatus returns the event status of the current user for an event
func (esc *EventStatusController) GetUserEventStatus(c *gin.Context) {
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

	eventStatusCollection := database.GetCollection("event_statuses")
	var eventStatus models.EventStatus

	err = eventStatusCollection.FindOne(context.TODO(), bson.M{
		"event_id": eventObjectID,
		"user_id":  userObjectID,
	}).Decode(&eventStatus)

	if err == mongo.ErrNoDocuments {
		utils.SuccessResponse(c, 200, "No event status found", gin.H{
			"status": "no_response",
		})
		return
	}

	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch event status")
		return
	}

	utils.SuccessResponse(c, 200, "Event status retrieved successfully", eventStatus.ToResponse())
}

// GetAttendeesByStatus returns attendees filtered by status for an event (organizer only)
func (esc *EventStatusController) GetAttendeesByStatus(c *gin.Context) {
	eventID := c.Param("id")
	status := c.Query("status") // Query parameter: going, maybe, not_going

	// Validate status
	validStatuses := map[string]bool{
		"going":     true,
		"maybe":     true,
		"not_going": true,
	}

	if status != "" && !validStatuses[status] {
		utils.ErrorResponse(c, 400, "Invalid status. Must be: going, maybe, or not_going")
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

	// Verify user is organizer of the event
	eventCollection := database.GetCollection("events")
	var event models.Event

	err = eventCollection.FindOne(context.TODO(), bson.M{"_id": eventObjectID}).Decode(&event)
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
		utils.ErrorResponse(c, 403, "Only event organizers can view attendees")
		return
	}

	// Build filter
	filter := bson.M{"event_id": eventObjectID}
	if status != "" {
		filter["status"] = models.EventStatusValue(status)
	}

	eventStatusCollection := database.GetCollection("event_statuses")
	cursor, err := eventStatusCollection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch attendees")
		return
	}
	defer cursor.Close(context.TODO())

	var eventStatuses []models.EventStatusResponse
	if err = cursor.All(context.TODO(), &eventStatuses); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process attendees")
		return
	}

	if len(eventStatuses) == 0 {
		eventStatuses = []models.EventStatusResponse{}
	}

	utils.SuccessResponse(c, 200, "Attendees retrieved successfully", eventStatuses)
}
