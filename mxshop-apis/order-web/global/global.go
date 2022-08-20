package global

import (
	ut "github.com/go-playground/universal-translator"
	"mxshop_api/order-web/config"
	"mxshop_api/order-web/proto"
)

var (
	ServerConfig *config.ServerConfig =&config.ServerConfig{}
	NacosConfig *config.NacosConfig = &config.NacosConfig{}
	Trans ut.Translator
	GoodsSrvClient proto.GoodsClient
	OrderSrvClient proto.OrderClient
	InvSrvClient proto.InventoryClient
)
