package routes

import (
	"net/http"

	"Scam/controllers"
	"Scam/services"
)

func SetupRoutes(oddsService *services.OddsService, bettingService *services.BettingService) *http.ServeMux {
	mux := http.NewServeMux()

	// Создаём контроллеры
	userController := controllers.NewUserController()
	matchController := controllers.NewMatchController(oddsService)
	betController := controllers.NewBetController(bettingService)

	// =============== USER ROUTES ===============
	mux.HandleFunc("/api/users/register", userController.Register)
	mux.HandleFunc("/api/users/login", userController.Login)
	mux.HandleFunc("/api/users/profile", userController.GetProfile)
	mux.HandleFunc("/api/users/balance", userController.GetBalance)
	mux.HandleFunc("/api/users/add-balance", userController.AddBalance)

	// =============== MATCH ROUTES ===============
	mux.HandleFunc("/api/matches", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			matchController.CreateMatch(w, r)
		} else if r.Method == http.MethodGet {
			matchController.GetMatches(w, r)
		}
	})
	mux.HandleFunc("/api/matches/details", matchController.GetMatch)

	// =============== BET ROUTES ===============
	mux.HandleFunc("/api/bets/place", betController.PlaceBet)
	mux.HandleFunc("/api/bets/user", betController.GetUserBets)
	mux.HandleFunc("/api/bets/settle", betController.SettleMatch)

	// =============== HEALTH CHECK ===============
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// =============== STATIC FILES ===============
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// =============== INDEX PAGE ===============
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./templates/index.html")
		}
	})

	return mux
}
