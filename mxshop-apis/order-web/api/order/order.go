package order

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
	"mxshop_api/order-web/api"
	"mxshop_api/order-web/forms"
	"mxshop_api/order-web/global"
	"mxshop_api/order-web/models"
	"mxshop_api/order-web/proto"
	"net/http"
	"strconv"
)

func List(c *gin.Context) {
	userId, _ := c.Get("userId")
	claims, _ := c.Get("claims")
	request := proto.OrderFilterRequest{}
	//如果是管理员用户则返回所有的订单
	model := claims.(*models.CustomClaims)
	//1是普通用户,2是管理员
	if model.AuthorityId == 1 {
		request.UserId = int32(userId.(uint))
	}
	pages := c.DefaultQuery("p", "0")
	pageInt, _ := strconv.Atoi(pages)
	request.Pages = int32(pageInt)

	perNums := c.DefaultQuery("pnum", "0")
	perNumsInt, _ := strconv.Atoi(perNums)
	request.PagePerNums = int32(perNumsInt)

	rsp, err := global.OrderSrvClient.OrderList(context.WithValue(context.Background(), "ginContext", c), &request)
	if err != nil {
		zap.S().Errorw("获取订单列表失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	reMap := gin.H{
		"total": rsp.Total,
	}
	orderList := make([]interface{}, 0)
	for _, item := range rsp.Data {
		tmpMap := map[string]interface{}{}
		tmpMap["id"] = item.Id
		tmpMap["status"] = item.Status
		tmpMap["pay_type"] = item.PayType
		tmpMap["user"] = item.UserId
		tmpMap["post"] = item.Post
		tmpMap["address"] = item.Address
		tmpMap["name"] = item.Name
		tmpMap["mobile"] = item.Mobile
		tmpMap["order_sn"] = item.OrderSn
		tmpMap["id"] = item.Id
		tmpMap["add_time"] = item.AddTime
		orderList = append(orderList, tmpMap)
	}
	reMap["data"] = orderList
	c.JSON(http.StatusOK, reMap)
}
func New(c *gin.Context) {
	orderForm := forms.CreateOrderForm{}
	if err := c.ShouldBindJSON(&orderForm); err != nil {
		api.HandleValidatorError(c, err)
	}
	userId, _ := c.Get("userId")
	rsp, err := global.OrderSrvClient.CreateOrder(context.WithValue(context.Background(), "ginContext", c), &proto.OrderRequest{
		UserId:  int32(userId.(uint)),
		Address: orderForm.Address,
		Name:    orderForm.Name,
		Mobile:  orderForm.Mobile,
		Post:    orderForm.Post,
	})
	if err != nil {
		zap.S().Errorw("新建订单失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	//生成支付宝的url
	client, err := alipay.New(global.ServerConfig.AlipayInfo.AppID, global.ServerConfig.AlipayInfo.PrivateKey, false)
	if err !=nil{
		zap.S().Errorw("生成支付宝url失败")
		c.JSON(http.StatusInternalServerError,gin.H{
			"msg":err.Error(),
		})
		return
	}
	err =client.LoadAliPayPublicKey(global.ServerConfig.AlipayInfo.AliPublicKey)
	if err !=nil{
		zap.S().Errorw("加载支付宝公钥失败")
		c.JSON(http.StatusInternalServerError,gin.H{
			"msg":err.Error(),
		})
		return
	}
	var p = alipay.TradePagePay{}
	p.NotifyURL = global.ServerConfig.AlipayInfo.NotifyUrl
	p.ReturnURL = global.ServerConfig.AlipayInfo.ReturnUrl
	p.Subject = "wam"+rsp.OrderSn
	p.OutTradeNo = rsp.OrderSn
	p.TotalAmount = strconv.FormatFloat(float64(rsp.Total),'f',2,64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url ,err :=client.TradePagePay(p)
	if err !=nil{
		zap.S().Errorw("生成支付url失败")
		c.JSON(http.StatusInternalServerError,gin.H{
			"msg":err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id": rsp.Id,
		"alipay":url.String(),
	})
}
func Detail(c *gin.Context) {
	id := c.Param("id")
	userId, _ := c.Get("userId")
	i, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg": "url格式出错",
		})
		return
	}

	//如果是管理员用户则返回所有的订单
	claims, _ := c.Get("claims")
	request := proto.OrderRequest{
		Id: int32(i),
	}
	model := claims.(*models.CustomClaims)
	if model.AuthorityId == 1 {
		request.UserId = int32(userId.(uint))
	}
	rsp, err := global.OrderSrvClient.OrderDetail(context.WithValue(context.Background(), "ginContext", c), &request)
	if err != nil {
		zap.S().Errorw("获取订单详情失败")
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	reMap := gin.H{}
	reMap["id"] = rsp.OrderInfo.Id
	reMap["status"] = rsp.OrderInfo.Status
	reMap["user"] = rsp.OrderInfo.UserId
	reMap["post"] = rsp.OrderInfo.Post
	reMap["total"] = rsp.OrderInfo.Total
	reMap["address"] = rsp.OrderInfo.Address
	reMap["name"] = rsp.OrderInfo.Name
	reMap["mobile"] = rsp.OrderInfo.Mobile
	reMap["pay_type"] = rsp.OrderInfo.PayType
	reMap["order_sb"] = rsp.OrderInfo.OrderSn

	goodList := make([]interface{}, 0)
	for _, item := range rsp.Goods {
		tmpMpa := gin.H{
			"id":    item.GoodsId,
			"name":  item.GoodsName,
			"image": item.GoodsImage,
			"price": item.GoodsPrice,
			"nums":  item.Nums,
		}
		goodList = append(goodList, tmpMpa)
	}
	client, err := alipay.New(global.ServerConfig.AlipayInfo.AppID, global.ServerConfig.AlipayInfo.PrivateKey, false)
	if err !=nil{
		zap.S().Errorw("生成支付宝url失败")
		c.JSON(http.StatusInternalServerError,gin.H{
			"msg":err.Error(),
		})
		return
	}
	err =client.LoadAliPayPublicKey(global.ServerConfig.AlipayInfo.AliPublicKey)
	if err !=nil{
		zap.S().Errorw("加载支付宝公钥失败")
		c.JSON(http.StatusInternalServerError,gin.H{
			"msg":err.Error(),
		})
		return
	}
	var p = alipay.TradePagePay{}
	p.NotifyURL = global.ServerConfig.AlipayInfo.NotifyUrl
	p.ReturnURL = global.ServerConfig.AlipayInfo.ReturnUrl
	p.Subject = "wam"+rsp.OrderInfo.OrderSn
	p.OutTradeNo = rsp.OrderInfo.OrderSn
	p.TotalAmount = strconv.FormatFloat(float64(rsp.OrderInfo.Total),'f',2,64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	url ,err :=client.TradePagePay(p)
	if err !=nil{
		zap.S().Errorw("生成支付url失败")
		c.JSON(http.StatusInternalServerError,gin.H{
			"msg":err.Error(),
		})
		return
	}
	reMap["goods"] = goodList
	reMap["alipay_urt"]=url.String()
	c.JSON(http.StatusOK, reMap)
}
