package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	libClient "github.com/gunni1/leipzig-library-game-stock-api/pkg/library-le"
)

type GetAvailableGamesParams struct {
	Branch   int    `uri:"branch"`
	Platform string `uri:"platform"`
}

func RegisterRoutes(router *gin.Engine) {
	router.GET("/games/:branch/:platform", func(c *gin.Context) {
		var params GetAvailableGamesParams
		if err := c.BindUri(&params); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		platform, encErr := url.QueryUnescape(params.Platform)
		if encErr != nil {
			c.AbortWithError(http.StatusBadRequest, encErr)
			return
		}

		client := libClient.Client{}
		games := client.FindAvailabelGames(params.Branch, platform)
		c.JSON(http.StatusOK, games)
	})
}
