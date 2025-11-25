# Event Management API Documentation

## Overview

This document outlines all the new event management features added to the tools-backend project. The system allows users to create events, manage attendees, track RSVPs, and perform advanced searches on events.

---

## Key Features Implemented

### 1. Event Management (Event Organizer Role Only)

- ✅ Create new events with title, date, time, location, and description
- ✅ View all events organized by the user
- ✅ View all events the user is invited to
- ✅ Invite others to events (organizer only)
- ✅ Delete events (organizer only)
- ✅ Update event details (organizer only)
- ✅ Users marked as "organizer" or "attendee" for each event

### 2. Response Management (RSVP)

- ✅ Attendees can indicate attendance status: Going, Maybe, Not Going
- ✅ Organizers can view all attendees and their RSVP statuses
- ✅ View RSVP counts by status
- ✅ Filter attendees by their response status

### 3. Search and Filtering

- ✅ Advanced search by keywords (event names, descriptions)
- ✅ Filter events by date range
- ✅ Filter events by location
- ✅ Filter events by user role (organizer/attendee)
- ✅ Combine multiple filters for complex searches

---

## Database Collections

### 1. events

```json
{
  "_id": ObjectId,
  "title": "string",
  "description": "string",
  "date": "YYYY-MM-DD",
  "time": "HH:MM",
  "location": "string",
  "participants": [
    {
      "user_id": ObjectId,
      "role": "organizer|attendee"
    }
  ],
  "created_at": ISODate,
  "updated_at": ISODate
}
```

### 2. rsvps

```json
{
  "_id": ObjectId,
  "event_id": ObjectId,
  "user_id": ObjectId,
  "status": "going|maybe|not_going",
  "created_at": ISODate,
  "updated_at": ISODate
}
```

---

## API Endpoints

### Authentication (Already Existing)

- `POST /api/v1/register` - Register a new user
- `POST /api/v1/login` - Login user and get JWT token

---

### Event Management Endpoints

#### 1. Create Event

```
POST /api/v1/events
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Team Meeting",
  "description": "Quarterly team sync meeting to discuss goals and progress",
  "date": "2024-12-20",
  "time": "14:00",
  "location": "Conference Room A"
}

Response: 201 Created
{
  "success": true,
  "message": "Event created successfully",
  "data": {
    "id": "ObjectId",
    "title": "Team Meeting",
    "description": "...",
    "date": "2024-12-20",
    "time": "14:00",
    "location": "Conference Room A",
    "participants": [
      {
        "user_id": "userId",
        "role": "organizer"
      }
    ],
    "created_at": "2024-11-25T10:00:00Z",
    "updated_at": "2024-11-25T10:00:00Z"
  }
}
```

#### 2. Get Event by ID

```
GET /api/v1/events/:id
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "Event retrieved successfully",
  "data": { /* event object */ }
}
```

#### 3. Get Organized Events

```
GET /api/v1/events/organized
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "Organized events retrieved successfully",
  "data": [ /* array of events where user is organizer */ ]
}
```

#### 4. Get Invited Events

```
GET /api/v1/events/invited
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "Invited events retrieved successfully",
  "data": [ /* array of events where user is attendee */ ]
}
```

#### 5. Update Event

```
PUT /api/v1/events/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Meeting Title",
  "date": "2024-12-21",
  "time": "15:00"
}

Note: Only organizer can update. All fields are optional.

Response: 200 OK
{
  "success": true,
  "message": "Event updated successfully",
  "data": null
}
```

#### 6. Delete Event

```
DELETE /api/v1/events/:id
Authorization: Bearer <token>

Note: Only organizer can delete. Also deletes all related RSVPs.

Response: 200 OK
{
  "success": true,
  "message": "Event deleted successfully",
  "data": null
}
```

#### 7. Invite Users to Event

```
POST /api/v1/events/:id/invite
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_ids": ["userId1", "userId2", "userId3"]
}

Note: Only organizer can invite. Users become attendees.

Response: 200 OK
{
  "success": true,
  "message": "Users invited successfully",
  "data": {
    "invited_count": 3
  }
}
```

---

### RSVP Management Endpoints

#### 1. Create/Update RSVP Response

```
POST /api/v1/events/:id/rsvp
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "going"  // or "maybe", "not_going"
}

Response: 201 Created (new) or 200 OK (updated)
{
  "success": true,
  "message": "RSVP created/updated successfully",
  "data": {
    "id": "ObjectId",
    "event_id": "eventId",
    "user_id": "userId",
    "status": "going",
    "created_at": "2024-11-25T10:00:00Z",
    "updated_at": "2024-11-25T10:00:00Z"
  }
}
```

#### 2. Get User's RSVP Status for an Event

```
GET /api/v1/events/:id/rsvp/status
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "RSVP status retrieved successfully",
  "data": {
    "id": "ObjectId",
    "event_id": "eventId",
    "user_id": "userId",
    "status": "going"
  }
}
```

#### 3. Get All Event Attendees and Summary (Organizer Only)

```
GET /api/v1/events/:id/attendees
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "Attendees retrieved successfully",
  "data": {
    "total": 10,
    "going": 7,
    "maybe": 2,
    "not_going": 1,
    "no_response": 0,
    "attendees": [
      {
        "id": "ObjectId",
        "event_id": "eventId",
        "user_id": "userId",
        "status": "going",
        "created_at": "2024-11-25T10:00:00Z",
        "updated_at": "2024-11-25T10:00:00Z"
      }
    ]
  }
}
```

#### 4. Get Attendees Filtered by Status (Organizer Only)

