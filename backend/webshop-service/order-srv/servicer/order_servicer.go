// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package servicer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"webshop-service/order-srv/common/global"
	"webshop-service/order-srv/model"
	pb "webshop-service/order-srv/proto"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 生成订单号（格式为：当前时间 + user_id + 随机数）
func generateOrderSn(userId uint32) string {
	randNumber := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(99-10) + 10
	return fmt.Sprintf("%s%d%d", time.Now().Format("20060102150405"), userId, randNumber) //f'{time.strftime("%Y%m%d%H%M%S")}{user_id}{Random().randint(10, 99)}'
}

type orderInfo struct {
	Id      uint32
	OrderSn string
	Total   float64
}

type executeResult struct {
	Code   codes.Code
	Detail string
	Order  *orderInfo
}

//
type MQTransactionListener struct {
	//localTrans *sync.Map
	//transactionIndex int32

	local_execute_map map[string]*executeResult
	sync.RWMutex
}

func NewMQTransactionListener() *MQTransactionListener {
	return &MQTransactionListener{
		//localTrans:        new(sync.Map),
		local_execute_map: make(map[string]*executeResult),
	}
}

//var local_execute_dict map[string]*executeResult
var local_producer_mq_transaction = NewMQTransactionListener()

/*
type MutexMap struct {
	local_execute_map map[string]*executeResult
	sync.RWMutex
}
func NewMutexMap() *MutexMap {
	return &MutexMap{
		local_execute_map: make(map[string]*executeResult),
	}
}
*/
//var local_producer_mq_transaction = NewMutexMap()

func (tl *MQTransactionListener) GetExecuteResult(msgkey string) (er *executeResult, ok bool) {
	tl.RLock()
	defer tl.RUnlock()

	//fmt.Printf("GetExecuteResult:  msgkey=[%v], key=[%v]  %v  %v\n", msgkey, key, msgkey == key, val)
	er, ok = tl.local_execute_map[msgkey]
	return
}

func (tl *MQTransactionListener) SetExecuteResult(msgkey string, er *executeResult) {
	tl.Lock()
	defer tl.Unlock()

	//fmt.Printf("SetExecuteResult:  msgkey=[%v], er=[%v]\n", msgkey, er)
	//
	tl.local_execute_map[msgkey] = er
}

func (tl *MQTransactionListener) RemoveExecuteResult(orderSn string) {
	tl.Lock()
	defer tl.Unlock()
	//
	delete(tl.local_execute_map, orderSn)
}

