package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"Scam/config"
	"Scam/models"
	"Scam/services"
	"Scam/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MATCH CONTROLLER

type MatchController struct {
	oddsService *services.OddsService
}

func NewMatchController(oddsService *services.OddsService) *MatchController {
	return &MatchController{oddsService: oddsService}
}

func (mc *MatchController) CreateMatch(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := context.Background()
	teamsColl := config.GetCollection("teams")

	// Получаем данные команд
	var teamA, teamB models.Team
	teamAID, _ := primitive.ObjectIDFromHex(req.TeamAID)
	teamBID, _ := primitive.ObjectIDFromHex(req.TeamBID)

	teamsColl.FindOne(ctx, bson.M{"_id": teamAID}).Decode(&teamA)
	teamsColl.FindOne(ctx, bson.M{"_id": teamBID}).Decode(&teamB)

	// Рассчитываем начальные коэффициенты
	pA, pB := mc.oddsService.CalculateProbabilities(teamA, teamB)
	oddA, oddB := mc.oddsService.CalculateDynamicOdds(pA, pB, 0, 0, 0.06)

	match := models.Match{
		ID: primitive.NewObjectID(),
		TeamA: models.TeamInMatch{
			ID:   teamAID,
			Name: teamA.TeamName,
			Logo: teamA.Logo,
			Odds: oddA,
		},
		TeamB: models.TeamInMatch{
			ID:   teamBID,
			Name: teamB.TeamName,
			Logo: teamB.Logo,
			Odds: oddB,
		},
		Tournament: req.Tournament,
		Format:     req.Format,
		Status:     models.MatchStatusUpcoming,
		StartTime:  req.StartTime,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	matchColl := config.GetCollection("matches")
	_, err := matchColl.InsertOne(ctx, match)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to create match")
		return
	}

	utils.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Match created successfully",
		"match":   match,
	})
}

func (mc *MatchController) GetMatches(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	ctx := context.Background()
	matchColl := config.GetCollection("matches")

	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	cursor, err := matchColl.Find(ctx, filter)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to fetch matches")
		return
	}
	defer cursor.Close(ctx)

	var matches []models.Match
	cursor.All(ctx, &matches)

	utils.RespondJSON(w, http.StatusOK, matches)
}

func (mc *MatchController) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchIDStr := r.URL.Query().Get("id")
	matchID, err := primitive.ObjectIDFromHex(matchIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	ctx := context.Background()
	matchColl := config.GetCollection("matches")

	var match models.Match
	err = matchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "Match not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, match)
}

// BET CONTROLLER

type BetController struct {
	bettingService *services.BettingService
}

func NewBetController(bettingService *services.BettingService) *BetController {
	return &BetController{bettingService: bettingService}
}

func (bc *BetController) PlaceBet(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  string  `json:"user_id"`
		MatchID string  `json:"match_id"`
		Team    string  `json:"team"`
		Amount  float64 `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID, _ := primitive.ObjectIDFromHex(req.UserID)
	matchID, _ := primitive.ObjectIDFromHex(req.MatchID)

	bet, err := bc.bettingService.PlaceBet(userID, matchID, req.Team, req.Amount)
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
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	bets, err := bc.bettingService.GetUserBets(userID)
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

	matchID, _ := primitive.ObjectIDFromHex(req.MatchID)

	err := bc.bettingService.SettleMatch(matchID, req.Winner)
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to settle match")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Match settled successfully",
	})
}
