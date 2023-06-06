package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gunni1/leipzig-library-game-stock-api/api"
)

func main() {
	router := gin.Default()
	api.RegisterRoutes(router)
	router.Run(":8080")
}
