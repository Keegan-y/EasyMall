package main

import (
	"crypto/md5"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mxshop_srvs/user_srv/model"
	"os"
	"time"

	"github.com/anaskhan96/go-password-encoder"

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
	dsn := "root:123456@tcp(192.168.124.4:3306)/mxshop_user_srv?charset=utf8mb4&parseTime=True&loc=Local"

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
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode("admin123", options)
	newpassword := fmt.Sprintf("$pdkdf2-sha512$%s$%s", salt, encodedPwd)
	for i := 0; i < 10; i++ {
		user := model.User{
			Mobile:   fmt.Sprintf("1818679199%d", i),
			Password: newpassword,
			NickName: fmt.Sprintf("wam%d", i),
		}
		DB.Save(&user)
	}
	//_ = db.AutoMigrate(&model.User{})
	//salt, encodedPwd := password.Encode("generic password", nil)
	//fmt.Println(salt)
	//fmt.Println(encodedPwd)
	//check := password.Verify("generic password", salt, encodedPwd, nil)
	//fmt.Println(check) // true

	//Using custom options
	//options := &password.Options{16, 100, 32, sha512.New}
	//salt, encodedPwd := password.Encode("generic password", options)
	//newpassword := fmt.Sprintf("$pdkdf2-sha512$%s$%s", salt, encodedPwd)
	//fmt.Println(len(newpassword))
	//fmt.Println(newpassword)
	//passwordInfo := strings.Split(newpassword, "$")
	//fmt.Println(passwordInfo)
	//check := password.Verify("generic password", passwordInfo[2], passwordInfo[3], options)
	//fmt.Println(check) // true
}
