package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventRole represents the role of a user in an event (organizer or attendee)
type EventRole string

const (
	RoleOrganizer EventRole = "organizer"
	RoleAttendee  EventRole = "attendee"
)

// RSVPStatus represents the attendance status for an event
type RSVPStatus string

const (
	StatusGoing    RSVPStatus = "going"
	StatusMaybe    RSVPStatus = "maybe"
	StatusNotGoing RSVPStatus = "not_going"
)

// EventParticipant represents a user's participation in an event
type EventParticipant struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id"`
	Role   EventRole          `json:"role" bson:"role" validate:"required,oneof=organizer attendee"`
}

// RSVP represents an attendee's response to an event invitation
type RSVP struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EventID   primitive.ObjectID `json:"event_id" bson:"event_id"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Status    RSVPStatus         `json:"status" bson:"status" validate:"required,oneof=going maybe not_going"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// Event represents an event in the system
type Event struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title" validate:"required,min=3,max=200"`
	Description string             `json:"description" bson:"description" validate:"required,min=10,max=2000"`
	Date        string             `json:"date" bson:"date" validate:"required"` // ISO 8601 format: YYYY-MM-DD
	Time        string             `json:"time" bson:"time" validate:"required"` // HH:MM format
	Location    string             `json:"location" bson:"location" validate:"required,min=5,max=500"`
	Participants []EventParticipant `json:"participants" bson:"participants"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateEventRequest represents the data for creating an event
type CreateEventRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description" validate:"required,min=10,max=2000"`
	Date        string `json:"date" validate:"required"` // ISO 8601 format: YYYY-MM-DD
	Time        string `json:"time" validate:"required"` // HH:MM format
	Location    string `json:"location" validate:"required,min=5,max=500"`
}

// InviteToEventRequest represents a request to invite users to an event
type InviteToEventRequest struct {
	UserIDs []primitive.ObjectID `json:"user_ids" validate:"required,min=1"`
}

// UpdateEventRequest represents the data for updating an event
type UpdateEventRequest struct {
	Title       string `json:"title" validate:"min=3,max=200"`
	Description string `json:"description" validate:"min=10,max=2000"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Location    string `json:"location" validate:"min=5,max=500"`
}

// RSVPRequest represents a request to update RSVP status
type RSVPRequest struct {
	Status RSVPStatus `json:"status" validate:"required,oneof=going maybe not_going"`
}

// EventResponse represents an event sent in API responses
type EventResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Date        string             `json:"date"`
	Time        string             `json:"time"`
	Location    string             `json:"location"`
	Participants []EventParticipant `json:"participants"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// RSVPResponse represents an RSVP sent in API responses
type RSVPResponse struct {
	ID        primitive.ObjectID `json:"id"`
	EventID   primitive.ObjectID `json:"event_id"`
	UserID    primitive.ObjectID `json:"user_id"`
	Status    RSVPStatus         `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// ToResponse converts Event to EventResponse
func (e *Event) ToResponse() EventResponse {
	return EventResponse{
		ID:          e.ID,
		Title:       e.Title,
		Description: e.Description,
		Date:        e.Date,
		Time:        e.Time,
		Location:    e.Location,
		Participants: e.Participants,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// ToResponse converts RSVP to RSVPResponse
func (r *RSVP) ToResponse() RSVPResponse {
	return RSVPResponse{
		ID:        r.ID,
		EventID:   r.EventID,
		UserID:    r.UserID,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