//执行本地事务业务逻辑
func (dl *MQTransactionListener) ExecuteLocalTransaction(msg *primitive.Message) (localTransactionState primitive.LocalTransactionState) {

	log.Printf("======执行本地事务业务逻辑 ExecuteLocalTransaction======开始")
	local_execute_result := &executeResult{}

	msg_keys := strings.Trim(msg.GetKeys(), " ")
	if msg_keys == "" {
		panic("消息Key值为空")
	}

	defer func() {

		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("ExecuteLocalTransaction 函数异常[%v:%v] ：%v\n", file, line, p)
			// 发生宕机时，获取panic传递的上下文并打印
			switch p.(type) {
			case runtime.Error: // 运行时错误
			default: // 非运行时错误
			}
			//local_execute_dict["unkonw"] = nil
			//err = status.Errorf(codes.Internal, "%v", p
			local_execute_result.Code = codes.Internal
			local_execute_result.Detail = fmt.Sprintf("%v", p)
			local_execute_result.Order = nil

		}
		//
		local_producer_mq_transaction.SetExecuteResult(msg_keys, local_execute_result)
	}()

	//nextIndex := atomic.AddInt32(&dl.transactionIndex, 1)
	//fmt.Printf("%v 执行本地事务业务逻辑 ExecuteLocalTransaction  nextIndex: %v for transactionID: %v\n", time.Now(), nextIndex, msg.TransactionId)
	//status := nextIndex % 3
	//dl.localTrans.Store(msg.TransactionId, primitive.LocalTransactionState(status+1))

	log.Printf("ExecuteLocalTransaction 事务ID[%s]：创建订单\n", msg.TransactionId)

	msg_inv_reback_body := &model.MQMessageBody{}
	err := json.Unmarshal([]byte(msg.Body), msg_inv_reback_body)
	if err != nil {
		panic(err)
	}

	log.Println("msg_orkder_inv_reback->body:", msg_inv_reback_body)

	// 用户ID
	user_id := msg_inv_reback_body.UserId
	// 订单号
	order_sn := msg_inv_reback_body.OrderSn
	//签收人
	name := msg_inv_reback_body.Name
	//手机号
	mobile := msg_inv_reback_body.Mobile
	//收货地址
	address := msg_inv_reback_body.Address
	//留言
	post := msg_inv_reback_body.Post

	// ======开启事务======
	tx, err := global.DBMgr.DB.Beginx()
	if err != nil {
		panic(err)
	}
	defer func() {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("ExecuteLocalTransaction 事务退出[%v:%v]\n", file, line)
		if p := recover(); p != nil {
			log.Printf("ExecuteLocalTransaction 事务异常：%v\n", p)
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			log.Printf("ExecuteLocalTransaction 错误 rollback：%v\n", err)
			if err_rollback := tx.Rollback(); err_rollback != nil { // err is non-nil; don't change it
				log.Printf("rollback error：%v\n", err_rollback)
			}
			//localTransactionState = primitive.RollbackMessageState
		} else {
			log.Printf("ExecuteLocalTransaction 成功 commit%v\n", p)
			if err_commit := tx.Commit(); err_commit != nil { // err is nil; if Commit returns error update err
				log.Printf("commit error：%v\n", err_commit)
			}
			//localTransactionState = primitive.CommitMessageState
		}
	}()

	////////////////////////////////////////////////////////////////////////////////////////
	//1. 获取购物车选中的所有商品
	fmt.Println("Step 1: 获取所有购物车记录所选中的商品，用户ID:", user_id)
	goodsIdNumInfoList := make([]*model.GoodsIdNumInfo, 0)
	if err := tx.Select(&goodsIdNumInfoList, "select goods_id, nums from shopping_cart where user_id=? and checked=1 and is_deleted=0", user_id); err != nil {
		//return nil, status.Error(codes.Internal, err.Error())
		log.Printf("获取所有购物车记录所选中的商品失败:用户ID=%v Error=%v\n", user_id, err)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.RollbackMessageState
	}

	//没有符合条件的商品，直接返回
	count := len(goodsIdNumInfoList)
	if count <= 0 {
		log.Printf("没有选中结算的商品:用户ID=%v\n", user_id)

		local_execute_result.Code = codes.NotFound
		local_execute_result.Detail = "没有选中结算的商品"
		return primitive.RollbackMessageState
	}

	goodsIDs := make([]uint32, count)
	for i := 0; i < count; i++ {
		goodsIDs[i] = goodsIdNumInfoList[i].GoodsId
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//2. 查询所有购买商品信息（主要为商品本店价格）
	fmt.Println("Step 2: 查询购买的商品列表，计算订单总金额和订单的商品列表，商品列表IDs:", goodsIDs)

	// 连接商品服务
	goodsGrpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", global.AppConfig.ConsulConfig.Host, global.AppConfig.ConsulConfig.Port, global.AppConfig.GoodsSrv.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
		grpc.WithTimeout(3),
	)
	if err != nil {
		log.Printf("[%s] 连接失败:%v\n", global.AppConfig.GoodsSrv.Name, err)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.RollbackMessageState
	}
	// 实例化商品服务GPRC客户端，远程调用 根据IDs获取商品列表
	global.GoodsSrvClient = pb.NewGoodsClient(goodsGrpcClientConnect)
	//商品ID列表
	goodsInfoByIdsRequest := &pb.GoodsInfoByIdsRequest{
		Ids: goodsIDs,
	}
	context_goods, _ := context.WithTimeout(context.Background(), time.Second*3)
	goodsListResponse, err := global.GoodsSrvClient.GetGoodsListByIds(context_goods, goodsInfoByIdsRequest)
	if err != nil {
		log.Printf("[%s] 调用[GetGoodsListByIds] 失败:%v\n", global.AppConfig.GoodsSrv.Name, err)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.RollbackMessageState
	}
	fmt.Println("商品列表个数", goodsListResponse.Total)

	/////////////////////////////////////////////
	count = len(goodsListResponse.Data)
	//订单总金额
	order_amount := 0.0
	//订单的商品列表（用于保存该订单的所有商品记录）
	order_goods_list := make([]*model.OrderGoodsInfo, count)

	//售卖商品列表（包含商品ID和购买商品的数量，用于扣减库存）
	goods_sell_info := make([]*pb.GoodsInvInfo, count)

	// goodsIdNumInfoList 与 goodsListResponse.Data Item的商品顺序一致（查询是按商品ID的升序排序）
	for idx := 0; idx < count; idx++ {
		goods := goodsListResponse.Data[idx]

		//订单总金额
		order_amount += goods.ShopPrice * float64(goodsIdNumInfoList[idx].Nums)
		//订单的商品列表
		order_goods_list[idx] = &model.OrderGoodsInfo{
			GoodsId:    goods.Id,
			GoodsName:  goods.Name,
			GoodsPrice: goods.ShopPrice,
			Nums:       goodsIdNumInfoList[idx].Nums,
			GoodsImage: goods.GoodsFrontImage,
		}
		//售卖信息
		goods_sell_info[idx] = &pb.GoodsInvInfo{
			GoodsId: goods.Id,                     //商品ID
			Nums:    goodsIdNumInfoList[idx].Nums, //购买商品的数量
		}
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//3. 扣减库存

	fmt.Println("Step 3: 开始扣减商品库存，并生成订单号:", order_sn)
	// 连接库存服务
	inventoryGrpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", global.AppConfig.ConsulConfig.Host, global.AppConfig.ConsulConfig.Port, global.AppConfig.InventorySrv.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
	)

	if err != nil {
		log.Printf("[%s] 连接失败:%v\n", global.AppConfig.InventorySrv.Name, err)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.RollbackMessageState
	}
	// 实例化商品服务GPRC客户端，远程调用 根据IDs获取商品列表
	global.InventorySrvClient = pb.NewInventoryClient(inventoryGrpcClientConnect)
	//售卖信息（扣减库存）
	sellInfo := &pb.SellInfo{
		OrderSn:   order_sn,
		GoodsInfo: goods_sell_info,
	}
	context_inv, _ := context.WithTimeout(context.Background(), time.Second*3)
	if _, err := global.InventorySrvClient.SellInv(context_inv, sellInfo); err != nil {
		log.Printf("[%s] 调用[SellInv] 失败:%v\n", global.AppConfig.InventorySrv.Name, err)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.RollbackMessageState
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//4. 创建订单记录
	fmt.Printf("Step 4: 创建订单记录，订单号:%v，订单总金额:%v\n", order_sn, order_amount)
	var sqlResult sql.Result
	var affected int64
	if sqlResult, err = tx.Exec("INSERT order_info (user_id, order_sn, order_mount, signer_name, singer_mobile, address, post, add_time) VALUES (?,?,?,?,?,?,?, NOW())",
		user_id,
		order_sn,
		order_amount,
		name,
		mobile,
		address,
		post); err != nil {
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}

	affected, err = sqlResult.RowsAffected()
	if err != nil {
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}

	if affected == 0 {
		//Insert未执行
		//return nil, fmt.Errorf("创建订单记录操作未执行, 用户ID：%v  订单号：%v", request.UserId, order_sn)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = fmt.Sprintf("创建订单记录操作未执行, 用户ID：%v  订单号：%v", user_id, order_sn)
		return primitive.CommitMessageState //库存归还
	}

	order_id, err := sqlResult.LastInsertId()
	if err != nil {
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}
	//批量插入订单商品表
	count = len(order_goods_list)
	for idx := 0; idx < count; idx++ {

		insert_stmt, err := tx.Preparex(global.DBMgr.DB.Rebind("INSERT order_goods (order_id, goods_id, goods_name, goods_price, nums, goods_image, add_time) VALUES (?,?,?,?,?,?,NOW())"))
		if err != nil {
			fmt.Println("Prepare error:", err)
			local_execute_result.Code = codes.Internal
			local_execute_result.Detail = err.Error()
			return primitive.CommitMessageState //库存归还
		}
		if sqlResult, err = insert_stmt.Exec(
			order_id,
			order_goods_list[idx].GoodsId,
			order_goods_list[idx].GoodsName,
			order_goods_list[idx].GoodsPrice,
			order_goods_list[idx].Nums,
			order_goods_list[idx].GoodsImage); err != nil {
			local_execute_result.Code = codes.Internal
			local_execute_result.Detail = err.Error()
			return primitive.CommitMessageState //库存归还
		}
		//var affected int64
		affected, err = sqlResult.RowsAffected()
		if err != nil {
			local_execute_result.Code = codes.Internal
			local_execute_result.Detail = err.Error()
			return primitive.CommitMessageState //库存归还
		}
		if affected == 0 { //Insert未执行
			//return nil, fmt.Errorf("批量插入订单商品操作未执行, 订单号：%v  商品ID：%v", order_sn, order_goods_list[idx].GoodsId)
			local_execute_result.Code = codes.Internal
			local_execute_result.Detail = fmt.Sprintf("批量插入订单商品操作未执行, 订单号：%v  商品ID：%v", order_sn, order_goods_list[idx].GoodsId)
			return primitive.CommitMessageState //库存归还
		}

	}

	////////////////////////////////////////////////////////////////////////////////////////
	//5. 删除购物车的记录
	fmt.Printf("Step 5: 删除购物车的记录，订单号:%v，订单总金额:%v\n", order_sn, order_amount)
	//var sqlResult sql.Result
	if sqlResult, err = tx.Exec("DELETE FROM shopping_cart WHERE user_id=? and checked=1", user_id); err != nil {
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}
	//var affected int64
	affected, err = sqlResult.RowsAffected()
	if err != nil {
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}
	if affected == 0 { //Insert未执行
		//return nil, fmt.Errorf("删除购物车的记录操作未执行, 用户ID：%v", request.UserId)
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = fmt.Sprintf("删除购物车的记录操作未执行, 用户ID：%v", user_id)
		return primitive.CommitMessageState //库存归还
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//6. 发送延时消息（订单支付的支付时间限制）
	fmt.Println("Step 6: 发送订单支付的支付时间限制消息")
	//RocketMQ分布式事务消息
	//订单支付超时 Producer

	producer_order_pay_timeout, err := rocketmq.NewProducer(
		producer.WithGroupName(global.AppConfig.RocketMQ.OrderPayTimeout.PGroupName), //group_p_order_pay_timeout	global.AppConfig.RocketMQ.OrderPayTimeout.GroupName
		producer.WithNsResolver(primitive.NewPassthroughResolver(global.AppConfig.RocketMQ.NameServers)),
		producer.WithRetry(2),
		//producer.WithInterceptor(UserFirstInterceptor(), UserSecondInterceptor()), //拦截器
	)
	if err != nil {
		fmt.Printf("NewProducer error: %s\n", err.Error())
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}
	//defer producer.Shutdown()

	if err := producer_order_pay_timeout.Start(); err != nil {
		fmt.Printf("start producer error: %s\n", err.Error())
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}

	//
	//Step 1: 发送订单支付超时半消息
	topic := global.AppConfig.RocketMQ.OrderPayTimeout.Topic // "topic_order_reback"
	tag := global.AppConfig.RocketMQ.OrderPayTimeout.Tag
	mq_pay_timeout_msg := &primitive.Message{}
	mq_pay_timeout_msg.Topic = topic
	mq_pay_timeout_msg.WithTag(tag)
	mq_pay_timeout_msg.WithKeys([]string{order_sn}) //订单号作为Key
	//设置5为超时时间1min
	mq_pay_timeout_msg.WithDelayTimeLevel(2)

	//
	msg_pay_timeout_body := &OrderOrderSN{
		OrderSN: order_sn,
	} //fmt.Sprintf(`{"order_sn": "%s"}`, order_sn)
	mq_pay_timeout_msg.Body, err = json.Marshal(msg_pay_timeout_body)
	if err != nil {
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}
	log.Println("MQ生产超时订单消息：", global.AppConfig.RocketMQ.OrderPayTimeout.PGroupName, topic, tag)

	//同步发送
	context, _ := context.WithTimeout(context.Background(), time.Second*1)
	res, err := producer_order_pay_timeout.SendSync(context, mq_pay_timeout_msg)
	if err != nil { //消息发送失败处理方式
		local_execute_result.Code = codes.Internal
		local_execute_result.Detail = err.Error()
		return primitive.CommitMessageState //库存归还
	}
	//​消息发送成功或者失败要打印消息日志，务必要打印SendResult和key字段。
	//发送成功会有多个状态，在sendResult里定义。以下对每个状态进行说明：
	switch res.Status {
	case primitive.SendOK: //消息发送成功。

	case primitive.SendFlushDiskTimeout: //消息发送成功但是服务器刷盘超时。
	case primitive.SendFlushSlaveTimeout: //消息发送成功，但是服务器同步到Slave时超时。
	case primitive.SendSlaveNotAvailable: //消息发送成功，但是此时Slave不可用。
	case primitive.SendUnknownError:
	default:

	}
	fmt.Printf("订单支付超时半消息发送成功，发送状态：%v %v\n", res.Status, res.MsgID)

	//producer.Shutdown()

	// ======结束事务======
	log.Println("======结束事务======", user_id, order_sn)

	create_order := &orderInfo{
		Id:      uint32(order_id),
		OrderSn: order_sn,
		Total:   order_amount,
	}
	local_execute_result.Code = codes.OK
	local_execute_result.Detail = ""
	local_execute_result.Order = create_order

	//time.Sleep(time.Second * 10)

	log.Printf("======执行本地事务业务逻辑 ExecuteLocalTransaction======结束\n")

	return primitive.RollbackMessageState //创建订单成功，无需库存归还
}

type OrderOrderSN struct {
	OrderSN string `json:"order_sn"` // 解析（encode/decode） 的时候，使用 `sname`，而不是 `Field`
}

func MQOrderPayTimeoutCallback(ctx context.Context, msgs ...*primitive.MessageExt) (consumeResult consumer.ConsumeResult, err error) {

	log.Println("======订单支付超时消息======开始")

	//log.Printf("消息内容：%v\n", msgs)

	defer func() {

		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("MQOrderPayTimeoutCallback 函数异常[%v:%v] ：%v\n", file, line, p)
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

		// 订单号
		orderOrderSN := new(OrderOrderSN)
		err := json.Unmarshal([]byte(msgs[idx].Body), orderOrderSN)
		if err != nil {
			log.Printf("解析消息[%v]Body失败: err=%v\n", idx, err)
			panic(err)
		}
		order_sn := orderOrderSN.OrderSN
		// ======开启事务======
		tx, err := global.DBMgr.DB.Beginx()
		if err != nil {
			log.Printf("订单号[%v] 开始DB事务失败: err=%v\n", order_sn, err)
			panic(err)
		}
		defer func() {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("MQOrderPayTimeoutCallback 事务退出[%v:%v]\n", file, line)
			if p := recover(); p != nil {
				log.Printf("MQOrderPayTimeoutCallback 事务异常：%v\n", p)
				tx.Rollback()
				panic(p) // re-throw panic after Rollback
			} else if err != nil {
				log.Printf("MQOrderPayTimeoutCallback 错误 rollback：%v\n", err)
				if err_rollback := tx.Rollback(); err_rollback != nil { // err is non-nil; don't change it
					log.Printf("rollback error：%v\n", err_rollback)
				}

			} else {
				log.Printf("MQOrderPayTimeoutCallback 成功 commit%v\n", p)
				if err_commit := tx.Commit(); err_commit != nil { // err is nil; if Commit returns error update err
					log.Printf("commit error：%v\n", err_commit)
				}

			}
		}()

		////////////////////////////////////////////////////////////////////////////////////////
		//1. 查询订单支付状态
		fmt.Println("Step 1: 查询订单支付状态，订单号:", order_sn)
		order_pay_status := 0
		if err := tx.Get(&order_pay_status, "select status from order_info where order_sn=? and is_deleted=0", order_sn); err != nil && err != sql.ErrNoRows {
			log.Printf("查询订单支付状态失败，订单号=%v Error=%v\n", order_sn, err)
			return consumer.ConsumeRetryLater, err
		}

		if err == sql.ErrNoRows {
			log.Printf("该订单不存在，订单号=%v\n", order_sn)
			return consumer.ConsumeSuccess, nil
		}
		//订单状态: 0为"交易创建" 1为"超时关闭" 2为"交易完成" 3为"交易结束"
		if order_pay_status != 2 {
			//关闭订单交易
			sqlResult, err := tx.Exec("update order_info set status=3, is_deleted=1 where order_sn=? and is_deleted=0", order_sn)
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
				log.Printf("该订单的订单状态Update未执行，订单号=%v\n", order_sn)
				return consumer.ConsumeSuccess, nil
			}

			//要给库存服务发送一个归还库存的消息

			////////////////////////////////////////////////////////////////////////////////////////
			//6. 发送延时消息（订单支付的支付时间限制）
			fmt.Println("Step 6: 发送订单支付超时-订单库存归消息")
			//RocketMQ同步消息
			//订单库存归还 Producer
			producer_order_pay_timeout, err := rocketmq.NewProducer(
				producer.WithGroupName(global.AppConfig.RocketMQ.OrderPayInvReback.PGroupName),
				producer.WithNsResolver(primitive.NewPassthroughResolver(global.AppConfig.RocketMQ.NameServers)),
				producer.WithRetry(2),
				//producer.WithInterceptor(UserFirstInterceptor(), UserSecondInterceptor()), //拦截器
			)
			if err != nil {
				fmt.Printf("NewProducer error: %s\n", err.Error())

				return consumer.ConsumeRetryLater, err
			}
			if err := producer_order_pay_timeout.Start(); err != nil {
				fmt.Printf("start producer error: %s\n", err.Error())
				return consumer.ConsumeRetryLater, err
			}
			//
			//Step 1: 发送订单支付超时-订单库存归还 半消息
			topic := global.AppConfig.RocketMQ.OrderPayInvReback.Topic // "topic_order_reback"
			tag := global.AppConfig.RocketMQ.OrderPayInvReback.Tag
			mq_pay_timeout_msg := &primitive.Message{}
			mq_pay_timeout_msg.Topic = topic
			mq_pay_timeout_msg.WithTag(tag)
			mq_pay_timeout_msg.WithKeys([]string{order_sn}) //订单号作为Key

			//
			msg_pay_timeout_body := &OrderOrderSN{
				OrderSN: order_sn,
			} //fmt.Sprintf(`{"order_sn": "%s"}`, order_sn)
			mq_pay_timeout_msg.Body, err = json.Marshal(msg_pay_timeout_body)
			if err != nil {
				return consumer.ConsumeRetryLater, err
			}
			log.Println("MQ生产订单支付超时-订单库存归消息：", global.AppConfig.RocketMQ.OrderPayInvReback.PGroupName, topic, tag)

			//同步发送
			context, _ := context.WithTimeout(context.Background(), time.Second*1)
			res, err := producer_order_pay_timeout.SendSync(context, mq_pay_timeout_msg)
			if err != nil { //消息发送失败处理方式
				return consumer.ConsumeRetryLater, err
			}
			//​消息发送成功或者失败要打印消息日志，务必要打印SendResult和key字段。
			//发送成功会有多个状态，在sendResult里定义。以下对每个状态进行说明：
			switch res.Status {
			case primitive.SendOK: //消息发送成功。

			case primitive.SendFlushDiskTimeout: //消息发送成功但是服务器刷盘超时。
			case primitive.SendFlushSlaveTimeout: //消息发送成功，但是服务器同步到Slave时超时。
			case primitive.SendSlaveNotAvailable: //消息发送成功，但是此时Slave不可用。
			case primitive.SendUnknownError:
			default:

			}
			fmt.Printf("订单支付超时-订单库存归半消息发送成功，发送状态：%v %v\n", res.Status, res.MsgID)

		}

	}

	log.Println("======订单支付超时消息======结束")
	return consumer.ConsumeSuccess, nil
}

