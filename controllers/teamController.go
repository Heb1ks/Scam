package controllers

import (
	"context"
	"net/http"
	"time"

	"Scam/database"
	"Scam/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateTeam(c *gin.Context) {
	var team models.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team.CreatedAt = time.Now().Unix()

	collection := database.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, team)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert team"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "team created",
		"id":      res.InsertedID,
	})
}

func GetTeams(c *gin.Context) {
	collection := database.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch teams"})
		return
	}
	defer cur.Close(ctx)

	var teams []models.Team
	if err := cur.All(ctx, &teams); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode teams"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

func GetTeamByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	collection := database.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var team models.Team
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
		return
	}

	c.JSON(http.StatusOK, team)
}

func UpdateTeam(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	var payload models.Team
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{
		"$set": bson.M{
			"teamName": payload.TeamName,
			"hltvRank": payload.HLTVRank,
			"formWins": payload.FormWins,
			"formLoss": payload.FormLoss,
			"mapPool":  payload.MapPool,
		},
	}

	collection := database.GetCollection("teams")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "team updated"})
}
