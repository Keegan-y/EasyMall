package router

import (
	"github.com/gin-gonic/gin"
	"mxshop_api/goods-web/api/goods"
	"mxshop_api/goods-web/middlewares"
)

// InitGoodsRouter goodsRouter
func InitGoodsRouter(Router *gin.RouterGroup) {
	GoodsRouter := Router.Group("goods").Use(middlewares.Trace())
	{
		GoodsRouter.GET("",goods.List)//获取商品列表
		GoodsRouter.POST("",middlewares.JWTAuth(),middlewares.IsAdminAuth(),goods.New) //新增商品 该接口需要管理员权限
		GoodsRouter.GET("/:id",goods.Detail)//获取商品详情
		GoodsRouter.DELETE("/:id",middlewares.JWTAuth(),middlewares.IsAdminAuth(),goods.Delete)//删除商品
		GoodsRouter.GET("/:id/stocks",goods.Stocks)//获取库存

		GoodsRouter.PUT("/:id",middlewares.JWTAuth(),middlewares.IsAdminAuth(),goods.Update)//更新商品
		GoodsRouter.PATCH("/:id",middlewares.JWTAuth(),middlewares.IsAdminAuth(),goods.UpdateStatus)//更新商品状态
	}
}
