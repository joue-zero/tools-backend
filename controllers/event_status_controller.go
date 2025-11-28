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

	var eventStatusesList []models.EventStatus
	if err = cursor.All(context.TODO(), &eventStatusesList); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process attendees")
		return
	}

	// Map UserID -> EventStatus
	statusMap := make(map[primitive.ObjectID]models.EventStatus)
	for _, es := range eventStatusesList {
		statusMap[es.UserID] = es
	}

	// Collect all participant UserIDs (excluding organizer if desired, but usually organizer is also a participant)
	// The requirement implies "attendees", usually meaning those invited.
	// Let's include all participants with RoleAttendee.
	var attendeeIDs []primitive.ObjectID
	for _, p := range event.Participants {
		if p.Role == models.RoleAttendee {
			attendeeIDs = append(attendeeIDs, p.UserID)
		}
	}

	// Fetch User details
	userCollection := database.GetCollection("users")
	userCursor, err := userCollection.Find(context.TODO(), bson.M{"_id": bson.M{"$in": attendeeIDs}})
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch user details")
		return
	}
	defer userCursor.Close(context.TODO())

	var users []models.User
	if err = userCursor.All(context.TODO(), &users); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process user details")
		return
	}

	// Map UserID -> User
	userMap := make(map[primitive.ObjectID]models.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Build response list
	var attendeesDetails []models.EventAttendeeDetail
	counts := gin.H{
		"going":       0,
		"maybe":       0,
		"not_going":   0,
		"no_response": 0,
	}

	for _, uid := range attendeeIDs {
		user, userExists := userMap[uid]
		if !userExists {
			continue // Should not happen if data is consistent
		}

		detail := models.EventAttendeeDetail{
			UserID: uid,
			Name:   user.Name,
			Email:  user.Email,
			Status: models.StatusNoResponse,
		}

		if status, hasStatus := statusMap[uid]; hasStatus {
			detail.Status = status.Status
			detail.UpdatedAt = status.UpdatedAt

			// Update counts
			switch status.Status {
			case models.StatusGoing:
				counts["going"] = counts["going"].(int) + 1
			case models.StatusMaybe:
				counts["maybe"] = counts["maybe"].(int) + 1
			case models.StatusNotGoing:
				counts["not_going"] = counts["not_going"].(int) + 1
			}
		} else {
			counts["no_response"] = counts["no_response"].(int) + 1
		}

		attendeesDetails = append(attendeesDetails, detail)
	}

	counts["total"] = len(attendeesDetails)
	counts["attendees"] = attendeesDetails

	utils.SuccessResponse(c, 200, "Attendees retrieved successfully", counts)
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

	var eventStatusesList []models.EventStatus
	if err = cursor.All(context.TODO(), &eventStatusesList); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process attendees")
		return
	}

	// Map UserID -> EventStatus
	statusMap := make(map[primitive.ObjectID]models.EventStatus)
	var userIDs []primitive.ObjectID
	for _, es := range eventStatusesList {
		statusMap[es.UserID] = es
		userIDs = append(userIDs, es.UserID)
	}

	if len(userIDs) == 0 {
		utils.SuccessResponse(c, 200, "Attendees retrieved successfully", []models.EventAttendeeDetail{})
		return
	}

	// Fetch User details
	userCollection := database.GetCollection("users")
	userCursor, err := userCollection.Find(context.TODO(), bson.M{"_id": bson.M{"$in": userIDs}})
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch user details")
		return
	}
	defer userCursor.Close(context.TODO())

	var users []models.User
	if err = userCursor.All(context.TODO(), &users); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process user details")
		return
	}

	// Map UserID -> User
	userMap := make(map[primitive.ObjectID]models.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Build response list
	var attendeesDetails []models.EventAttendeeDetail
	for _, uid := range userIDs {
		user, userExists := userMap[uid]
		if !userExists {
			continue
		}

		es := statusMap[uid]
		attendeesDetails = append(attendeesDetails, models.EventAttendeeDetail{
			UserID:    uid,
			Name:      user.Name,
			Email:     user.Email,
			Status:    es.Status,
			UpdatedAt: es.UpdatedAt,
		})
	}

	utils.SuccessResponse(c, 200, "Attendees retrieved successfully", attendeesDetails)
}
