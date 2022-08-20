package global

import (
	"mxshop_srvs/userop_srv/config"

	"github.com/go-redis/redis/v8"

	"gorm.io/gorm"
)

var (
	DB           *gorm.DB
	RDB          *redis.Client
	ServerConfig config.ServerConfig
	NacosConfig  = &config.NacosConfig{}
)

//func Init() {
//	dsn := "root:123456@tcp(192.168.124.4:3306)/mxshop_userop_srv?charset=utf8mb4&parseTime=True&loc=Local"
//	newLogger := logger.New(
//		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
//		logger.Config{
//			SlowThreshold: time.Second, // 慢 SQL 阈值
//			LogLevel:      logger.Info, // Log level
//			Colorful:      true,        // 禁用彩色打印
//		},
//	)
//
//	// 全局模式
//	var err error
//	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//		NamingStrategy: schema.NamingStrategy{
//			SingularTable: true,
//		},
//		Logger: newLogger,
//	})
//	if err != nil {
//		panic(err)
//	}
//
//}
