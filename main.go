package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"Scam/config"
	"Scam/routes"
	"Scam/services"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env file not found, using defaults")
	}

	// Подключаемся к MongoDB
	if err := config.ConnectDB(); err != nil {
		log.Fatal("❌ Failed to connect to MongoDB:", err)
	}
	defer config.DisconnectDB()

	// Создаём сервисы
	oddsService := services.NewOddsService()
	bettingService := services.NewBettingService(oddsService)

	// Запускаем воркер для обновления коэффициентов каждые 30 секунд
	oddsService.StartOddsUpdateWorker(30 * time.Second)

	// Настраиваем маршруты
	router := routes.SetupRoutes(oddsService, bettingService)

	// Настраиваем сервер
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Выводим информацию

	log.Printf("Server started on http://localhost:%s\n", port)
	log.Println("API Endpoints:")
	log.Println("   POST   /api/users/register")
	log.Println("   POST   /api/users/login")
	log.Println("   GET    /api/users/profile")
	log.Println("   POST   /api/matches")
	log.Println("   GET    /api/matches")
	log.Println("   POST   /api/bets/place")
	log.Println("   GET    /api/bets/user")
	log.Println("   POST   /api/bets/settle")

	// Запускаем сервер
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
