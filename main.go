package main

import (
	"Scam/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(gin.Logger())

	routes.UserRoutes(r)

	r.Run(":8080")
}
