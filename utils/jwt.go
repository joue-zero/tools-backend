package utils

import (
	"time"
	"tools-backend/config"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT generates a JWT token (similar to Laravel's JWT token generation)
func GenerateJWT(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetJWTSecret()))
}
