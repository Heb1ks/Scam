package routes

import (
	"Scam/controllers"

	"github.com/gin-gonic/gin"
)

func TeamRoutes(router *gin.RouterGroup) {
	router.POST("/teams", controllers.CreateTeam)
	router.GET("/teams", controllers.GetTeams)
	router.GET("/teams/:id", controllers.GetTeamByID)
	router.PUT("/teams/:id", controllers.UpdateTeam)
}
