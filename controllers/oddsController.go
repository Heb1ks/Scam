package controllers

import (
	"context"
	"math"
	"net/http"
	"time"

	"Scam/database"
	"Scam/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func round2(x float64) float64 {
	return math.Round(x*100) / 100
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

// мейн формула
func calculateProbability(teamA, teamB models.Team) (float64, float64) {
	// weights
	wRank := 0.55
	wMap := 0.25
	wForm := 0.20

	//  ранк скор
	rankPowerA := 1.0 / float64(teamA.HLTVRank)
	rankPowerB := 1.0 / float64(teamB.HLTVRank)
	scoreRankA := rankPowerA / (rankPowerA + rankPowerB)

	// форм скор
	formPowerA := float64(teamA.FormWins) / 5.0
	formPowerB := float64(teamB.FormWins) / 5.0
	if formPowerA+formPowerB == 0 {
		formPowerA = 0.5
		formPowerB = 0.5
	}
	scoreFormA := formPowerA / (formPowerA + formPowerB)

	// скор маппула
	mapPowerA := (teamA.MapPool.Mirage + teamA.MapPool.Nuke + teamA.MapPool.Ancient) / 3.0
	mapPowerB := (teamB.MapPool.Mirage + teamB.MapPool.Nuke + teamB.MapPool.Ancient) / 3.0
	scoreMapA := mapPowerA / (mapPowerA + mapPowerB)

	// финал общий
	pA := wRank*scoreRankA + wMap*scoreMapA + wForm*scoreFormA
	pA = clamp01(pA)
	pB := 1 - pA
	return pA, pB
}

func CalcOdds(c *gin.Context) {
	var req models.OddsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Margin <= 0 {
		req.Margin = 0.06 // обычно 0.06 гдето
	}

	collection := database.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objA, err := primitive.ObjectIDFromHex(req.TeamAID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TeamA id"})
		return
	}
	objB, err := primitive.ObjectIDFromHex(req.TeamBID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid TeamB id"})
		return
	}

	var teamA models.Team
	var teamB models.Team

	if err := collection.FindOne(ctx, bson.M{"_id": objA}).Decode(&teamA); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teamA not found"})
		return
	}
	if err := collection.FindOne(ctx, bson.M{"_id": objB}).Decode(&teamB); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "teamB not found"})
		return
	}

	pA, pB := calculateProbability(teamA, teamB)

	// добавляем маржу мани
	pA_m := pA * (1 + req.Margin)
	pB_m := pB * (1 + req.Margin)

	oddA := round2(1 / pA_m)
	oddB := round2(1 / pB_m)

	resp := models.OddsResponse{
		TeamA: teamA.TeamName,
		TeamB: teamB.TeamName,
		ProbA: round2(pA),
		ProbB: round2(pB),
		OddA:  oddA,
		OddB:  oddB,
	}

	c.JSON(http.StatusOK, resp)
}
