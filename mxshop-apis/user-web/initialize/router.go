package initialize

import
(
	"mxshop_api/user-web/middlewares"
	"mxshop_api/user-web/router"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Routers() *gin.Engine {
	Router := gin.Default()
	Router.GET("/health", func(c *gin.Context){
		c.JSON(http.StatusOK, gin.H{
			"code":http.StatusOK,
			"success":true,
		})
	})
	//配置跨域
	Router.Use(middlewares.Cors())
	ApiGroup := Router.Group("/v1")
	router.InitUserRouter(ApiGroup)
	router.InitBaseRouter(ApiGroup)
	return Router
}
