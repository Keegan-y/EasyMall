package initialize

import (
	"fmt"
	_ "github.com/mbobakov/grpc-consul-resolver"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"mxshop_api/user-web/global"
	"mxshop_api/user-web/proto"
)

func InitSrvConn()  {
	consulInfo := global.ServerConfig.ConsulInfo
	userConn, err :=grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s",consulInfo.Host,consulInfo.Port,global.ServerConfig.UserSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err !=nil{
		zap.S().Fatal("[InitSrvConn] 连接 [用户服务失败]")
	}
	userSrcClient := proto.NewUserClient(userConn)
	global.UserSrvClient = userSrcClient
}

