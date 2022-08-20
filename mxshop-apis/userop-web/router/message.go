package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/userop-web/api/message"
	"mxshop_api/userop-web/middlewares"
)

func InitMessageRouter(Router *gin.RouterGroup) {
	MessageRouter := Router.Group("message").Use(middlewares.JWTAuth())
	{
		MessageRouter.GET("", message.List)          // 获取留言信息
		MessageRouter.POST("", message.New)       //新建留言信息
	}
}