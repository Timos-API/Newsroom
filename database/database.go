package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBDriver struct {
	Collection mongo.Collection
}

var Database *mongo.Database
var client *mongo.Client

func Connect() {

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	c, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to MongoDB")

	client = c
	Database = c.Database("TimosAPI")
}

func Disconnect() {
	if client != nil {
		client.Disconnect(context.Background())
	}
}
