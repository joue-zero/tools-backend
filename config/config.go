package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env file (similar to Laravel's .env)
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// GetEnv gets environment variable with default value (similar to Laravel's env() helper)
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetMongoURI returns MongoDB connection string
func GetMongoURI() string {
	return GetEnv("MONGODB_URI", "mongodb://localhost:27017")
}

// GetDatabaseName returns database name
func GetDatabaseName() string {
	return GetEnv("DATABASE_NAME", "tools_db")
}

// GetJWTSecret returns JWT secret
func GetJWTSecret() string {
	return GetEnv("JWT_SECRET", "your-secret-key")
}
