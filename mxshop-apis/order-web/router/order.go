package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/order-web/api/order"
	"mxshop_api/order-web/api/pay"
	"mxshop_api/order-web/middlewares"
)

func InitOrderRouter(Router *gin.RouterGroup) {
	OrderRouter := Router.Group("order").Use( middlewares.JWTAuth()).Use(middlewares.Trace())
	{
		OrderRouter.GET("",order.List) // 获取订单列表
		OrderRouter.POST("",order.New)       //新建订单
		OrderRouter.GET("/:id",order.Detail) //获取订单详情
	}
	PayRouter :=Router.Group("pay")
	{
		PayRouter.POST("alipay/notify",pay.Notify)
	}
}