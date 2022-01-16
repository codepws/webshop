// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package servicer

import (
	"common/redislock"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"
	"webshop-service/inventory-srv/common/global"
	"webshop-service/inventory-srv/dao"
	"webshop-service/inventory-srv/model"
	pb "webshop-service/inventory-srv/proto"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// greeterServer 定义一个结构体用于实现 .proto文件中定义的方法
// 新版本 gRPC 要求必须嵌入 pb.UnimplementedGreeterServer 结构体
type InventoryServicer struct {
	pb.UnimplementedInventoryServer
}

//rpc GetInvDetailByGoodsId(GoodsId) returns (GoodsInvInfo); //获取库存信息
//rpc SetInv(GoodsInvInfo) returns (google.protobuf.Empty); //设置库存
//rpc SellInv(SellInfo) returns (google.protobuf.Empty); //扣减库存
//rpc RebackInv(SellInfo) returns(google.protobuf.Empty); //库存归还

// 指定商品的库存详情
func (inv *InventoryServicer) GetInvDetailByGoodsId(ctx context.Context, request *pb.GoodsId) (*pb.GoodsInvInfo, error) {

	log.Println("服务方法[GetInvDetail]：获取指定商品的库存详情")

	goodsInv, err := dao.GetInvDetailByGoodsId(request.GoodsId)
	if err != nil {
		return nil, err
	}

	rep := &pb.GoodsInvInfo{
		GoodsId: goodsInv.GoodsId,
		Nums:    goodsInv.Stocks,
	}

	return rep, nil
}

// 用于后台管理的创建商品库存 或 商城创建订单时的修改商品库存
func (inv *InventoryServicer) SetInv(ctx context.Context, request *pb.GoodsInvInfo) (*emptypb.Empty, error) {

	log.Println("服务方法[SetInv]：用于后台管理的创建商品库存 或 商城创建订单时的修改商品库存")

	//#这个接口是设置库存的。但是后面如果要修改库存，也可以使用这个接口
	goodsInvInfo := &model.GoodsInvInfo{
		GoodsId: request.GoodsId,
		Stocks:  request.Nums,
	}

	err := dao.SetInv(goodsInvInfo)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (inv *InventoryServicer) SellInv(ctx context.Context, request *pb.SellInfo) (empty *emptypb.Empty, err error) {

	log.Println("服务方法[SellInv]：订单售卖")
	defer func() {
		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("CreateOrder 函数退出[%v:%v] ：%v\n", file, line, p)
			// 发生宕机时，获取panic传递的上下文并打印
			switch p.(type) {
			case runtime.Error: // 运行时错误
				//fmt.Println("runtime panic:", p)
			default: // 非运行时错误
				//fmt.Println("default panic:", p)
			}
			err = fmt.Errorf("%v", p)
		}
	}()

	// var alterSql *sql.Stmt
	// alterSql, err = tx.Prepare(`ALTER TABLE inventory AUTO_INCREMENT=1;`)
	// if err != nil {
	// 	return nil, err
	// }
	// defer alterSql.Close()
	//if _, err := alterSql.Exec(); err != nil {
	//	return nil, err
	//}

	//开始事务
	var tx *sqlx.Tx
	tx, err = global.DBMgr.DB.Beginx() //dao.StartTransaction()
	if err != nil {
		return nil, err
	}
	defer func() {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("SellInv 函数退出[%v:%v]\n", file, line)
		if p := recover(); p != nil {
			log.Printf("SellInv 函数退出 ：%v\n", p)
			tx.Rollback()
			err = fmt.Errorf("%v", p) //panic(p) // re-throw panic after Rollback
		} else if err != nil {
			log.Printf("SellInv函数退出 rollback：%v\n", err)
			if err_rollback := tx.Rollback(); err_rollback != nil { // err is non-nil; don't change it
				log.Printf("rollback error：%v\n", err_rollback)
			}
		} else {
			log.Printf("函数退出 commit%v\n", p)
			if err_commit := tx.Commit(); err_commit != nil { // err is nil; if Commit returns error update err
				log.Printf("commit error：%v\n", err_commit)
			}
		}
	}()

	// -- 定义执行sql语句
	var selectSql *sqlx.Stmt
	selectSql, err = tx.Preparex(`select goods_id,stocks,version,freeze from inventory where goods_id=?;`)
	if err != nil {
		return nil, err
	}
	defer selectSql.Close()

	//
	var insertSql_inventory *sqlx.Stmt
	insertSql_inventory, err = tx.Preparex("INSERT inventory (goods_id, stocks, add_time) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE stocks=?;")
	if err != nil {
		return nil, err
	}
	defer insertSql_inventory.Close()

	//
	var insertSql_inventory_history *sqlx.Stmt
	insertSql_inventory_history, err = tx.Preparex("INSERT inventory_history (order_sn, order_inv_detail, add_time) VALUES (?, ?, NOW());")
	if err != nil {
		return nil, err
	}
	defer insertSql_inventory_history.Close()

	// -- 执行事务处理
	inv_detail_maps := make([]map[string]interface{}, len(request.GoodsInfo))
	// 购买商品数量列表
	fmt.Println("Step 1: 遍历购买商品数量，并扣减库存", len(request.GoodsInfo))
	for idx, goods := range request.GoodsInfo {
		//查询库存
		//redis_addr := fmt.Sprintf("%s:%d", global.AppConfig.Caches.MasterRedis.Host, global.AppConfig.Caches.MasterRedis.Port)
		//lock := redislock.NewRedisLock(redis_addr, global.AppConfig.Caches.MasterRedis.Password, global.AppConfig.Caches.MasterRedis.PoolSize, global.AppConfig.Caches.MasterRedis.MinIdleConns)
		lock := redislock.NewRedisLockByCluster(global.AppConfig.Caches.LockRedis)
		lockKeyName := fmt.Sprintf("lock:goods_%d", goods.GoodsId)
		if err = lock.Acquire(lockKeyName, time.Second*5); err != nil {
			return nil, err
		}
		defer lock.Release()

		//////////////////////////////////////////////////////////
		//获取指定商品的库存
		//var goodsInv *model.GoodsInvInfo
		//goodsInv, err = dao.GetInvDetailByGoodsId(goods.GoodsId)
		//if err != nil {
		//	return nil, err
		//}
		/*
			var goodsInv model.GoodsInvInfo
			err := selectSql.QueryRow(goods.GoodsId).Scan(&goodsInv.GoodsId, &goodsInv.Stocks, &goodsInv.Version, &goodsInv.Freeze)
			if err != nil && err != sql.ErrNoRows {
				dao.SimplePanic(errors.New("获取指定商品的库存错误"))
				return nil, err
			}
		*/

		goodsInv := new(model.GoodsInvInfo) //var goodsInv *model.GoodsInvInfo
		err = dao.TransactionGet(selectSql, goodsInv, goods.GoodsId)
		if err != nil && err != sql.ErrNoRows {
			dao.SimplePanic(errors.New("获取指定商品的库存错误"))
			return nil, err
		}

		fmt.Printf("商品[%d] -> 库存=%d, 扣减=%d\n", goods.GoodsId, goodsInv.Stocks, goods.Nums)
		//库存数量检测
		if goodsInv.Stocks < goods.Nums { //库存不足
			return nil, fmt.Errorf("商品[%d]库存不足 -> 库存=%d, 扣减=%d", goods.GoodsId, goodsInv.Stocks, goods.Nums)
		}

		fmt.Printf("         -> 开始扣减库存，库存剩余：%d\n", goodsInv.Stocks-goods.Nums)
		//////////////////////////////////////////////////////////
		//扣减库存
		goodsInv.Stocks -= goods.Nums
		err = dao.TransactionExec(insertSql_inventory, goodsInv.GoodsId, goodsInv.Stocks, goodsInv.Stocks) //dao.SetInv(goodsInv)
		if err != nil {
			dao.SimplePanic(err)
			return nil, err
		}

		//////////////////////////////////////////////////////////
		//商品预扣减信息
		goods_num_map := make(map[string]interface{}, 2)
		goods_num_map["goods_id"] = strconv.FormatUint(uint64(goods.GoodsId), 10)
		goods_num_map["num"] = strconv.FormatUint(uint64(goods.Nums), 10)
		//
		inv_detail_maps[idx] = goods_num_map

		//遍历下一个商品
	}

	fmt.Println("Step 2: 保存订单的所有商品出库历史记录")
	//
	var inv_detail_bytes []byte
	inv_detail_bytes, err = json.Marshal(inv_detail_maps)
	if err != nil {
		return nil, err
	}

	inventoryHistory := &model.InventoryHistory{}
	inventoryHistory.OrderSn = request.OrderSn //订单号
	inventoryHistory.OrderInvDetail = string(inv_detail_bytes)
	//inventoryHistory.Status = 1 //出库状态:1为"已扣减"
	err = dao.TransactionExec(insertSql_inventory_history, inventoryHistory.OrderSn, inventoryHistory.OrderInvDetail) //dao.AddInventoryHistory(inventoryHistory)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (inv *InventoryServicer) RebackInv(ctx context.Context, request *pb.OrderSnInfo) (empty *emptypb.Empty, err error) {

	log.Println("服务方法[RebackInv]：")

	empty = &emptypb.Empty{}

	//库存的归还， 有两种情况会归还： 1. 订单超时自动归还 2. 订单创建失败 ，需要归还之前的库存 3. 手动归还
	//开始事务
	var tx *sqlx.Tx
	tx, err = global.DBMgr.DB.Beginx() //dao.StartTransaction()
	if err != nil {
		return nil, err
	}
	defer func() {
		empty = &emptypb.Empty{}
		if p := recover(); p != nil {
			log.Printf("RebackInv 函数退出 ：%v\n", p)
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			log.Printf("RebackInv 函数退出 rollback：%v\n", err)

			if err_rollback := tx.Rollback(); err_rollback != nil { // err is non-nil; don't change it
				log.Printf("rollback error：%v\n", err_rollback)
			}

			err = status.Errorf(codes.Code(111111), err.Error())

		} else {
			log.Printf("RebackInv 函数退出 commit%v\n", p)

			if err_commit := tx.Commit(); err_commit != nil { // err is nil; if Commit returns error update err
				log.Printf("commit error：%v\n", err_commit)
			}

		}

	}()

	// -- 定义执行sql语句
	var selectSql_inventory_history *sqlx.Stmt
	selectSql_inventory_history, err = tx.Preparex(`select order_inv_detail from inventory_history where order_sn=? and status=1;`) //查询已扣减的
	if err != nil {
		return nil, err
	}
	defer selectSql_inventory_history.Close()

	//
	var selectSql_inventory *sqlx.Stmt
	selectSql_inventory, err = tx.Preparex(`select stocks from inventory where goods_id=?;`)
	if err != nil {
		return nil, err
	}
	defer selectSql_inventory.Close()

	//
	var updateSql_inventory *sqlx.Stmt
	updateSql_inventory, err = tx.Preparex("update inventory set stocks=stocks+? where goods_id=?;") //归还库存
	if err != nil {
		return nil, err
	}
	defer updateSql_inventory.Close()

	//
	var updateSql_inventory_history *sqlx.Stmt
	updateSql_inventory_history, err = tx.Preparex("update inventory_history set status=2 where order_sn=?;") // 出库状态: 2为"已归还"'
	if err != nil {
		return nil, err
	}
	defer updateSql_inventory.Close()

	// -- 执行事务处理

	//////////////////////////////////////////////////////////
	//获取指定的订单历史信息
	invHistory := new(model.InventoryHistory)
	err = dao.TransactionGet(selectSql_inventory_history, invHistory, request.OrderSn)

	if err != nil && err != sql.ErrNoRows {
		dao.SimplePanic(errors.New("获取指定订单的已扣减的商品错误"))
		return nil, err
	}

	if err == sql.ErrNoRows {
		return nil, errors.New("查无此订单")
	}

	fmt.Printf("订单[%v] -> 归还商品列表=%v\n", request.OrderSn, invHistory.OrderInvDetail)
	//fmt.Printf("订单[%d] -> 已扣减扣减=%d\n", request.OrderSn, goods.GoodsId, goodsInv.Stocks, goods.Num)

	//inv_detail_maps := make([]map[string]interface{}, len(request.GoodsInfo))
	//inv_detail_maps := make([]map[string]interface{}, 0)
	goodsNumInfoList := make([]*model.GoodsNumInfo, 8)
	err = json.Unmarshal([]byte(invHistory.OrderInvDetail), &goodsNumInfoList)
	if err != nil {
		dao.SimplePanic(err)
		return nil, err
	}

	//goods_num_map := make(map[string]interface{}, 2)
	//goods_num_map["goods_id"] = goods.GoodsId
	//goods_num_map["num"] = goods.Num
	//
	//inv_detail_maps[idx] = goods_num_map

	// 购买商品数量列表
	for _, goods := range goodsNumInfoList {

		goods_id := goods.GoodsId                                      //goods["goods_id"]
		goods_num, _ := strconv.ParseUint(goods.Nums.String(), 10, 32) //goods["num"]

		//查询库存
		//redis_addr := fmt.Sprintf("%s:%d", global.AppConfig.Caches.MasterRedis.Host, global.AppConfig.Caches.MasterRedis.Port)
		//lock := redislock.NewRedisLock(redis_addr, global.AppConfig.Caches.MasterRedis.Password, global.AppConfig.Caches.MasterRedis.PoolSize, global.AppConfig.Caches.MasterRedis.MinIdleConns)
		lock := redislock.NewRedisLockByCluster(global.AppConfig.Caches.LockRedis)
		lockKeyName := fmt.Sprintf("lock:goods_%d", goods_id)
		if err = lock.Acquire(lockKeyName, time.Second*5); err != nil {
			return nil, err
		}
		defer lock.Release()

		//查询库存
		//goodsInv := new(model.GoodsInvInfo) //var goodsInv *model.GoodsInvInfo
		var goodsInv_stocks uint64
		err = dao.TransactionGet(selectSql_inventory, &goodsInv_stocks, goods_id)
		if err != nil && err != sql.ErrNoRows {
			dao.SimplePanic(errors.New("获取指定商品的库存错误"))
			return nil, err
		}
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("商品[%d]无库存信息", goods_id)
		}

		fmt.Printf("商品[%v] -> 当前库存=%v, 归还=%v\n", goods_id, goodsInv_stocks, goods_num)

		//////////////////////////////////////////////////////////
		//归还库存
		//goodsInv_stocks += goods_num //goodsInv.Stocks
		//dao.SetInv(goodsInv)
		err = dao.TransactionExec(updateSql_inventory, goods_num, goods_id)
		if err != nil {
			dao.SimplePanic(err)
			return nil, err
		}

		//////////////////////////////////////////////////////////
		//商品预扣减信息

		//遍历下一个商品 updateSql_inventory_history
	}

	//inventoryHistory.Status = 1 //出库状态:2为"已归还"
	err = dao.TransactionExec(updateSql_inventory_history, request.OrderSn) //dao.AddInventoryHistory(inventoryHistory)
	if err != nil {
		dao.SimplePanic(err)
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

type OrderOrderSN struct {
	OrderSN string `json:"order_sn"` // 解析（encode/decode） 的时候，使用 `sname`，而不是 `Field`
}

func MQOrderInvRebackCallback(ctx context.Context, msgs ...*primitive.MessageExt) (consumeResult consumer.ConsumeResult, err error) {

	log.Println("======库存归还消息======开始")

	defer func() {

		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("MQOrderInvRebackCallback 函数异常[%v:%v] ：%v\n", file, line, p)
			// 发生宕机时，获取panic传递的上下文并打印
			switch p.(type) {
			case runtime.Error: // 运行时错误
			default: // 非运行时错误
			}

			//
			consumeResult = consumer.ConsumeRetryLater
			err = fmt.Errorf("%v", p)

		}
	}()

	for idx := range msgs {
		log.Printf("\nsubscribe callback: %v \n\n", msgs[idx])

		maxOffsetStr := msgs[idx].GetProperty(primitive.PropertyMaxOffset)
		maxOffset, _ := strconv.ParseInt(maxOffsetStr, 10, 64)
		offset := msgs[idx].QueueOffset
		diff := maxOffset - offset
		log.Printf("            偏移 maxOffset=%v, offset=%v，diff=%v\n", maxOffsetStr, offset, diff)
		if diff > 100000 {
			// TODO 消息堆积情况的特殊处理
			log.Printf("消息堆积情况的特殊处理，直接返回: maxOffset=%v, offset=%v\n", maxOffsetStr, offset)
			return consumer.ConsumeSuccess, nil
		}
		// TODO 正常消费过程
		log.Printf("消息[%v][%v]：%v\n", idx, msgs[idx].TransactionId, string(msgs[idx].Body))

		//通过msg的body中的order_sn来确定库存的归还
		// 订单号
		//orderOrderSN := new(OrderOrderSN)
		orderMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(msgs[idx].Body), &orderMap)
		if err != nil {
			log.Printf("解析消息[%v]Body失败: err=%v\n", idx, err)
			panic(err)
		}
		order_sn := orderMap["order_sn"] //orderOrderSN.OrderSN

		//为了要用事务来做： 我们查询库存扣减历史记录，并逐个归还商品库存
		// ======开启事务======
		tx, err := global.DBMgr.DB.Beginx()
		if err != nil {
			log.Printf("订单号[%v] 开始DB事务失败: err=%v\n", order_sn, err)
			panic(err)
		}
		defer func() {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("MQOrderInvRebackCallback 事务退出[%v:%v]\n", file, line)
			if p := recover(); p != nil {
				log.Printf("MQOrderInvRebackCallback 事务异常：%v\n", p)
				tx.Rollback()
				panic(p) // re-throw panic after Rollback
			} else if err != nil {
				log.Printf("MQOrderInvRebackCallback 错误 rollback：%v\n", err)
				if err_rollback := tx.Rollback(); err_rollback != nil { // err is non-nil; don't change it
					log.Printf("rollback error：%v\n", err_rollback)
				}

			} else {
				log.Printf("MQOrderInvRebackCallback 成功 commit%v\n", p)
				if err_commit := tx.Commit(); err_commit != nil { // err is nil; if Commit returns error update err
					log.Printf("commit error：%v\n", err_commit)
				}

			}
		}()

		////////////////////////////////////////////////////////////////////////////////////////
		//1. 为了防止没有扣减库存反而归还库存的情况，这里我们要先查询有没有库存扣减记录
		fmt.Println("Step 1: 查询订单支付状态，订单号:", order_sn)

		order_inv_history := new(model.InventoryHistory)
		if err := tx.Get(order_inv_history, "select order_sn, order_inv_detail, status from inventory_history where order_sn=? and status=0 and is_deleted=0", order_sn); err != nil && err != sql.ErrNoRows {
			log.Printf("查询库存扣减记录失败，订单号=%v Error=%v\n", order_sn, err)
			return consumer.ConsumeRetryLater, err
		}

		if err == sql.ErrNoRows {
			log.Printf("该订单库存扣减记录不存在，订单号=%v\n", order_sn)
			return consumer.ConsumeSuccess, nil
		}
		//出库状态:0为"已扣减", 1为"已归还
		if order_inv_history.Status == 0 {
			//关闭订单交易
			sqlResult, err := tx.Exec("update inventory_history set status=1, is_deleted=1 where order_sn=?", order_sn)
			if err != nil {
				log.Printf("关闭订单支付失败，订单号=%v Error=%v\n", order_sn, err)
				return consumer.ConsumeRetryLater, err
			}

			affected, err := sqlResult.RowsAffected()
			if err != nil {
				return consumer.ConsumeRetryLater, err
			}

			if affected == 0 {
				//Update未执行
				log.Printf("该订单的出库状态Update未执行，订单号=%v\n", order_sn)
				return consumer.ConsumeSuccess, nil
			}

			//琢个商品库存归还
			goodsNumInfoList := make([]model.GoodsNumInfo, 0)
			err = json.Unmarshal([]byte(order_inv_history.OrderInvDetail), &goodsNumInfoList)
			if err != nil {
				log.Printf("解析数据[%v]失败: err=%v\n", order_inv_history.OrderInvDetail, err)
				return consumer.ConsumeRetryLater, err
			}

			for idx := range goodsNumInfoList {

				log.Printf("该订单的商品[%v]库存归还数量为[%v]，订单号=%v\n", goodsNumInfoList[idx].GoodsId, goodsNumInfoList[idx].Nums, order_sn)

				sqlResult, err := tx.Exec("update inventory set stocks=stocks+? where goods_id=?", goodsNumInfoList[idx].Nums, goodsNumInfoList[idx].GoodsId)
				if err != nil {
					log.Printf("此商品库存归还失败，订单号=%v  商品ID=%v, Error=%v\n", order_sn, goodsNumInfoList[idx].GoodsId, err)
					return consumer.ConsumeRetryLater, err
				}

				affected, err := sqlResult.RowsAffected()
				if err != nil {
					return consumer.ConsumeRetryLater, err
				}

				if affected == 0 {
					//Update未执行
					log.Printf("此商品库存归还Update未执行，订单号=%v  商品ID=%v, Error=%v\n", order_sn, goodsNumInfoList[idx].GoodsId, err)
					return consumer.ConsumeSuccess, nil
				}

			}
		}

	}

	log.Println("======库存归还消息======结束")
	return consumer.ConsumeSuccess, nil
}