//回查本地事务状态
func (dl *MQTransactionListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	log.Printf("回查本地事务状态CheckLocalTransaction查询订单是否入库: %v %v\n", msg.MsgId, msg.TransactionId)
	/*
		v, existed := dl.localTrans.Load(msg.TransactionId)
		if !existed {
			fmt.Printf("            unknow msg: %v, return Commit\n", msg)
			return primitive.CommitMessageState
		}

		state := v.(primitive.LocalTransactionState)
		switch state {
		case 1:
			fmt.Printf("checkLocalTransaction COMMIT_MESSAGE: %v\n", msg)
			return primitive.CommitMessageState
		case 2:
			fmt.Printf("checkLocalTransaction ROLLBACK_MESSAGE: %v\n", msg)
			return primitive.RollbackMessageState
		case 3:
			fmt.Printf("checkLocalTransaction unknow: %v\n", msg)
			return primitive.UnknowState
		default:
			fmt.Printf("checkLocalTransaction default COMMIT_MESSAGE: %v\n", msg)
			return primitive.CommitMessageState
		}
	*/

	msg_body := &model.MQMessageBody{}
	err := json.Unmarshal([]byte(msg.Body), msg_body)
	if err != nil {
		return primitive.RollbackMessageState
	}
	//
	user_id := msg_body.UserId
	//
	order_sn := msg_body.OrderSn

	//查询本地数据库 看一下order_sn的订单是否已经入库了
	count := 0
	err = global.DBMgr.DB.Select(&count, "select count(*) from order_info where user_id=? and order_sn=?", user_id, order_sn)
	if err != nil {
		fmt.Printf("回查本地事务状态CheckLocalTransaction查询订单是否入库error: %v\n", err)
		return primitive.UnknowState
	}

	if count > 0 {
		return primitive.RollbackMessageState //已存在，说明已经入库
	} else {
		return primitive.CommitMessageState //不存在
	}

}

