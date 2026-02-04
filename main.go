package main

import (
	"log"
	"net/http"
	"time"

	"cs2-betting-platform/config"
	"cs2-betting-platform/routes"
	"cs2-betting-platform/services"
)

func main() {

	if err := config.ConnectDB(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer config.DisconnectDB()

	oddsService := services.NewOddsService()
	bettingService := services.NewBettingService(oddsService)

	oddsService.StartOddsUpdateWorker(30 * time.Second)

	router := routes.SetupRoutes(oddsService, bettingService)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println(" CS2 Betting Platform started on http://localhost:8080")
	log.Println(" API Documentation:")
	log.Println("   POST   /api/users/register      - Register new user")
	log.Println("   POST   /api/users/login         - Login user")
	log.Println("   GET    /api/users/profile       - Get user profile")
	log.Println("   GET    /api/users/balance       - Get user balance")
	log.Println("   POST   /api/users/add-balance   - Add balance (testing)")
	log.Println("   POST   /api/matches             - Create match")
	log.Println("   GET    /api/matches             - Get all matches")
	log.Println("   GET    /api/matches/details     - Get match details")
	log.Println("   PUT    /api/matches/status      - Update match status")
	log.Println("   POST   /api/bets/place          - Place bet")
	log.Println("   GET    /api/bets/user           - Get user bets")
	log.Println("   POST   /api/bets/settle         - Settle match")
	log.Println("   GET    /api/health              - Health check")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
