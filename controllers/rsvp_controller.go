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

type RSVPController struct{}

// CreateOrUpdateRSVP creates or updates an RSVP response for an event
func (rc *RSVPController) CreateOrUpdateRSVP(c *gin.Context) {
	eventID := c.Param("id")
	var req models.RSVPRequest

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

	// Check if RSVP already exists
	rsvpCollection := database.GetCollection("rsvps")
	var existingRSVP models.RSVP

	err = rsvpCollection.FindOne(context.TODO(), bson.M{
		"event_id": eventObjectID,
		"user_id":  userObjectID,
	}).Decode(&existingRSVP)

	if err == nil {
		// RSVP exists, update it
		_, err = rsvpCollection.UpdateOne(
			context.TODO(),
			bson.M{"_id": existingRSVP.ID},
			bson.M{"$set": bson.M{
				"status":     req.Status,
				"updated_at": time.Now(),
			}},
		)

		if err != nil {
			utils.ErrorResponse(c, 500, "Failed to update RSVP")
			return
		}

		existingRSVP.Status = req.Status
		existingRSVP.UpdatedAt = time.Now()
		utils.SuccessResponse(c, 200, "RSVP updated successfully", existingRSVP.ToResponse())
	} else if err == mongo.ErrNoDocuments {
		// Create new RSVP
		newRSVP := models.RSVP{
			EventID:   eventObjectID,
			UserID:    userObjectID,
			Status:    req.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result, err := rsvpCollection.InsertOne(context.TODO(), newRSVP)
		if err != nil {
			utils.ErrorResponse(c, 500, "Failed to create RSVP")
			return
		}

		newRSVP.ID = result.InsertedID.(primitive.ObjectID)
		utils.SuccessResponse(c, 201, "RSVP created successfully", newRSVP.ToResponse())
	} else {
		utils.ErrorResponse(c, 500, "Failed to check RSVP status")
	}
}

// GetEventAttendees returns all attendees and their RSVP statuses for an event (organizer only)
func (rc *RSVPController) GetEventAttendees(c *gin.Context) {
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

	// Get all RSVPs for this event
	rsvpCollection := database.GetCollection("rsvps")
	cursor, err := rsvpCollection.Find(context.TODO(), bson.M{"event_id": eventObjectID})
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch attendees")
		return
	}
	defer cursor.Close(context.TODO())

	var rsvps []models.RSVPResponse
	if err = cursor.All(context.TODO(), &rsvps); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process attendees")
		return
	}

	if len(rsvps) == 0 {
		rsvps = []models.RSVPResponse{}
	}

	// Group attendees by status
	attendeesSummary := gin.H{
		"total":         len(event.Participants) - 1, // Total minus organizer
		"going":         0,
		"maybe":         0,
		"not_going":     0,
		"no_response":   0,
		"attendees":     rsvps,
	}

	rsvpStatusMap := make(map[primitive.ObjectID]models.RSVPStatus)
	for _, rsvp := range rsvps {
		rsvpStatusMap[rsvp.UserID] = rsvp.Status
		switch rsvp.Status {
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
			if _, exists := rsvpStatusMap[participant.UserID]; !exists {
				noResponseCount++
			}
		}
	}
	attendeesSummary["no_response"] = noResponseCount

	utils.SuccessResponse(c, 200, "Attendees retrieved successfully", attendeesSummary)
}

// GetUserRSVPStatus returns the RSVP status of the current user for an event
func (rc *RSVPController) GetUserRSVPStatus(c *gin.Context) {
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

	rsvpCollection := database.GetCollection("rsvps")
	var rsvp models.RSVP

	err = rsvpCollection.FindOne(context.TODO(), bson.M{
		"event_id": eventObjectID,
		"user_id":  userObjectID,
	}).Decode(&rsvp)

	if err == mongo.ErrNoDocuments {
		utils.SuccessResponse(c, 200, "No RSVP status found", gin.H{
			"status": "no_response",
		})
		return
	}

	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch RSVP status")
		return
	}

	utils.SuccessResponse(c, 200, "RSVP status retrieved successfully", rsvp.ToResponse())
}

// GetAttendeesByStatus returns attendees filtered by status for an event (organizer only)
func (rc *RSVPController) GetAttendeesByStatus(c *gin.Context) {
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
		filter["status"] = models.RSVPStatus(status)
	}

	rsvpCollection := database.GetCollection("rsvps")
	cursor, err := rsvpCollection.Find(context.TODO(), filter)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to fetch attendees")
		return
	}
	defer cursor.Close(context.TODO())

	var rsvps []models.RSVPResponse
	if err = cursor.All(context.TODO(), &rsvps); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process attendees")
		return
	}

	if len(rsvps) == 0 {
		rsvps = []models.RSVPResponse{}
	}

	utils.SuccessResponse(c, 200, "Attendees retrieved successfully", rsvps)
}
