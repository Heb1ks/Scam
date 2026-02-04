package routes

import (
	"Scam/controllers"

	"github.com/gin-gonic/gin"
)

func OddsRoutes(router *gin.RouterGroup) {
	router.POST("/odds/calc", controllers.CalcOdds)
}
