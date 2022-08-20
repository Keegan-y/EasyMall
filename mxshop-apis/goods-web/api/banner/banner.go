package banners

import (
	"context"
	"mxshop_api/goods-web/api"
	"mxshop_api/goods-web/forms"
	"mxshop_api/goods-web/global"
	"mxshop_api/goods-web/proto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/empty"


)


func List(c *gin.Context) {
	rsp, err := global.GoodsSrvClient.BannerList(context.Background(), &empty.Empty{})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["index"] = value.Index
		reMap["image"] = value.Image
		reMap["url"] = value.Url

		result = append(result, reMap)
	}

	c.JSON(http.StatusOK, result)
}

func New(c *gin.Context) {
	bannerForm := forms.BannerForm{}
	if err := c.ShouldBindJSON(&bannerForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	rsp, err := global.GoodsSrvClient.CreateBanner(context.Background(), &proto.BannerRequest{
		Index:      int32(bannerForm.Index),
		Url: bannerForm.Url,
		Image: bannerForm.Image,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	response := make(map[string]interface{})
	response["id"] = rsp.Id
	response["index"] = rsp.Index
	response["url"] = rsp.Url
	response["image"] = rsp.Image

	c.JSON(http.StatusOK, response)
}

func Update(c *gin.Context) {
	bannerForm := forms.BannerForm{}
	if err := c.ShouldBindJSON(&bannerForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateBanner(context.Background(), &proto.BannerRequest{
		Id:         int32(i),
		Index:      int32(bannerForm.Index),
		Url:       bannerForm.Url,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.Status(http.StatusOK)
}


func Delete(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteBanner(context.Background(), &proto.BannerRequest{Id: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.JSON(http.StatusOK, "")
}
