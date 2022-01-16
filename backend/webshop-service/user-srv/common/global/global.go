package global

import (
	"webshop-service/user-srv/config"
	"webshop-service/user-srv/dao/sqldb"
)

var (
	AppConfig *config.AppConfig = &config.AppConfig{}

	DBMgr *sqldb.DBMgr = &sqldb.DBMgr{}

	//NacosConfig *config.NacosConfig = &config.NacosConfig{}

	//GoodsSrvClient proto.GoodsClient

	//OrderSrvClient proto.OrderClient

	//InventorySrvClient proto.InventoryClient
)
