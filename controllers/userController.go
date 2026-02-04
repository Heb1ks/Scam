package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetProfile(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"user": "Student", "balance": 1000.0})
}

func PlaceBet(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"status": "Ставка принята!"})
}
