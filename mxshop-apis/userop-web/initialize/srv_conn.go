package initialize

import (
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"mxshop_api/userop-web/global"
	"mxshop_api/userop-web/proto"
)

func InitSrvConn()  {
	consulInfo := global.ServerConfig.ConsulInfo
	goodsConn, err :=grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",consulInfo.Host,consulInfo.Port,global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err !=nil{
		zap.S().Fatal("[InitSrvConn] 连接 [商品服务失败]")
	}
	global.GoodsSrvClient = proto.NewGoodsClient(goodsConn)

	userOpconn, err :=grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",consulInfo.Host,consulInfo.Port,global.ServerConfig.UserOpSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err !=nil{
		zap.S().Fatal("[InitSrvConn] 连接 [用户操作服务失败]")
	}
	global.UserFavClient = proto.NewUserFavClient(userOpconn)
	global.MessageClient = proto.NewMessageClient(userOpconn)
	global.AddressClient = proto.NewAddressClient(userOpconn)


}

