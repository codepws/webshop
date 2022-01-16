package global

import (
	"webshop-service/inventory-srv/config"
	"webshop-service/inventory-srv/dao/sqldb"
)

var (
	AppConfig *config.AppConfig = &config.AppConfig{}

	DBMgr *sqldb.DBMgr = &sqldb.DBMgr{}

	//NacosConfig *config.NacosConfig = &config.NacosConfig{}

	//GoodsSrvClient proto.GoodsClient

	//OrderSrvClient proto.OrderClient

	//InventorySrvClient proto.InventoryClient
)
