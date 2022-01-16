package global

import (
	"webshop-api/user-web/config"
	"webshop-api/user-web/proto"
)

var (
	AppConfig *config.AppConfig = &config.AppConfig{}

	UserSrvGrpcClient proto.UserClient

	//GlobalMgrApp *GlobalMgr = &GlobalMgr{}

	//NacosConfig *config.NacosConfig = &config.NacosConfig{}

	//GoodsSrvClient proto.GoodsClient

	//OrderSrvClient proto.OrderClient

	//InventorySrvClient proto.InventoryClient
)