```
GET /api/v1/events/:id/attendees/status?status=going
Authorization: Bearer <token>

Query Parameters:
- status (required): "going", "maybe", or "not_going"

Response: 200 OK
{
  "success": true,
  "message": "Attendees retrieved successfully",
  "data": [ /* array of RSVPs with specified status */ ]
}
```

---

### Search and Filtering Endpoints

#### 1. Advanced Search (POST with Filters)

```
POST /api/v1/search
Authorization: Bearer <token>
Content-Type: application/json

{
  "keyword": "team",
  "start_date": "2024-12-01",
  "end_date": "2024-12-31",
  "user_role": "organizer",
  "location": "Conference Room"
}

Note: All fields are optional. Combine for complex filtering.

Response: 200 OK
{
  "success": true,
  "message": "Search completed successfully",
  "data": {
    "total_results": 5,
    "events": [ /* matching events */ ]
  }
}
```

#### 2. Advanced Search (GET with Query Parameters)

```
GET /api/v1/search/advanced?keyword=team&start_date=2024-12-01&end_date=2024-12-31&user_role=organizer&location=Conference
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "Advanced search completed successfully",
  "data": {
    "filters": { /* applied filters */ },
    "total_results": 5,
    "events": [ /* matching events */ ]
  }
}
```

#### 3. Get All User Events

```
GET /api/v1/all-events
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "message": "All user events retrieved successfully",
  "data": [ /* all events (organized + invited) */ ]
}
```

#### 4. Search by Keyword

```
GET /api/v1/search/keyword?q=meeting
Authorization: Bearer <token>

Query Parameters:
- q (required): Search term to match in title or description

Response: 200 OK
{
  "success": true,
  "message": "Events filtered by keyword successfully",
  "data": {
    "keyword": "meeting",
    "results": [ /* matching events */ ]
  }
}
```

#### 5. Filter Events by Date Range

```
GET /api/v1/search/date?start_date=2024-12-01&end_date=2024-12-31
Authorization: Bearer <token>

Query Parameters:
- start_date (optional): ISO 8601 format (YYYY-MM-DD)
- end_date (optional): ISO 8601 format (YYYY-MM-DD)

Response: 200 OK
{
  "success": true,
  "message": "Events filtered by date successfully",
  "data": [ /* events in date range */ ]
}
```

#### 6. Filter Events by User Role

```
GET /api/v1/search/role?role=organizer
Authorization: Bearer <token>

Query Parameters:
- role (required): "organizer" or "attendee"

Response: 200 OK
{
  "success": true,
  "message": "Events filtered by role successfully",
  "data": [ /* events where user has specified role */ ]
}
```

---

## Request/Response Examples

### Example 1: Complete Event Creation and Invitation Flow

```bash
# 1. User 1 creates an event
POST /api/v1/events
{
  "title": "Project Kickoff",
  "description": "Kickoff meeting for Q1 project",
  "date": "2024-12-20",
  "time": "10:00",
  "location": "Room 101"
}

# 2. User 1 invites User 2 and User 3
POST /api/v1/events/{eventId}/invite
{
  "user_ids": ["user2Id", "user3Id"]
}

# 3. User 2 responds with "Going"
POST /api/v1/events/{eventId}/rsvp
{
  "status": "going"
}

# 4. User 3 responds with "Maybe"
POST /api/v1/events/{eventId}/rsvp
{
  "status": "maybe"
}

# 5. User 1 views all attendees and their statuses
GET /api/v1/events/{eventId}/attendees
```

### Example 2: Advanced Search

```bash
# Search for all events in December where user is organizer, with "project" in title
GET /api/v1/search/advanced?keyword=project&start_date=2024-12-01&end_date=2024-12-31&user_role=organizer
```

---

## Error Responses

All endpoints follow consistent error handling:

```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

### Common Error Codes:

- `400` - Bad Request (invalid data or missing required fields)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (user doesn't have permission)
- `404` - Not Found (event or resource doesn't exist)
- `500` - Internal Server Error

---

## Authorization & Security

- **All protected endpoints** require a valid JWT token in the Authorization header
- **Event creators** are automatically marked as "organizers"
- **Only organizers** can:
  - Invite users to events
  - Update event details
  - Delete events
  - View attendee list and RSVP statuses
- **Attendees** can:
  - View event details
  - Update their own RSVP status
  - View their own response

---

## Data Validation

### Event Fields:

- `title`: 3-200 characters (required)
- `description`: 10-2000 characters (required)
- `date`: ISO 8601 format YYYY-MM-DD (required)
- `time`: HH:MM format (required)
- `location`: 5-500 characters (required)

### RSVP Status:

- Must be one of: "going", "maybe", "not_going"

### User Role:

- Must be one of: "organizer", "attendee"

---

## Usage Tips

1. **Always include Authorization header** with Bearer token for protected endpoints
2. **Date format** must be ISO 8601 (YYYY-MM-DD)
3. **Time format** must be 24-hour (HH:MM)
4. **Search is case-insensitive** for keywords and locations
5. **Combine filters** in advanced search for more specific results
6. **RSVP status** is per-user per-event (creating/updating works the same way)

---

## Implementation Notes

- Models created in `/models/event.go`
- Event Controller logic in `/controllers/event_controller.go`
- RSVP Controller logic in `/controllers/rsvp_controller.go`
- Search Controller logic in `/controllers/search_controller.go`
- All routes configured in `/routes/routes.go`
- Uses existing utility functions: `ValidateStruct()`, `SuccessResponse()`, `ErrorResponse()`, `ValidationErrorResponse()`
- Authentication middleware applied to all protected routes
- MongoDB collections: `events` and `rsvps`
