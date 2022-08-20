package main

import (
	"flag"
	"fmt"
	"mxshop_srvs/userop_srv/global"
	"mxshop_srvs/userop_srv/handler"
	"mxshop_srvs/userop_srv/initialize"
	"mxshop_srvs/userop_srv/proto"
	"mxshop_srvs/userop_srv/utils"
	"mxshop_srvs/userop_srv/utils/consul"
	"net"
	"os"
	"os/signal"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/health"

	"google.golang.org/grpc/health/grpc_health_v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	IP := flag.String("ip", "192.168.124.10", "ip地址")
	Port := flag.Int("port", 50055, "端口号")

	//初始化
	initialize.InitLogger()
	initialize.InitConfig()
	initialize.InitDB()
	flag.Parse()
	zap.S().Info("ip: ", *IP)
	if *Port == 0 {
		*Port, _ = utils.GetFreePort()
	}
	zap.S().Info("port: ", *Port)
	server := grpc.NewServer()
	proto.RegisterAddressServer(server, &handler.UserOpServer{})
	proto.RegisterMessageServer(server, &handler.UserOpServer{})
	proto.RegisterUserFavServer(server, &handler.UserOpServer{})

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
	//接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Panic("注销失败:", err.Error())
	} else {
		zap.S().Infof("注销成功")
	}
}
