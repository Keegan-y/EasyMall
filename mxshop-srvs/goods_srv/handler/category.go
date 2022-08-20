package handler

import (
	"context"
	"encoding/json"
	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"

	"github.com/golang/protobuf/ptypes/empty"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAllCategorysList 获取所有商品分类
func (s *GoodsServer) GetAllCategorysList(c context.Context, req *emptypb.Empty) (*proto.CategoryListResponse, error) {
	var categorys []model.Category
	global.DB.Where(&model.Category{Level: 1}).Preload("SubCategory.SubCategory").Find(&categorys)
	b, _ := json.Marshal(&categorys)
	return &proto.CategoryListResponse{JsonData: string(b)}, nil
}

// GetSubCategory 获取子分类
func (s *GoodsServer) GetSubCategory(c context.Context, req *proto.CategoryListRequest) (*proto.SubCategoryListResponse, error) {
	var categoryListResponse = proto.SubCategoryListResponse{}
	var category model.Category
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	categoryListResponse.Info = &proto.CategoryInfoResponse{
		Id:             category.ID,
		Name:           category.Name,
		ParentCategory: category.ParentCategoryID,
		Level:          category.Level,
		IsTab:          category.IsTab,
	}
	var subCategorys []model.Category
	var subCategoryResponse []*proto.CategoryInfoResponse
	Preloads := "SubCategory"
	if category.Level == 1 {
		Preloads = "SubCategory.SubCategory"
	}
	global.DB.Where(&model.Category{ParentCategoryID: req.Id}).Preload(Preloads).Find(&subCategorys)
	for _, subCategory := range subCategorys {
		subCategoryResponse = append(subCategoryResponse, &proto.CategoryInfoResponse{
			Id:             subCategory.ID,
			Name:           subCategory.Name,
			Level:          subCategory.Level,
			IsTab:          subCategory.IsTab,
			ParentCategory: subCategory.ParentCategoryID,
		})
	}
	categoryListResponse.SubCategorys = subCategoryResponse
	return &categoryListResponse, nil
}

// CreateCategory 新增分类
func (s *GoodsServer) CreateCategory(c context.Context, req *proto.CategoryInfoRequest) (*proto.CategoryInfoResponse, error) {
	category := model.Category{}
	category.Name = req.Name
	category.Level = req.Level
	if req.Level != 1 {
		category.ParentCategoryID = req.ParentCategory
	}
	category.IsTab = req.IsTab
	global.DB.Save(&category)
	return &proto.CategoryInfoResponse{Id: int32(category.ID)}, nil
}

// DeleteCategory 删除商品分类
func (s *GoodsServer) DeleteCategory(c context.Context, req *proto.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Category{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	return &empty.Empty{}, nil
}

// UpdateCategory 更新商品分类
func (s *GoodsServer) UpdateCategory(c context.Context, req *proto.CategoryInfoRequest) (*emptypb.Empty, error) {
	category := model.Category{}
	if result := global.DB.First(&category, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "商品分类不存在")
	}
	if req.Name != "" {
		category.Name = req.Name
	}
	if req.ParentCategory == 0 {
		category.ParentCategoryID = req.ParentCategory
	}
	if req.Level != 0 {
		category.Level = req.Level
	}
	if req.IsTab {
		category.IsTab = req.IsTab
	}
	global.DB.Save(&category)
	return &empty.Empty{}, nil
}
