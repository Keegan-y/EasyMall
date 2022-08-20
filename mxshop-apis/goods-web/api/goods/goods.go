package goods

import (
	"context"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"

	"mxshop_api/goods-web/api"
	"mxshop_api/goods-web/forms"
	"mxshop_api/goods-web/global"
	"mxshop_api/goods-web/proto"
	"net/http"
	"strconv"
)

func List(c *gin.Context) {
	request := &proto.GoodsFilterRequest{}

	priceMin := c.DefaultQuery("pmin", "0")
	priceMinInt, _ := strconv.Atoi(priceMin)
	request.PriceMin = int32(priceMinInt)

	priceMax := c.DefaultQuery("pmax", "0")
	priceMaxInt, _ := strconv.Atoi(priceMax)
	request.PriceMax = int32(priceMaxInt)

	isHot := c.DefaultQuery("ih", "0")
	if isHot == "1" {
		request.IsHot = true
	}

	isNew := c.DefaultQuery("in", "0")
	if isNew == "1" {
		request.IsNew = true
	}

	isTab := c.DefaultQuery("it", "0")
	if isTab == "1" {
		request.IsTab = true
	}
	categoryId := c.DefaultQuery("c", "0")
	categoryIdInt, _ := strconv.Atoi(categoryId)
	request.TopCategory = int32(categoryIdInt)

	pages := c.DefaultQuery("p", "0")
	pageInt, _ := strconv.Atoi(pages)
	request.Pages = int32(pageInt)

	perNums := c.DefaultQuery("pnum", "0")
	perNumsInt, _ := strconv.Atoi(perNums)
	request.PagePerNums = int32(perNumsInt)

	keywords := c.DefaultQuery("q", "")
	request.KeyWords = keywords

	brandId := c.DefaultQuery("b", "0")
	brandIdInt, _ := strconv.Atoi(brandId)
	request.Brand = int32(brandIdInt)

	//配置熔断机制
	e, b := sentinel.Entry("goods-list", sentinel.WithTrafficType(base.Inbound))
	if b != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"msg":"请求过于频繁，请稍后重试",
		})
		return
	}
	//请求商品的service服务
	rsp, err := global.GoodsSrvClient.GoodsList(context.WithValue(context.Background(), "ginContext", c), request)
	if err != nil {
		zap.S().Errorw("[List] 查询 [商品列表] 失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	e.Exit()

	reMap := map[string]interface{}{
		"total": rsp.Total,
	}

	goodsList := make([]interface{}, 0)
	for _, value := range rsp.Data {
		goodsList = append(goodsList, map[string]interface{}{
			"id":          value.Id,
			"name":        value.Name,
			"goods_brief": value.GoodsBrief,
			"desc":        value.GoodsDesc,
			"ship_free":   value.ShipFree,
			"images":      value.Images,
			"desc_images": value.DescImages,
			"front_image": value.GoodsFrontImage,
			"shop_price":  value.ShopPrice,
			"category": map[string]interface{}{
				"id":   value.Category.Id,
				"name": value.Category.Name,
			},
			"brand": map[string]interface{}{
				"id":   value.Brand.Id,
				"name": value.Brand.Name,
				"logo": value.Brand.Logo,
			},
			"is_hot":  value.IsHot,
			"is_new":  value.IsNew,
			"on_sale": value.OnSale,
		})
	}
	reMap["data"] = goodsList
	c.JSON(http.StatusOK, reMap)
}
func New(c *gin.Context) {
	goodsForm := forms.GoodsForm{}
	if err := c.ShouldBindJSON(&goodsForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}
	goodsClient := global.GoodsSrvClient
	rsp, err := goodsClient.CreateGoods(context.WithValue(context.Background(), "ginContext", c), &proto.CreateGoodsInfo{
		Name:            goodsForm.Name,
		GoodsSn:         goodsForm.GoodsSn,
		Stocks:          goodsForm.Stocks,
		MarketPrice:     goodsForm.MarketPrice,
		ShopPrice:       goodsForm.ShopPrice,
		GoodsBrief:      goodsForm.GoodsBrief,
		ShipFree:        *goodsForm.ShipFree,
		Images:          goodsForm.Images,
		DescImages:      goodsForm.Images,
		GoodsFrontImage: goodsForm.FrontImage,
		CategoryId:      goodsForm.CategoryId,
		BrandId:         goodsForm.Brand,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	//设置库存
	InventoryClient :=global.InvSrvClient
	_,err =InventoryClient.SetInv(context.WithValue(context.Background(), "ginContext", c), &proto.GoodsInvInfo{
		Num:goodsForm.Nums,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.JSON(http.StatusOK, rsp)
}
func Detail(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	r, err := global.GoodsSrvClient.GetGoodsDetail(context.WithValue(context.Background(), "ginContext", c), &proto.GoodInfoRequest{
		Id: int32(i)},
	)
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	//去库存服务查询库存
	r2,err :=global.InvSrvClient.InvDetail(context.WithValue(context.Background(), "ginContext", c),&proto.GoodsInvInfo{
		GoodsId: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	rsp := map[string]interface{}{
		"id":          r.Id,
		"name":        r.Name,
		"goods_brief": r.GoodsBrief,
		"desc":        r.GoodsDesc,
		"ship_free":   r.ShipFree,
		"images":      r.Images,
		"desc_images": r.DescImages,
		"front_image": r.GoodsFrontImage,
		"shop_price":  r.ShopPrice,
		"category": map[string]interface{}{
			"id":   r.Category.Id,
			"name": r.Category.Name,
		},
		"brand": map[string]interface{}{
			"id":   r.Brand.Id,
			"name": r.Brand.Name,
			"logo": r.Brand.Logo,
		},
		"is_hot":  r.IsHot,
		"is_new":  r.IsNew,
		"on_sale": r.OnSale,
		"stocks":r2.Num,
	}
	c.JSON(http.StatusOK, rsp)
}
func Delete(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteGoods(context.WithValue(context.Background(), "ginContext", c), &proto.DeleteGoodsInfo{
		Id: int32(i)},
	)
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.Status(http.StatusOK)
	return
}
func Stocks(c *gin.Context) {
	id := c.Param("id")
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	//TODO 商品的库存
	return
}
func UpdateStatus(c *gin.Context) {
	goodsStatusForm := forms.GoodsStatusForm{}
	if err := c.ShouldBindJSON(&goodsStatusForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if _, err = global.GoodsSrvClient.UpdateGoods(context.WithValue(context.Background(), "ginContext", c), &proto.CreateGoodsInfo{
		Id:     int32(i),
		IsHot:  *goodsStatusForm.IsHot,
		IsNew:  *goodsStatusForm.IsNew,
		OnSale: *goodsStatusForm.OnSale,
	}); err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "修改成功",
	})
}
func Update(c *gin.Context) {
	goodsForm := forms.GoodsForm{}
	if err := c.ShouldBindJSON(&goodsForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if _, err = global.GoodsSrvClient.UpdateGoods(context.WithValue(context.Background(), "ginContext", c), &proto.CreateGoodsInfo{
		Id:              int32(i),
		Name:            goodsForm.Name,
		GoodsSn:         goodsForm.GoodsSn,
		Stocks:          goodsForm.Stocks,
		MarketPrice:     goodsForm.MarketPrice,
		ShopPrice:       goodsForm.ShopPrice,
		GoodsBrief:      goodsForm.GoodsBrief,
		ShipFree:        *goodsForm.ShipFree,
		Images:          goodsForm.Images,
		DescImages:      goodsForm.DescImages,
		GoodsFrontImage: goodsForm.FrontImage,
		CategoryId:      goodsForm.CategoryId,
		BrandId:         goodsForm.Brand,
	}); err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": "更新成功",
	})
}
