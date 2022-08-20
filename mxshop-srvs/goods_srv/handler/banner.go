package handler

import (
	"context"
	"mxshop_srvs/goods_srv/global"
	"mxshop_srvs/goods_srv/model"
	"mxshop_srvs/goods_srv/proto"

	"github.com/golang/protobuf/ptypes/empty"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"google.golang.org/protobuf/types/known/emptypb"
)

//轮播图相关接口
func (s *GoodsServer) BannerList(c context.Context, req *emptypb.Empty) (*proto.BannerListResponse, error) {
	bannerListResponse := proto.BannerListResponse{}
	var banners []model.Banner

	result := global.DB.Find(&banners)
	bannerListResponse.Total = int32(result.RowsAffected)
	var bannerResponse []*proto.BannerResponse
	for _, banner := range banners {
		bannerResponse = append(bannerResponse, &proto.BannerResponse{
			Id:    banner.ID,
			Index: banner.Index,
			Image: banner.Image,
			Url:   banner.Url,
		})
	}
	bannerListResponse.Data = bannerResponse
	return &bannerListResponse, nil
}

func (s *GoodsServer) CreateBanner(c context.Context, req *proto.BannerRequest) (*proto.BannerResponse, error) {
	banner := model.Banner{}
	banner.Image = req.Image
	banner.Index = req.Index
	banner.Url = req.Url
	global.DB.Save(&banner)
	return &proto.BannerResponse{Id: int32(banner.ID)}, nil
}

func (s *GoodsServer) DeleteBanner(c context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	if result := global.DB.Delete(&model.Banner{}, req.Id); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	return &emptypb.Empty{}, nil
}

func (s *GoodsServer) UpdateBanner(c context.Context, req *proto.BannerRequest) (*emptypb.Empty, error) {
	banner := model.Banner{}
	result := global.DB.First(&banner, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "轮播图不存在")
	}
	if req.Url != "" {
		banner.Url = req.Url
	}
	if req.Image != "" {
		banner.Image = req.Image
	}
	if req.Index != 0 {
		banner.Index = req.Index
	}
	global.DB.Save(&banner)
	return &empty.Empty{}, nil
}
