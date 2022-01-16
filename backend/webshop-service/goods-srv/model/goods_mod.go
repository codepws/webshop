package model

//商品类别信息
type CategoryInfo struct {
	Id               uint32          `json:"id,string" db:"id"`
	Name             string          `json:"name" db:"name"`
	ParentCategoryId uint32          `json:"parent_category_id" db:"parent_category_id"`
	Level            uint8           `json:"level" db:"level"`
	IsTab            bool            `json:"is_tab" db:"is_tab"`
	SubCategory      []*CategoryInfo `json:"sub_category,omitempty"` //序列化的时候忽略0值或者空值
}

//商品品牌信息
type BrandInfo struct {
	Id   uint32 `json:"id,string" db:"id"`
	Name string `json:"name" db:"name"`
	Logo string `json:"logo" db:"logo"`
}

//商品详情
type GoodsInfo struct {
	Id              uint32  `json:"id,string" db:"id"`
	CategoryId      uint32  `json:"category_id" db:"category_id"` //外键：'商品类别ID'
	BrandsName      string  `json:"brands_name" db:"brands_name"` //外键： 品牌名称
	Name            string  `json:"name" db:"name"`
	GoodsSn         string  `json:"goods_sn" db:"goods_sn"`
	ForSale         bool    `json:"for_sale" db:"for_sale"`
	ClickNum        uint32  `json:"click_num" db:"click_num"`                 //点击数
	SoldNum         uint32  `json:"sold_num" db:"sold_num"`                   //商品销售量
	FavNum          uint32  `json:"fav_num" db:"fav_num"`                     //收藏数
	MarketPrice     float64 `json:"market_price" db:"market_price"`           //市场价格
	ShopPrice       float64 `json:"shop_price" db:"shop_price"`               //本店价格
	IsShipFree      bool    `json:"is_ship_free" db:"is_ship_free"`           //是否免运费：0为false, 非0为真
	IsNew           bool    `json:"is_new" db:"is_new"`                       //是否新品：0为false, 非0为真
	IsHot           bool    `json:"is_hot" db:"is_hot"`                       //是否热销：0为false, 非0为真
	GoodsBrief      string  `json:"goods_brief" db:"goods_brief"`             //商品简短描述
	GoodsFrontImage string  `json:"goods_front_image" db:"goods_front_image"` //商品封面图
	ImagesJson      string  `json:"images_json" db:"images_json"`             //商品轮播图(Json格式)'
	DescImagesJson  string  `json:"desc_images_json" db:"desc_images_json"`   //详情页图片(Json格式)
	IsDeleted       bool    `json:"is_deleted" db:"is_deleted"`               //是否删除：0为false, 非0为真
	UpdateTime      string  `json:"update_time" db:"update_time"`             //更新时间
	AddTime         string  `json:"add_time" db:"add_time"`                   //添加时间

	Category CategoryInfo `json:"category,omitempty" db:"category,omitempty"` //外键关联
	Brand    BrandInfo    `json:"brand,omitempty" db:"brand,omitempty"`       //外键关联
}
