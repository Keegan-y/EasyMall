package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mxshop_srvs/inventory_srv/model"

	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func genMd5(code string) string {
	Md5 := md5.New()
	_, _ = io.WriteString(Md5, code)
	return hex.EncodeToString(Md5.Sum(nil))
}

func main() {
	dsn := "root:123456@tcp(192.168.124.4:3306)/mxshop_inventory_srv?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	// 全局模式
	var err error
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		panic(err)
	}
	//_ = DB.AutoMigrate(&model.StockSellDetail{})
	//orderDetail := model.StockSellDetail{
	//	OrderSn: "wam",
	//	Status:  1,
	//	Detail:  []model.GoodsDetail{{1, 2}, {2, 3}},
	//}
	//DB.Create(&orderDetail)
	var sellDetail model.StockSellDetail
	DB.Where(model.StockSellDetail{OrderSn: "wam"}).First(&sellDetail)
	fmt.Println(sellDetail.Detail)
}
