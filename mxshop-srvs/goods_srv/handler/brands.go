package handler

import (
	"context"
	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"

	"github.com/golang/protobuf/ptypes/empty"

	"google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"
)

//品牌和轮播图
func (s *GoodsServer) BrandList(c context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	brandListResponse := proto.BrandListResponse{}
	var brands []model.Brands

	result := global.DB.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}
	var total int64
	global.DB.Model(&model.Brands{}).Count(&total)
	brandListResponse.Total = int32(total)
	var brandResponses []*proto.BrandInfoResponse
	for _, brand := range brands {
		brandResponses = append(brandResponses, &proto.BrandInfoResponse{
			Id:   brand.ID,
			Name: brand.Name,
			Logo: brand.Logo,
		})
	}
	brandListResponse.Data = brandResponses
	return &brandListResponse, nil
}

func (s *GoodsServer) CreateBrand(c context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	if result := global.DB.First(&model.Brands{}); result.RowsAffected == 1 {
		return nil, status.Errorf(codes.InvalidArgument, "品牌已存在")
	}
	brand := model.Brands{}
	brand.Name = req.Name
	brand.Logo = req.Logo
	global.DB.Save(brand)
	return &proto.BrandInfoResponse{Id: brand.ID}, nil
}

func (s *GoodsServer) DeleteBrand(c context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Brands{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	return &empty.Empty{}, nil
}

func (s *GoodsServer) UpdateBrand(c context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	var brand model.Brands
	result := global.DB.First(&brand, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "品牌不存在")
	}
	if req.Name != "" {
		brand.Name = req.Name
	}
	if req.Logo != "" {
		brand.Logo = req.Logo
	}
	global.DB.Save(&brand)
	return &empty.Empty{}, nil
}
