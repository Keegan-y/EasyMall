package initialize

import (
	"mxshop_api/userop-web/middlewares"
	"mxshop_api/userop-web/router"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.Default()
	Router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"success": true,
		})
	})
	//配置跨域
	Router.Use(middlewares.Cors())
	ApiGroup := Router.Group("/v1")
	router.InitMessageRouter(ApiGroup)
	router.InitAddressRouter(ApiGroup)
	router.InitUserFavRouter(ApiGroup)
	return Router
}
