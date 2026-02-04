package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB *mongo.Database

func ConnectMongo() {
	uri := os.Getenv("MONGO_URI")
	dbName := os.Getenv("DB_NAME")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("mongo connect error:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("mongo ping error:", err)
	}

	DB = client.Database(dbName)
	log.Println("mongoDB connected:", dbName)
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
