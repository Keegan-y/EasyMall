package initialize

import (
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"mxshop_api/goods-web/global"
	"mxshop_api/goods-web/proto"
	"mxshop_api/goods-web/utils/otgrpc"
)

func InitSrvConn()  {
	consulInfo := global.ServerConfig.ConsulInfo
	userConn, err :=grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",consulInfo.Host,consulInfo.Port,global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err !=nil{
		zap.S().Fatal("[InitSrvConn] 连接 [用户服务失败]")
	}
	global.GoodsSrvClient = proto.NewGoodsClient(userConn)
	invConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.InvSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 [库存服务失败]")
	}
	global.InvSrvClient = proto.NewInventoryClient(invConn)
}

