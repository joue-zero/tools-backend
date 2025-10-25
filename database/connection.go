package database

import (
	"context"
	"log"
	"time"

	"tools-backend/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database
var Client *mongo.Client

// Connect establishes connection to MongoDB (similar to Laravel's database connection)
func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.GetMongoURI()))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to ping MongoDB:", err)
	}

	Client = client
	DB = client.Database(config.GetDatabaseName())
	
	log.Println("Successfully connected to MongoDB!")
}

// GetCollection returns a MongoDB collection (similar to Laravel's DB::table())
func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if Client != nil {
		Client.Disconnect(ctx)
		log.Println("Disconnected from MongoDB")
	}
}
