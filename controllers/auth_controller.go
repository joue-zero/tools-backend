package controllers

import (
	"context"
	// "log"
	"time"
	"tools-backend/database"
	"tools-backend/models"
	"tools-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	// "io"
)

type AuthController struct{}

// Register handles user registration (similar to Laravel's register method)
func (ac *AuthController) Register(c *gin.Context) {
	var req models.UserRegistrationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Validate registration data
	if errors := utils.ValidateStruct(req); len(errors) > 0 {
		utils.ValidationErrorResponse(c, errors)
		return
	}

	// Check if user already exists
	collection := database.GetCollection("users")
	var existingUser models.User
	err := collection.FindOne(context.TODO(), bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, 400, "User with this email already exists")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to hash password")
		return
	}

	// Create user from request data
	user := models.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert user
	result, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to create user")
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	utils.SuccessResponse(c, 201, "User created successfully", user.ToResponse())
}

// Login handles user login (similar to Laravel's login method)
func (ac *AuthController) Login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		utils.ErrorResponse(c, 400, "Invalid request data")
		return
	}

	// Validate login data
	if errors := utils.ValidateStruct(loginData); len(errors) > 0 {
		utils.ValidationErrorResponse(c, errors)
		return
	}

	// Find user
	collection := database.GetCollection("users")
	var user models.User
	err := collection.FindOne(context.TODO(), bson.M{"email": loginData.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.ErrorResponse(c, 401, "Invalid credentials")
		} else {
			utils.ErrorResponse(c, 500, "Database error")
		}
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		utils.ErrorResponse(c, 401, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, 200, "Login successful", gin.H{
		"user":  user.ToResponse(),
		"token": token,
	})
}
