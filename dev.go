//go:build dev
// +build dev

package main


import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Badge struct {
	User  string   `bson:"User"`
	Flags []string `bson:"FLAGS"`
}

func main() {
	// Check if user ID is provided
	if len(os.Args) < 2 {
		fmt.Println("[ERROR] >> Developer Badge >> Please provide a member ID!")
		os.Exit(1)
	}
	memberID := os.Args[1]

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("[ERROR] >> Developer Badge >> Failed to load .env file: %v", err)
	}

	mongoToken := os.Getenv("MONGO_TOKEN")
	if mongoToken == "" {
		log.Fatal("[ERROR] >> Developer Badge >> MONGO_TOKEN is not set in .env file")
	}

	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoToken))
	if err != nil {
		log.Fatalf("[ERROR] >> Developer Badge >> Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("[ERROR] >> Developer Badge >> Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	fmt.Println("[SUCCESS] >> Developer Badge >> Connected to the database!")

	// Select database and collection
	collection := client.Database("yourDatabaseName").Collection("yourCollectionName") // Replace with your actual database and collection names

	// Search for the user in the database
	var badge Badge
	err = collection.FindOne(ctx, bson.M{"User": memberID}).Decode(&badge)
	if err == mongo.ErrNoDocuments {
		// Create a new document if user doesn't exist
		badge = Badge{
			User:  memberID,
			Flags: []string{"DEVELOPER"},
		}
		_, err := collection.InsertOne(ctx, badge)
		if err != nil {
			log.Fatalf("[ERROR] >> Developer Badge >> Failed to insert new user: %v", err)
		}
		fmt.Printf("[SUCCESS] >> Developer Badge has been added to the user: %s\n", memberID)
	} else if err != nil {
		log.Fatalf("[ERROR] >> Developer Badge >> Error finding user: %v", err)
	} else {
		// Update existing user's FLAGS
		badge.Flags = append(badge.Flags, "DEVELOPER")
		_, err := collection.UpdateOne(ctx, bson.M{"User": memberID}, bson.M{"$set": bson.M{"FLAGS": badge.Flags}})
		if err != nil {
			log.Fatalf("[ERROR] >> Developer Badge >> Failed to update user: %v", err)
		}
		fmt.Printf("[SUCCESS] >> Developer Badge has been added to the user: %s\n", memberID)
	}
}
