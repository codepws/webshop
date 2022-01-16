package model

import "encoding/json"

// 商品库存详情
type GoodsInvInfo struct {
	GoodsId uint32 `json:"goods_id,string" db:"goods_id"` //商品ID
	Stocks  uint32 `json:"stocks" db:"stocks"`            //库存数量

	Version uint32 `json:"version" db:"version"` // 当前记录版本，分布式锁的乐观锁
	Freeze  uint32 `json:"freeze" db:"freeze"`   // 冻结数量
}

// 商品库存详情
type InventoryHistory struct {
	OrderSn        string `json:"order_sn" db:"order_sn"`                 //订单编号
	OrderInvDetail string `json:"order_inv_detail" db:"order_inv_detail"` //订单购买商品数量详情
	Status         uint32 `json:"status" db:"status"`                     // '出库状态:1为"已扣减", 2为"已归还"',
}

type GoodsNumInfo struct {
	GoodsId int32       `json:"goods_id,string"` //商品ID
	Nums    json.Number `json:"num"`             //商品购买数量
}
