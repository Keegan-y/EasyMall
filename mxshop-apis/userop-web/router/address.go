package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/userop-web/api/address"
	"mxshop_api/userop-web/middlewares"
)

func InitAddressRouter(Router *gin.RouterGroup) {
	AddressRouter := Router.Group("address").Use(middlewares.JWTAuth())
	{
		AddressRouter.GET("",address.List)          //获取收货地址
		AddressRouter.DELETE("/:id",address.Delete) //删除收货地址
		AddressRouter.POST("",address.New)       //新建收货地址
		AddressRouter.PUT("/:id",address.Update) //更新收货地址
	}
}