# Quick Reference: Event Management API Endpoints

## All Endpoints at a Glance

### 🔐 Authentication (No Token Required)

```
POST /api/v1/register
POST /api/v1/login
```

---

### 📅 Event Management Routes (Token Required)

| Method | Endpoint                    | Description                | Who Can Use           |
| ------ | --------------------------- | -------------------------- | --------------------- |
| POST   | `/api/v1/events`            | Create new event           | All logged-in users   |
| GET    | `/api/v1/events/:id`        | Get event details          | All users with access |
| GET    | `/api/v1/events/organized`  | View my organized events   | All users             |
| GET    | `/api/v1/events/invited`    | View events I'm invited to | All users             |
| PUT    | `/api/v1/events/:id`        | Update event               | Organizer only        |
| DELETE | `/api/v1/events/:id`        | Delete event               | Organizer only        |
| POST   | `/api/v1/events/:id/invite` | Invite users to event      | Organizer only        |

---

### 📋 Event Status Management Routes (Token Required)

| Method | Endpoint                                           | Description                  | Who Can Use      |
| ------ | -------------------------------------------------- | ---------------------------- | ---------------- |
| POST   | `/api/v1/events/:id/status`                        | Submit/update event status   | Event attendees  |
| GET    | `/api/v1/events/:id/status`                        | Get my event status          | All participants |
| GET    | `/api/v1/events/:id/attendees`                     | View all attendees + summary | Organizer only   |
| GET    | `/api/v1/events/:id/attendees/status?status=going` | Get attendees by status      | Organizer only   |

---

### 🔍 Search & Filtering Routes (Token Required)

| Method | Endpoint                             | Description             | Query Params                                       |
| ------ | ------------------------------------ | ----------------------- | -------------------------------------------------- |
| POST   | `/api/v1/search`                     | Advanced search (body)  | keyword, start_date, end_date, user_role, location |
| GET    | `/api/v1/search/advanced`            | Advanced search (query) | keyword, start_date, end_date, user_role, location |
| POST   | `/api/v1/search/advanced`            | Advanced search (body)  | Same as above                                      |
| GET    | `/api/v1/all-events`                 | Get all my events       | None                                               |
| GET    | `/api/v1/search/keyword?q=meeting`   | Search by keyword       | q (required)                                       |
| GET    | `/api/v1/search/date`                | Filter by date          | start_date, end_date                               |
| GET    | `/api/v1/search/role?role=organizer` | Filter by role          | role (required)                                    |
| GET    | `/api/v1/users/search?q=john`        | Search users to invite  | q (required)                                       |

---

## cURL Examples

### 1. Create an Event

```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Meeting",
    "description": "Quarterly sync meeting",
    "date": "2024-12-20",
    "time": "14:00",
    "location": "Conference Room A"
  }'
```

### 2. Get Organized Events

```bash
curl -X GET http://localhost:8080/api/v1/events/organized \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. Invite Users to Event

```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/invite \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": ["user1Id", "user2Id"]
  }'
```

### 4. Submit Event Status Response

```bash
curl -X POST http://localhost:8080/api/v1/events/{eventId}/status \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "going"}'
```

### 5. Get Attendees and Summary

```bash
curl -X GET http://localhost:8080/api/v1/events/{eventId}/attendees \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 6. Search by Keyword

```bash
curl -X GET "http://localhost:8080/api/v1/search/keyword?q=meeting" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 7. Advanced Search with Multiple Filters

```bash
curl -X GET "http://localhost:8080/api/v1/search/advanced?keyword=meeting&start_date=2024-12-01&end_date=2024-12-31&user_role=organizer" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 8. Filter by Date Range

```bash
curl -X GET "http://localhost:8080/api/v1/search/date?start_date=2024-12-01&end_date=2024-12-31" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 9. Filter by User Role

```bash
curl -X GET "http://localhost:8080/api/v1/search/role?role=organizer" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 10. Get Attendees by Response Status

```bash
curl -X GET "http://localhost:8080/api/v1/events/{eventId}/attendees/status?status=going" \
```

### 11. Search Users to Invite

```bash
curl -X GET "http://localhost:8080/api/v1/users/search?q=john" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Response Status Codes

| Code | Meaning                         |
| ---- | ------------------------------- |
| 200  | Success                         |
| 201  | Created                         |
| 400  | Bad Request (validation error)  |
| 401  | Unauthorized (no/invalid token) |
| 403  | Forbidden (no permission)       |
| 404  | Not Found                       |
| 500  | Server Error                    |

---

## Event Status Values

- `going` - Will attend
- `maybe` - Might attend
- `not_going` - Won't attend

---

## User Role Values in Events

- `organizer` - Event creator/organizer
- `attendee` - Invited participant

---

## Required Headers for Protected Endpoints

```
Authorization: Bearer {JWT_TOKEN}
Content-Type: application/json (for POST/PUT requests)
```

---

## Common Request Patterns

### Create Event

```json
{
  "title": "string",
  "description": "string",
  "date": "YYYY-MM-DD",
  "time": "HH:MM",
  "location": "string"
}
```

### Update Event (all fields optional)

```json
{
  "title": "string",
  "description": "string",
  "date": "YYYY-MM-DD",
  "time": "HH:MM",
  "location": "string"
}
```

### Invite Users

```json
{
  "user_ids": ["userId1", "userId2"]
}
```

### Submit Event Status

```json
{
  "status": "going|maybe|not_going"
}
```

### Advanced Search

```json
{
  "keyword": "string",
  "start_date": "YYYY-MM-DD",
  "end_date": "YYYY-MM-DD",
  "user_role": "organizer|attendee",
  "location": "string"
}
```

---

## Common Response Format

### Success Response

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    /* response data */
  }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Error description",
  "data": null
}
```

### Validation Error Response

```json
{
  "success": false,
  "message": "Validation failed",
  "errors": {
    "field_name": "error message"
  }
}
```

---

## Workflow Example: Complete Event Flow

```
1. User A: POST /api/v1/register → Get account
2. User A: POST /api/v1/login → Get JWT token
3. User A: POST /api/v1/events → Create event (becomes organizer)
4. User A: POST /api/v1/events/{id}/invite → Invite User B & C
5. User B: POST /api/v1/events/{id}/status → Submit "going"
6. User C: POST /api/v1/events/{id}/status → Submit "maybe"
7. User A: GET /api/v1/events/{id}/attendees → View summary (1 going, 1 maybe)
8. User A: GET /api/v1/events/{id}/attendees/status?status=going → See who's going
9. User A: PUT /api/v1/events/{id} → Update event details
10. User A: DELETE /api/v1/events/{id} → Delete event (removes event statuses too)
```

---

## Testing Tips

1. **Use Postman** - Import collection and test endpoints
2. **Keep JWT token** - Copy from login response, use in all protected endpoints
3. **Use MongoDB ObjectIds** - Valid format for IDs: `507f1f77bcf86cd799439011`
4. **Test permissions** - Try as both organizer and attendee
5. **Check response format** - All responses follow standard JSON format
6. **Validate dates** - Use ISO 8601 format (YYYY-MM-DD)
