package routes

import (
	"Scam/controllers"
	"github.com/gin-gonic/gin"
)

func UserRoutes(r *gin.Engine) {
	userGroup := r.Group("/user")
	{
		userGroup.GET("/profile", controllers.GetProfile)
		userGroup.POST("/bet", controllers.PlaceBet)

	}
}
