package global

import (
	ut "github.com/go-playground/universal-translator"
	"mxshop_api/userop-web/config"
	"mxshop_api/userop-web/proto"
)

var (
	ServerConfig *config.ServerConfig =&config.ServerConfig{}
	NacosConfig *config.NacosConfig = &config.NacosConfig{}
	Trans ut.Translator
	GoodsSrvClient proto.GoodsClient
	MessageClient proto.MessageClient
	AddressClient proto.AddressClient
	UserFavClient proto.UserFavClient
)
