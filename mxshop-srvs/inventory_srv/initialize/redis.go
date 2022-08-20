package initialize

import (
	"context"
	"fmt"
	"mxshop_srvs/inventory_srv/global"
	"time"

	"github.com/go-redis/redis/v8"
)

func InitRedis() {
	global.RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
		Password: global.ServerConfig.RedisInfo.Password, // no password set
		DB:       0,                                      // use default DB
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := global.RDB.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}
