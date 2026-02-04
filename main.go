package main

import (
	"log"
	"net/http"
	"time"

	"cs2-betting-platform/config"
)

func main() {

	if err := config.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer config.DisconnectDB()

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("CS2 Betting Platform started on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
