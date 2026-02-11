package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB     *mongo.Database
	Client *mongo.Client
)

func ConnectDB() error {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // fallback для разработки
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "cs2_betting" // fallback
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	Client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	err = Client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	DB = Client.Database(dbName)
	log.Printf("✅ Connected to MongoDB: %s", dbName)

	// Создаём индексы для оптимизации запросов
	createIndexes()

	return nil
}

func createIndexes() {
	ctx := context.Background()

	// Индексы для коллекции users
	usersCollection := DB.Collection("users")
	usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"email": 1},
		Options: options.Index().SetUnique(true),
	})
	usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    map[string]interface{}{"username": 1},
		Options: options.Index().SetUnique(true),
	})

	// Индексы для коллекции matches
	matchesCollection := DB.Collection("matches")
	matchesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"status": 1, "startTime": -1},
	})

	// Индексы для коллекции bets
	betsCollection := DB.Collection("bets")
	betsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"userId": 1, "createdAt": -1},
	})
	betsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{"matchId": 1},
	})

	log.Println("📊 Database indexes created")
}

func DisconnectDB() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("❌ Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("👋 Disconnected from MongoDB")
		}
	}
}

func GetCollection(collectionName string) *mongo.Collection {
	return DB.Collection(collectionName)
}
