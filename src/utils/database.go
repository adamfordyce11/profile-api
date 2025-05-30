package utils

import (
	"context"
	"log"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectDB creates a connection to the MongoDB database and returns a reference to the client
func ConnectDB(uri string) (*mongo.Client, error) {
	//clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	clientOptions := options.Client().ApplyURI(uri).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
		panic(err)
	}
	//defer client.Disconnect(context.Background())
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Error pinging MongoDB: %v", err)
		panic(err)
	}
	log.Println("Connected to MongoDB!")
	return client, nil
}

// generateID generates a new UUID
func GenerateID() string {
	return uuid.New().String()
}
