package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cs2-betting-platform/config"
	"cs2-betting-platform/models"
	"cs2-betting-platform/services"
	"cs2-betting-platform/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MatchController struct {
	oddsService *services.OddsService
}

func NewMatchController(oddsService *services.OddsService) *MatchController {
	return &MatchController{
		oddsService: oddsService,
	}
}

func (mc *MatchController) CreateMatch(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.TeamAName == "" || req.TeamBName == "" || req.Tournament == "" {
		utils.RespondError(w, http.StatusBadRequest, "Team names and tournament are required")
		return
	}

	oddsA, oddsB := mc.oddsService.CalculateInitialOdds()

	match := models.Match{
		ID: primitive.NewObjectID(),
		TeamA: models.Team{
			Name:      req.TeamAName,
			Logo:      req.TeamALogo,
			Odds:      oddsA,
			TotalBets: 0,
		},
		TeamB: models.Team{
			Name:      req.TeamBName,
			Logo:      req.TeamBLogo,
			Odds:      oddsB,
			TotalBets: 0,
		},
		Tournament: req.Tournament,
		Format:     req.Format,
		Status:     models.MatchStatusUpcoming,
		StartTime:  req.StartTime,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	matchColl := config.GetCollection("matches")
	ctx := context.Background()

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

	matchColl := config.GetCollection("matches")
	ctx := context.Background()

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
	if err := cursor.All(ctx, &matches); err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "Failed to decode matches")
		return
	}

	utils.RespondJSON(w, http.StatusOK, matches)
}

func (mc *MatchController) GetMatch(w http.ResponseWriter, r *http.Request) {
	matchIDStr := r.URL.Query().Get("id")
	if matchIDStr == "" {
		utils.RespondError(w, http.StatusBadRequest, "Match ID is required")
		return
	}

	matchID, err := primitive.ObjectIDFromHex(matchIDStr)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid match ID")
		return
	}

	matchColl := config.GetCollection("matches")
	ctx := context.Background()

	var match models.Match
	err = matchColl.FindOne(ctx, bson.M{"_id": matchID}).Decode(&match)
	if err != nil {
		utils.RespondError(w, http.StatusNotFound, "Match not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, match)
}

func (mc *MatchController) UpdateMatchStatus(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MatchID string             `json:"match_id"`
		Status  models.MatchStatus `json:"status"`
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

	matchColl := config.GetCollection("matches")
	ctx := context.Background()

	result, err := matchColl.UpdateOne(
		ctx,
		bson.M{"_id": matchID},
		bson.M{
			"$set": bson.M{
				"status":     req.Status,
				"updated_at": time.Now(),
			},
		},
	)

	if err != nil || result.MatchedCount == 0 {
		utils.RespondError(w, http.StatusNotFound, "Match not found")
		return
	}

	utils.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Match status updated successfully",
	})
}
