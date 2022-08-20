package main

import (
	"flag"
	"fmt"
	"mxshop_srvs/inventory_srv/global"
	"mxshop_srvs/inventory_srv/handler"
	"mxshop_srvs/inventory_srv/initialize"
	"mxshop_srvs/inventory_srv/proto"
	"mxshop_srvs/inventory_srv/utils"
	"mxshop_srvs/inventory_srv/utils/consul"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/health"

	"google.golang.org/grpc/health/grpc_health_v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	IP := flag.String("ip", "192.168.124.10", "ip地址")
	Port := flag.Int("port", 50053, "端口号")

	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	initialize.InitRedis()
	flag.Parse()
	zap.S().Info("ip: ", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port: ", *Port)
	server := grpc.NewServer()
	proto.RegisterInventoryServer(server, &handler.InventoryServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *IP, *Port))
	if err != nil {
		panic("failed to listen:" + err.Error())
	}
	//注册健康检查
	grpc_health_v1.RegisterHealthServer(server, health.NewServer())
	//启动服务
	go func() {
		err = server.Serve(lis)
		if err != nil {
			panic("failed to start grpc:" + err.Error())
		}
	}()
	//服务注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err = register_client.Register(global.ServerConfig.Host, *Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败", err.Error())
	}
	zap.S().Debugf("启动服务器，端口:%d", *Port)
	//监听库存归还Topic
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"192.168.124.4:9876"}),
		consumer.WithGroupName("wam-inventory"),
	)

	if err := c.Subscribe("order_reback", consumer.MessageSelector{}, handler.AutoReback); err != nil {
		fmt.Println("读取消息失败")
	}
	_ = c.Start()
	//不能让主goroutine退出
	//接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = c.Shutdown()
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Panic("注销失败:", err.Error())
	} else {
		zap.S().Infof("注销成功")
	}
}
