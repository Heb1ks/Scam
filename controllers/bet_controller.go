package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"cs2-betting-platform/services"
	"cs2-betting-platform/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BetController struct {
	bettingService *services.BettingService
}

func NewBetController(bettingService *services.BettingService) *BetController {
	return &BetController{
		bettingService: bettingService,
	}
}

func (bc *BetController) PlaceBet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID       string  `json:"user_id"`
		MatchID      string  `json:"match_id"`
		SelectedTeam string  `json:"selected_team"`
		Amount       float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" || req.MatchID == "" || req.SelectedTeam == "" || req.Amount <= 0 {
		utils.RespondError(w, http.StatusBadRequest, "All fields are required and amount must be positive")
		return
	}

	userID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	matchID, err := primitive.ObjectIDFromHex(req.MatchID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	ctx := context.Background()
	bet, err := bc.bettingService.PlaceBet(ctx, userID, matchID, req.SelectedTeam, req.Amount)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Bet placed successfully",
		"bet":     bet,
	})
}

func (bc *BetController) GetUserBets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		utils.RespondError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	ctx := context.Background()
	bets, err := bc.bettingService.GetUserBets(ctx, userID)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch bets")
		return
	}

	utils.RespondJSON(w, http.StatusOK, bets)
}

func (bc *BetController) SettleMatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MatchID string `json:"match_id"`
		Winner  string `json:"winner"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	matchID, err := primitive.ObjectIDFromHex(req.MatchID)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	ctx := context.Background()
	err = bc.bettingService.SettleMatch(ctx, matchID, req.Winner)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Match settled successfully",
	})
}
