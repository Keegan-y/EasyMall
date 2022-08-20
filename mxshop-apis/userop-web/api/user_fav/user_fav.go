package user_fav

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mxshop_api/userop-web/api"
	"mxshop_api/userop-web/forms"
	"mxshop_api/userop-web/global"
	"mxshop_api/userop-web/proto"

	"net/http"
	"strconv"
)

func List(c *gin.Context) {
	userId, _ := c.Get("userId")
	userFavRsp, err := global.UserFavClient.GetFavList(context.Background(), &proto.UserFavRequest{
		UserId:  int32(userId.(uint)),
	})
	if err != nil {
		zap.S().Errorw("获取收藏列表失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	ids := make([]int32, 0)
	for _, item := range userFavRsp.Data{
		ids = append(ids, item.GoodsId)
	}

	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"total":0,
		})
		return
	}

	//请求商品服务
	goods, err := global.GoodsSrvClient.BatchGetGoods(context.Background(), &proto.BatchGoodsIdInfo{
		Id: ids,
	})
	if err != nil {
		zap.S().Errorw("[List] 批量查询【商品列表】失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	reMap := map[string]interface{}{
		"total": userFavRsp.Total,
	}

	goodsList := make([]interface{}, 0)
	for _, item := range userFavRsp.Data{
		data := gin.H{
			"id":item.GoodsId,
		}

		for _, good := range goods.Data {
			if item.GoodsId == good.Id {
				data["name"] = good.Name
				data["shop_price"] = good.ShopPrice
			}
		}

		goodsList = append(goodsList, data)
	}
	reMap["data"] = goodsList
	c.JSON(http.StatusOK, reMap)
}

func New(c *gin.Context) {
	userFavForm := forms.UserFavForm{}
	if err := c.ShouldBindJSON(&userFavForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	//去商品服务查询商品是否存在
	_, err := global.GoodsSrvClient.GetGoodsDetail(context.Background(), &proto.GoodInfoRequest{
		Id: userFavForm.GoodsId,
	})
	if err != nil {
		zap.S().Errorw("查询 [商品信息]失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	userId, _ := c.Get("userId")
	_, err = global.UserFavClient.AddUserFav(context.Background(), &proto.UserFavRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: userFavForm.GoodsId,
	})

	if err != nil {
		zap.S().Errorw("添加收藏记录失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func Delete(c *gin.Context) {
	//传入的ID为商品的ID
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	userId, _ := c.Get("userId")
	_, err = global.UserFavClient.DeleteUserFav(context.Background(), &proto.UserFavRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: int32(i),
	})
	if err != nil {
		zap.S().Errorw("删除收藏记录失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":"删除成功",
	})
}

func Detail(c *gin.Context) {
	goodsId := c.Param("id")
	goodsIdInt, err := strconv.ParseInt(goodsId, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	userId, _ := c.Get("userId")
	_, err = global.UserFavClient.GetUserFavDetail(context.Background(), &proto.UserFavRequest{
		UserId:  int32(userId.(uint)),
		GoodsId: int32(goodsIdInt),
	})
	if err != nil {
		zap.S().Errorw("查询收藏状态失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.Status(http.StatusOK)
}