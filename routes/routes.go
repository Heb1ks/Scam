package routes

import (
	"cs2-betting-platform/controllers"
	"cs2-betting-platform/services"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoutes(oddsService *services.OddsService, bettingService *services.BettingService) *mux.Router {
	router := mux.NewRouter()

	userController := controllers.NewUserController()
	matchController := controllers.NewMatchController(oddsService)
	betController := controllers.NewBetController(bettingService)

	router.HandleFunc("/api/users/register", userController.Register).Methods("POST")
	router.HandleFunc("/api/users/login", userController.Login).Methods("POST")
	router.HandleFunc("/api/users/profile", userController.GetUser).Methods("GET")
	router.HandleFunc("/api/users/balance", userController.GetBalance).Methods("GET")
	router.HandleFunc("/api/users/add-balance", userController.AddBalance).Methods("POST")

	router.HandleFunc("/api/matches", matchController.CreateMatch).Methods("POST")
	router.HandleFunc("/api/matches", matchController.GetMatches).Methods("GET")
	router.HandleFunc("/api/matches/details", matchController.GetMatch).Methods("GET")
	router.HandleFunc("/api/matches/status", matchController.UpdateMatchStatus).Methods("PUT")

	router.HandleFunc("/api/bets/place", betController.PlaceBet).Methods("POST")
	router.HandleFunc("/api/bets/user", betController.GetUserBets).Methods("GET")
	router.HandleFunc("/api/bets/settle", betController.SettleMatch).Methods("POST")

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}).Methods("GET")

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("/", serveIndex).Methods("GET")

	return router
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./templates/index.html")
}
