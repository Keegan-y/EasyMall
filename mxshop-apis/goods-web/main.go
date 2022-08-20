package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
	"mxshop_api/goods-web/global"
	"mxshop_api/goods-web/initialize"
	"mxshop_api/goods-web/utils"
	"mxshop_api/goods-web/utils/consul"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//0.初始化logger
	initialize.InitLogger()
	//初始化配置文件
	initialize.InitConfig()
	//初始化Routers
	Router := initialize.Routers()
	//初始化翻译
	if err := initialize.InitTrans("zh"); err != nil {
		panic(err)
	}
	//初始化srv的连接
	initialize.InitSrvConn()
	//初始化Sentinel
	initialize.InitSentinel()
	//yaml文件中配置debug,true 为开发环境 false代表生产环境
	debug := global.ServerConfig.DebugInfo.Debug
	if debug == false {
		port, err := utils.GetFreePort()
		if err == nil {
			global.ServerConfig.Port = port
		}
	}
	//服务注册
	register_client := consul.NewRegistryClient(global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	serviceId := fmt.Sprintf("%s", uuid.NewV4())
	err := register_client.Register(global.ServerConfig.Host, global.ServerConfig.Port, global.ServerConfig.Name, global.ServerConfig.Tags, serviceId)
	if err != nil {
		zap.S().Panic("服务注册失败", err.Error())
	}
	zap.S().Debugf("启动服务器，端口:%d", global.ServerConfig.Port)
	go func() {
		if err := Router.Run(fmt.Sprintf(":%d", global.ServerConfig.Port)); err != nil {
			zap.S().Panic("启动失败", err.Error())
		}
	}()
	//接受终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = register_client.DeRegister(serviceId); err != nil {
		zap.S().Panic("注销失败:", err.Error())
	} else {
		zap.S().Infof("注销成功")
	}
}