// greeterServer 定义一个结构体用于实现 .proto文件中定义的方法
// 新版本 gRPC 要求必须嵌入 pb.UnimplementedGreeterServer 结构体
type OrderServicer struct {
	pb.UnimplementedOrderServer
}

// 购买下订单（MQ分布式事务方式）
func (order *OrderServicer) CreateOrder(ctx context.Context, request *pb.OrderRequest) (orderInfoResponse *pb.OrderInfoResponse, err error) {

	log.Println("======购买下单开始======", request.UserId, request.Name)
	//RocketMQ分布式事务消息
	//mq_transaction_listener := NewMQTransactionListener()
	//库存归还 Producer
	producer_order_inv_reback, err := rocketmq.NewTransactionProducer(
		NewMQTransactionListener(), //
		producer.WithGroupName(global.AppConfig.RocketMQ.OrderCreateInvReback.PGroupName),
		producer.WithNsResolver(primitive.NewPassthroughResolver(global.AppConfig.RocketMQ.NameServers)),
		producer.WithRetry(2),
		//producer.WithInterceptor(UserFirstInterceptor(), UserSecondInterceptor()), //拦截器
	)
	if err != nil {
		fmt.Printf("NewProducer error: %s\n", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}
	if err := producer_order_inv_reback.Start(); err != nil {
		fmt.Printf("start producer error: %s\n", err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	/////////////////////////////////////////////////////////////////////
	//生成订单号
	order_sn := generateOrderSn(request.UserId)

	//
	//Step 1: 发送扣减库存的归还半消息
	topic := global.AppConfig.RocketMQ.OrderCreateInvReback.Topic // "topic_order_reback"
	tag := global.AppConfig.RocketMQ.OrderCreateInvReback.Tag
	msg_keys := order_sn
	msg := &primitive.Message{}
	msg.Topic = topic
	msg.WithTag(tag)
	msg.WithKeys([]string{msg_keys}) // Keys的格式为：[Key1 Key2 Key3 ]（注意最后一个Key后面多了一个空格）     订单号作为消息Key（唯一）
	//
	msg_body := &model.MQMessageBody{
		OrderSn: order_sn,
		UserId:  request.UserId,
		Name:    request.Name,
		Mobile:  request.Mobile,
		Address: request.Address,
		Post:    request.Post,
		//ParentSpanId: parent_span.context.span_id,
	}
	msg.Body, err = json.Marshal(msg_body)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//
	context, _ := context.WithTimeout(context.Background(), time.Second*1)
	res, err := producer_order_inv_reback.SendMessageInTransaction(context, msg)
	if err != nil {
		//消息发送失败处理方式
		fmt.Printf("\nsend message error: %s\n", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	//fmt.Printf("\nsend message success: result=%s\n", res.String())
	//​消息发送成功或者失败要打印消息日志，务必要打印SendResult和key字段。
	//发送成功会有多个状态，在sendResult里定义。以下对每个状态进行说明：
	switch res.Status {
	case primitive.SendOK: //消息发送成功。

	case primitive.SendFlushDiskTimeout: //消息发送成功但是服务器刷盘超时。
	case primitive.SendFlushSlaveTimeout: //消息发送成功，但是服务器同步到Slave时超时。
	case primitive.SendSlaveNotAvailable: //消息发送成功，但是此时Slave不可用。
	case primitive.SendUnknownError:
	default:

	}
	fmt.Printf("库存归还半消息发送成功，发送状态：%v %v\n", res.Status, res.MsgID)

	//Step 2: 执行本地事务业务逻辑
	//TODO: 异步

	for {
		//Step 3: 下单结束
		executeResult, ok := local_producer_mq_transaction.GetExecuteResult(msg_keys)
		if ok {

			//删除
			defer local_producer_mq_transaction.RemoveExecuteResult(msg_keys)

			log.Println("======购买下单结束======", request.UserId, order_sn, msg_keys)

			if executeResult.Code != codes.OK {
				fmt.Printf("下单失败 error: %s\n", executeResult.Detail)
				return nil, status.Error(executeResult.Code, executeResult.Detail)
			}

			orderInfoResponse = &pb.OrderInfoResponse{
				Id:      executeResult.Order.Id,
				OrderSn: executeResult.Order.OrderSn,
				Total:   executeResult.Order.Total,
			}

			//if err := producer.Shutdown(); err != nil {
			//	fmt.Printf("Producer.Shutdown error: %v\n", err)
			//}

			return orderInfoResponse, nil

		}
		//
		time.Sleep(time.Millisecond * 100)

		//fmt.Printf("延时等待购买下单结束...\n")
	}

	//return nil, nil
}

// 购买下订单（普通方式）
func (order *OrderServicer) CreateOrderEx(ctx context.Context, request *pb.OrderRequest) (orderInfoResponse *pb.OrderInfoResponse, err error) {

	log.Println("服务方法[CreateOrder]：创建订单")

	defer func() {

		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			_, file, line, _ := runtime.Caller(1)
			log.Printf("CreateOrder 函数退出[%v:%v] ：%v\n", file, line, p)
			// 发生宕机时，获取panic传递的上下文并打印
			switch p.(type) {
			case runtime.Error: // 运行时错误
			default: // 非运行时错误
			}
			orderInfoResponse = nil
			err = status.Errorf(codes.Internal, "%v", p)
		}
	}()

	// ======开启事务======
	tx, err := global.DBMgr.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("CreateOrder 退出[%v:%v]\n", file, line)
		if p := recover(); p != nil {
			log.Printf("CreateOrder 异常：%v\n", p)
			tx.Rollback()
			orderInfoResponse = nil
			err = status.Errorf(codes.Internal, "%v", p) //panic(p) // re-throw panic after Rollback
		} else if err != nil {
			log.Printf("CreateOrder 错误 rollback：%v\n", err)
			if err_rollback := tx.Rollback(); err_rollback != nil { // err is non-nil; don't change it
				log.Printf("rollback error：%v\n", err_rollback)
			}
			//err = status.Errorf(codes.Code(111111), err.Error())
		} else {
			log.Printf("CreateOrder 成功 commit%v\n", p)
			if err_commit := tx.Commit(); err_commit != nil { // err is nil; if Commit returns error update err
				log.Printf("commit error：%v\n", err_commit)
			}
		}
	}()

	////////////////////////////////////////////////////////////////////////////////////////
	//1. 获取购物车选中的所有商品
	fmt.Println("Step 1: 获取所有购物车记录所选中的商品，用户ID:", request.UserId)
	goodsIdNumInfoList := make([]*model.GoodsIdNumInfo, 0)
	if err := tx.Select(&goodsIdNumInfoList, "select goods_id, nums from shopping_cart where user_id=? and checked=1 and is_deleted=0", request.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//没有符合条件的商品，直接返回
	count := len(goodsIdNumInfoList)
	if count <= 0 {
		return nil, status.Error(codes.NotFound, "没有符合条件的商品")
	}

	goodsIDs := make([]uint32, count)
	for i := 0; i < count; i++ {
		goodsIDs[i] = goodsIdNumInfoList[i].GoodsId
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//2. 查询所有购买商品信息（主要为商品本店价格）
	fmt.Println("Step 2: 查询购买的商品列表，计算订单总金额和订单的商品列表，商品列表IDs:", goodsIDs)

	// 连接商品服务
	goodsGrpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", global.AppConfig.ConsulConfig.Host, global.AppConfig.ConsulConfig.Port, global.AppConfig.GoodsSrv.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
	)
	if err != nil {
		log.Printf("[%s] 连接失败:%v\n", global.AppConfig.GoodsSrv.Name, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	// 实例化商品服务GPRC客户端，远程调用 根据IDs获取商品列表
	global.GoodsSrvClient = pb.NewGoodsClient(goodsGrpcClientConnect)
	//商品ID列表
	goodsInfoByIdsRequest := &pb.GoodsInfoByIdsRequest{
		Ids: goodsIDs,
	}
	goodsListResponse, err := global.GoodsSrvClient.GetGoodsListByIds(context.Background(), goodsInfoByIdsRequest)
	if err != nil {
		log.Printf("[%s] 调用[GetGoodsListByIds] 失败:%v\n", global.AppConfig.GoodsSrv.Name, err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	fmt.Println("商品列表个数", goodsListResponse.Total)

	/////////////////////////////////////////////
	count = len(goodsListResponse.Data)
	//订单总金额
	order_amount := 0.0
	//订单的商品列表（用于保存该订单的所有商品记录）
	order_goods_list := make([]*model.OrderGoodsInfo, count)

	//售卖商品列表（包含商品ID和购买商品的数量，用于扣减库存）
	goods_sell_info := make([]*pb.GoodsInvInfo, count)

	// goodsIdNumInfoList 与 goodsListResponse.Data Item的商品顺序一致（查询是按商品ID的升序排序）
	for idx := 0; idx < count; idx++ {
		goods := goodsListResponse.Data[idx]

		//订单总金额
		order_amount += goods.ShopPrice * float64(goodsIdNumInfoList[idx].Nums)
		//订单的商品列表
		order_goods_list[idx] = &model.OrderGoodsInfo{
			GoodsId:    goods.Id,
			GoodsName:  goods.Name,
			GoodsPrice: goods.ShopPrice,
			Nums:       goodsIdNumInfoList[idx].Nums,
			GoodsImage: goods.GoodsFrontImage,
		}
		//售卖信息
		goods_sell_info[idx] = &pb.GoodsInvInfo{
			GoodsId: goods.Id,                     //商品ID
			Nums:    goodsIdNumInfoList[idx].Nums, //购买商品的数量
		}
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//3. 扣减库存
	//生成订单号
	order_sn := generateOrderSn(request.UserId)
	fmt.Println("Step 3: 开始扣减商品库存，并生成订单号:", order_sn)
	// 连接库存服务
	inventoryGrpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", global.AppConfig.ConsulConfig.Host, global.AppConfig.ConsulConfig.Port, global.AppConfig.InventorySrv.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
	)
	if err != nil {
		log.Printf("[%s] 连接失败:%v\n", global.AppConfig.InventorySrv.Name, err)
		return nil, err
	}
	// 实例化商品服务GPRC客户端，远程调用 根据IDs获取商品列表
	global.InventorySrvClient = pb.NewInventoryClient(inventoryGrpcClientConnect)
	//售卖信息
	sellInfo := &pb.SellInfo{
		OrderSn:   order_sn,
		GoodsInfo: goods_sell_info,
	}
	if _, err := global.InventorySrvClient.SellInv(context.Background(), sellInfo); err != nil {
		log.Printf("[%s] 调用[SellInv] 失败:%v\n", global.AppConfig.InventorySrv.Name, err)
		return nil, err
	}

	////////////////////////////////////////////////////////////////////////////////////////
	//4. 创建订单记录
	fmt.Printf("Step 4: 创建订单记录，订单号:%v，订单总金额:%v", order_sn, order_amount)
	var sqlResult sql.Result
	var affected int64
	if sqlResult, err = tx.Exec("INSERT order_info (user_id, order_sn, order_mount, signer_name, singer_mobile, address, post, add_time) VALUES (?,?,?,?,?,?,?, NOW())",
		request.UserId,
		order_sn,
		order_amount,
		request.Name,
		request.Mobile,
		request.Address,
		request.Post); err != nil {
		return nil, err
	}

	affected, err = sqlResult.RowsAffected()
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		//Insert未执行
		return nil, fmt.Errorf("创建订单记录操作未执行, 用户ID：%v  订单号：%v", request.UserId, order_sn)
	}

	order_id, err := sqlResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	//批量插入订单商品表
	count = len(order_goods_list)
	for idx := 0; idx < count; idx++ {

		insert_stmt, err := tx.Preparex(global.DBMgr.DB.Rebind("INSERT order_goods (order_id, goods_id, goods_name, goods_price, nums, goods_image, add_time) VALUES (?,?,?,?,?,?,NOW())"))
		if err != nil {
			fmt.Println("Prepare error:", err)
			return nil, status.Error(codes.Internal, err.Error())
		}
		//if sqlResult, err = tx.Exec("INSERT order_goods (order_id, goods_id, goods_name, goods_price, nums, goods_image, add_time) VALUES (?,?,?,?,?,?,NOW())",
		//var sqlResult sql.Result
		if sqlResult, err = insert_stmt.Exec(
			order_id,
			order_goods_list[idx].GoodsId,
			order_goods_list[idx].GoodsName,
			order_goods_list[idx].GoodsPrice,
			order_goods_list[idx].Nums,
			order_goods_list[idx].GoodsImage); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		//var affected int64
		affected, err = sqlResult.RowsAffected()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if affected == 0 { //Insert未执行
			return nil, fmt.Errorf("批量插入订单商品操作未执行, 订单号：%v  商品ID：%v", order_sn, order_goods_list[idx].GoodsId)
		}

	}

	////////////////////////////////////////////////////////////////////////////////////////
	//5. 删除购物车的记录
	fmt.Printf("Step 5: 删除购物车的记录，订单号:%v，订单总金额:%v", order_sn, order_amount)
	//var sqlResult sql.Result
	if sqlResult, err = tx.Exec("DELETE FROM shopping_cart WHERE user_id=? and checked=1", request.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	//var affected int64
	affected, err = sqlResult.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if affected == 0 { //Insert未执行
		return nil, fmt.Errorf("删除购物车的记录操作未执行, 用户ID：%v", request.UserId)
	}

	// ======结束事务======
	orderInfoResponse = &pb.OrderInfoResponse{

		OrderSn: order_sn,
		Total:   order_amount,

		UserId:  request.UserId,
		Address: request.Address,
		Mobile:  request.Mobile,
		Name:    request.Name,
		Post:    request.Post,
	}

	return orderInfoResponse, nil
}
