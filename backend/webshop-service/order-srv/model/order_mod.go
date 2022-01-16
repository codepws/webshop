package model

//购物车中已选商品的ID和数量
type GoodsIdNumInfo struct {
	GoodsId uint32 `json:"goods_id,string" db:"goods_id"` //商品ID
	Nums    uint32 `json:"nums" db:"nums"`                //商品购买数量
}

//订单的商品详情
type OrderGoodsInfo struct {
	OrderId    uint32  `json:"order_id,string" db:"order_id"` //订单ID
	GoodsId    uint32  `json:"goods_id,string" db:"goods_id"` //商品ID
	GoodsName  string  `json:"goods_name" db:"goods_name"`    //商品名称
	GoodsPrice float64 `json:"goods_price" db:"goods_price"`  //商品价格（购买时）

	Nums       uint32 `json:"nums" db:"nums"`       //商品数量
	GoodsImage string `json:"version" db:"version"` //商品图片
}

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

type MQMessageBody struct {
	OrderSn      string `json:"order_sn"`       //订单编号
	UserId       uint32 `json:"user_id,string"` //用户ID
	Name         string `json:"name"`           //签收人
	Mobile       string `json:"mobile"`         //手机号
	Address      string `json:"address"`        //收货地址
	Post         string `json:"post"`           //留言
	ParentSpanId string `json:"parent_span_id"` //链路追踪
}
