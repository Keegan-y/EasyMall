package brands

import (
	"context"
	"mxshop_api/goods-web/api"
	"mxshop_api/goods-web/forms"
	"mxshop_api/goods-web/global"
	"mxshop_api/goods-web/proto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"


)


func BrandList(c *gin.Context) {
	pn := c.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	pSize := c.DefaultQuery("psize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)

	rsp, err := global.GoodsSrvClient.BrandList(context.Background(), &proto.BrandFilterRequest{
		Pages: int32(pnInt),
		PagePerNums: int32(pSizeInt),
	})

	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	result := make([]interface{}, 0)
	reMap := make(map[string]interface{})
	reMap["total"] = rsp.Total
	for _, value := range rsp.Data[pnInt:pnInt*pSizeInt+pSizeInt] {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["name"] = value.Name
		reMap["logo"] = value.Logo

		result = append(result, reMap)
	}

	reMap["data"] = result

	c.JSON(http.StatusOK, reMap)
}

func NewBrand(c *gin.Context) {
	brandForm := forms.BrandForm{}
	if err := c.ShouldBindJSON(&brandForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	rsp, err := global.GoodsSrvClient.CreateBrand(context.Background(), &proto.BrandRequest{
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	request := make(map[string]interface{})
	request["id"] = rsp.Id
	request["name"] = rsp.Name
	request["logo"] = rsp.Logo

	c.JSON(http.StatusOK, request)
}

func DeleteBrand(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteBrand(context.Background(), &proto.BrandRequest{Id: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.Status(http.StatusOK)
}

func UpdateBrand(c *gin.Context) {
	brandForm := forms.BrandForm{}
	if err := c.ShouldBindJSON(&brandForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateBrand(context.Background(), &proto.BrandRequest{
		Id:   int32(i),
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.Status(http.StatusOK)
}


func GetCategoryBrandList(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	rsp, err := global.GoodsSrvClient.GetCategoryBrandList(context.Background(), &proto.CategoryInfoRequest{
		Id:int32(i),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["name"] = value.Name
		reMap["logo"] = value.Logo

		result = append(result, reMap)
	}

	c.JSON(http.StatusOK, result)
}

func CategoryBrandList(c *gin.Context) {
	//所有的list返回的数据结构
	/*
		{
			"total": 100,
			"data":[{},{}]
		}
	*/
	rsp, err := global.GoodsSrvClient.CategoryBrandList(context.Background(), &proto.CategoryBrandFilterRequest{})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	reMap := map[string]interface{}{
		"total": rsp.Total,
	}

	result := make([]interface{}, 0)
	for _, value := range rsp.Data {
		reMap := make(map[string]interface{})
		reMap["id"] = value.Id
		reMap["category"] = map[string]interface{}{
			"id": value.Category.Id,
			"name": value.Category.Name,
		}
		reMap["brand"] = map[string]interface{}{
			"id": value.Brand.Id,
			"name": value.Brand.Name,
			"logo": value.Brand.Logo,
		}

		result = append(result, reMap)
	}

	reMap["data"] = result
	c.JSON(http.StatusOK, reMap)
}

func NewCategoryBrand(c *gin.Context) {
	categoryBrandForm := forms.CategoryBrandForm{}
	if err := c.ShouldBindJSON(&categoryBrandForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	rsp, err := global.GoodsSrvClient.CreateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		CategoryId: int32(categoryBrandForm.CategoryId),
		BrandId:       int32(categoryBrandForm.BrandId),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	response := make(map[string]interface{})
	response["id"] = rsp.Id

	c.JSON(http.StatusOK, response)
}

func UpdateCategoryBrand(c *gin.Context) {
	categoryBrandForm := forms.CategoryBrandForm{}
	if err := c.ShouldBindJSON(&categoryBrandForm); err != nil {
		api.HandleValidatorError(c, err)
		return
	}

	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	_, err = global.GoodsSrvClient.UpdateCategoryBrand(context.Background(), &proto.CategoryBrandRequest{
		Id:         int32(i),
		CategoryId:      int32(categoryBrandForm.CategoryId),
		BrandId:       int32(categoryBrandForm.BrandId),
	})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}
	c.Status(http.StatusOK)
}


func DeleteCategoryBrand(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodsSrvClient.DeleteCategoryBrand(context.Background(), &proto.CategoryBrandRequest{Id: int32(i)})
	if err != nil {
		api.HandleGrpcErrorToHttp(err, c)
		return
	}

	c.JSON(http.StatusOK, "")
}