package controllers

import (
	"context"
	"tools-backend/database"
	"tools-backend/models"
	"tools-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserController struct{}

// SearchUsers searches for users by name or email
func (uc *UserController) SearchUsers(c *gin.Context) {
	query := c.Query("q")

	// If query is empty, return empty list or require at least 1 char?
	// Best practice: don't return all users if query is empty unless specifically requested/paginated.
	// Let's require at least 1 character to start searching.
	if len(query) < 1 {
		utils.SuccessResponse(c, 200, "Enter a search term", []models.UserResponse{})
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

	currentUserID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.ErrorResponse(c, 400, "Invalid user ID")
		return
	}

	collection := database.GetCollection("users")

	// Case-insensitive regex search for name or email
	regexPattern := bson.M{"$regex": query, "$options": "i"}

	filter := bson.M{
		"_id": bson.M{"$ne": currentUserID}, // Exclude current user
		"$or": bson.A{
			bson.M{"name": regexPattern},
			bson.M{"email": regexPattern},
		},
	}

	// Limit results to 20 to prevent massive payloads
	findOptions := options.Find().SetLimit(20)

	cursor, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to search users")
		return
	}
	defer cursor.Close(context.TODO())

	var users []models.User
	if err = cursor.All(context.TODO(), &users); err != nil {
		utils.ErrorResponse(c, 500, "Failed to process users")
		return
	}

	// Convert to UserResponse
	userResponses := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	utils.SuccessResponse(c, 200, "Users found", userResponses)
}
