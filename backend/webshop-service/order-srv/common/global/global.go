package global

import (
	"webshop-service/order-srv/config"
	"webshop-service/order-srv/dao/sqldb"
	"webshop-service/order-srv/proto"

	"github.com/apache/rocketmq-client-go/v2"
)

var (
	AppConfig *config.AppConfig = &config.AppConfig{}

	DBMgr *sqldb.DBMgr = &sqldb.DBMgr{}

	//NacosConfig *config.NacosConfig = &config.NacosConfig{}

	GoodsSrvClient     proto.GoodsClient
	InventorySrvClient proto.InventoryClient

	MQProducerOrderPayTimeout rocketmq.Producer
)
