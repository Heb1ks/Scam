package main

import (
	"log"
	"os"

	"Scam/database"
	"Scam/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not founf")
	}

	database.ConnectMongo()

	r := gin.Default()

	api := r.Group("/api")
	routes.TeamRoutes(api)
	routes.OddsRoutes(api)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("server port:", port)
	r.Run(":" + port)
}
