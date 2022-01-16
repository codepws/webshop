// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package servicer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"webshop-service/goods-srv/dao"
	"webshop-service/goods-srv/model"

	pb "webshop-service/goods-srv/proto"

	"google.golang.org/protobuf/types/known/emptypb"
)

// greeterServer 定义一个结构体用于实现 .proto文件中定义的方法
// 新版本 gRPC 要求必须嵌入 pb.UnimplementedGreeterServer 结构体
type GoodsServicer struct {
	pb.UnimplementedGoodsServer
}

// 获取所有商品类别列表
func (user *GoodsServicer) GetAllCategorysList(ctx context.Context, request *emptypb.Empty) (*pb.CategoryListResponse, error) {

	log.Println("服务方法[GetAllCategorysList]： 获取所有商品类别")

	//新建用户, 表单验证，没有必要

	categorylist, err := dao.GetAllCategorysList()
	if err != nil {
		return nil, err
	}

	level1 := make([]*model.CategoryInfo, 0)
	level2 := make([]*model.CategoryInfo, 0)
	level3 := make([]*model.CategoryInfo, 0)
	//
	category_list_rsp := new(pb.CategoryListResponse)
	category_list_rsp.Total = int32(len(categorylist))
	category_list_rsp.Data = make([]*pb.CategoryInfoResponse, category_list_rsp.Total)

	//log.Printf("Total:%v,  Data.len=%v,  Data.cap=%v\n", category_list_rsp.Total, len(category_list_rsp.Data), cap(category_list_rsp.Data))

	for idx, category := range categorylist {

		category_rsp := new(pb.CategoryInfoResponse)
		category_rsp.Id = category.Id
		category_rsp.Name = category.Name
		category_rsp.ParentCategory = category.ParentCategoryId
		category_rsp.Level = uint32(category.Level)
		category_rsp.IsTab = category.IsTab

		//所有类别
		category_list_rsp.Data[idx] = category_rsp

		//类别归类（当前类别旗下的子类别）
		if category.Level == 1 {
			level1 = append(level1, category)
		} else if category.Level == 2 {
			level2 = append(level2, category)
		} else if category.Level == 3 {
			level3 = append(level3, category)
		}

	}

	//Level 2 旗下的所有子类别（即Level 3）
	for _, data_3 := range level3 {
		for _, data_2 := range level2 {
			if data_3.ParentCategoryId == data_2.Id {
				data_2.SubCategory = append(data_2.SubCategory, data_3)
			}
		}
	}
	//Level 1 旗下的所有子类别（即Level 2）
	for _, data_2 := range level2 {
		for _, data_1 := range level1 {
			if data_2.ParentCategoryId == data_1.Id {
				data_1.SubCategory = append(data_1.SubCategory, data_2)
			}
		}
	}

	//转成Jsno字符串格式
	jsonBytes, err := json.Marshal(level1)
	if err != nil {
		return nil, err
	}
	category_list_rsp.JsonData = string(jsonBytes)

	//log.Printf("Total:%v,  Data.len=%v,  Data.cap=%v\n", category_list_rsp.Total, len(category_list_rsp.Data), cap(category_list_rsp.Data))

	return category_list_rsp, nil
}

// 批量获取商品信息  Batch
func (user *GoodsServicer) GetGoodsListByIds(ctx context.Context, request *pb.GoodsInfoByIdsRequest) (goodsListResponse *pb.GoodsListResponse, err error) {

	log.Println("服务方法[GetAllCategorysList]： 获取所有商品类别")

	defer func() {

		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("GetGoodsListByIds 函数退出[%v:%v] ：%v\n", file, line, p)
			// 发生宕机时，获取panic传递的上下文并打印
			switch p.(type) {
			case runtime.Error: // 运行时错误
				//fmt.Println("runtime panic:", p)
			default: // 非运行时错误
				//fmt.Println("default panic:", p)
			}
			goodsListResponse = nil
			err = fmt.Errorf("%v", p)
		}
	}()

	//新建用户, 表单验证，没有必要

	goodslist, err := dao.GetGoodsListByIds(request.Ids)
	if err != nil {
		return nil, err
	}

	goods_list_rsp := new(pb.GoodsListResponse)

	goods_list_rsp.Total = uint32(len(goodslist))
	goods_list_rsp.Data = make([]*pb.GoodsInfoResponse, goods_list_rsp.Total)

	for idx, goods := range goodslist {
		goodsInfo := &pb.GoodsInfoResponse{
			Id:         goods.Id,
			CategoryId: goods.CategoryId,

			Name:        goods.Name,
			GoodsSn:     goods.GoodsSn,
			ClickNum:    goods.ClickNum,
			SoldNum:     goods.SoldNum,
			FavNum:      goods.FavNum,
			MarketPrice: goods.MarketPrice,
			ShopPrice:   goods.ShopPrice,
			GoodsBrief:  goods.GoodsBrief,
			//GoodsDesc:       goods.GoodsDescJson,
			ShipFree: goods.IsShipFree,
			//Images:          goods.ImagesJson,
			//DescImages:      goods.DescImagesJson,
			GoodsFrontImage: goods.GoodsFrontImage,
			IsNew:           goods.IsNew,
			IsHot:           goods.IsHot,
			ForSale:         goods.ForSale,
			//AddTime:goods.AddTime,
			//Category: goods.Category,
			//Brand:    goods.Brand,
		}

		//
		goods_list_rsp.Data[idx] = goodsInfo
	}

	//log.Printf("Total:%v,  Data.len=%v,  Data.cap=%v\n", category_list_rsp.Total, len(category_list_rsp.Data), cap(category_list_rsp.Data))

	return goods_list_rsp, nil
}

// 指定ID的商品详细信息
func (user *GoodsServicer) GetGoodsDetailById(ctx context.Context, request *pb.GoodInfoByIdRequest) (*pb.GoodsInfoResponse, error) {

	log.Println("服务方法[GetGoodsDetailById]： 获取所有商品类别")

	//新建用户, 表单验证，没有必要

	goods, err := dao.GetGoodsDetailById(request.Id)
	if err != nil {
		return nil, err
	}

	goodsInfo := &pb.GoodsInfoResponse{
		Id:         goods.Id,
		CategoryId: goods.CategoryId,

		Name:        goods.Name,
		GoodsSn:     goods.GoodsSn,
		ClickNum:    goods.ClickNum,
		SoldNum:     goods.SoldNum,
		FavNum:      goods.FavNum,
		MarketPrice: goods.MarketPrice,
		ShopPrice:   goods.ShopPrice,
		GoodsBrief:  goods.GoodsBrief,
		//GoodsDesc:       goods.GoodsDescJson,
		ShipFree: goods.IsShipFree,
		//Images:          goods.ImagesJson,
		//DescImages:      goods.DescImagesJson,
		GoodsFrontImage: goods.GoodsFrontImage,
		IsNew:           goods.IsNew,
		IsHot:           goods.IsHot,
		ForSale:         goods.ForSale,
		//AddTime:goods.AddTime,
		//Category: goods.Category,
		//Brand:    goods.Brand,
	}

	return goodsInfo, nil
}
