package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/order-web/api/shop_cart"
	"mxshop_api/order-web/middlewares"
)

func InitShopCartRouter(Router *gin.RouterGroup) {
	OrderRouter := Router.Group("shopcarts").Use(middlewares.JWTAuth())
	{
		OrderRouter.GET("", shop_cart.List) // 购物车列表
		OrderRouter.DELETE("/:id",  shop_cart.Delete)       //删除条目
		OrderRouter.POST("", shop_cart.New) //添加商品到购物车
		OrderRouter.PATCH("/:id",shop_cart.Update) //修改条目
	}
}