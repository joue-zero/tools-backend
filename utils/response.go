package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a success response (similar to Laravel's response()->json())
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse sends an error response (similar to Laravel's response()->json())
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, gin.H{
		"success": false,
		"message": message,
		"data":    nil,
	})
}

// ValidationErrorResponse sends validation error response (similar to Laravel's validation errors)
func ValidationErrorResponse(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation failed",
		"errors":  errors,
	})
}
