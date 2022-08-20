package initialize

import (
	"fmt"
	"mxshop_srvs/order_srv/global"
	"mxshop_srvs/order_srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func InitSrvConn() {
	//初始化商品服务
	consulInfo := global.ServerConfig.ConsulInfo
	goodsConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.GoodsSrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 [商品服务失败]")
	}
	global.GoodsSrcClient = proto.NewGoodsClient(goodsConn)
	invConn, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?wait=14s", consulInfo.Host, consulInfo.Port, global.ServerConfig.InventorySrvInfo.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	if err != nil {
		zap.S().Fatal("[InitSrvConn] 连接 [库存服务失败]")
	}
	global.InventorySrvClient = proto.NewInventoryClient(invConn)
}
