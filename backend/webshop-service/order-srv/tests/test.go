package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	pb "webshop-service/order-srv/proto"

	"github.com/go-redis/redis"                  // /v8
	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"google.golang.org/grpc"
)

//http://10.0.101.68:8500/v1/agent/services
func main() {

	defer func() {
		//总结：panic配合recover使用，recover要在defer函数中直接调用才生效
		if p := recover(); p != nil {
			log.Printf("main 函数退出 ：%v\n", p)

		}
	}()

	//命令行参数
	args := os.Args

	// 连接服务
	consul_host := "127.0.0.1"
	consul_port := 8500
	server_name := "order-srv"
	//参数target: grpc.Dial:的第一个参数，这个参数的主要作用的通过它来找到对应的服务端地址，target传入是一个字符串，统一格式为scheme://authority/endpoint，然后通过以下方式解析为Target struct
	grpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", consul_host, consul_port, server_name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
		grpc.WithTimeout(3),
	)
	if err != nil {
		log.Fatal("[grpc.Dial] 连接失败", err)
	}

	log.Println("商品服务 连接成功")

	// 实例化GPRC客户端
	grpcClient := pb.NewOrderClient(grpcClientConnect)

	/////////////////////////////////////////////////////////
	//

	log.Println("命令行参数：", args)

	l := "0"
	//命令行参数
	if len(args) > 1 {
		l = args[1] //
	}

	switch l {
	case "1":
		TestCreateOrder(grpcClient)
	case "2":
		//TestSetInv(grpcClient)
	case "3":
		//TestSellInv(grpcClient)
	case "4":
		TestRedis()
	default:
		//TestCreateOrder(grpcClient)
		//TestSetInv(grpcClient)
		//TestSellInv(grpcClient)
		//TestRebackInv(grpcClient)
		log.Println("请输入正确命令")
	}

}

func TestCreateOrder(grpclient pb.OrderClient) {

	log.Println("===================== TestCreateOrder Begin =====================")

	userId := 2
	orderRequest := &pb.OrderRequest{
		UserId:  uint32(userId),
		Address: "上海市浦东新区XXX路XXX号",
		Mobile:  "13000000000",
		Name:    "小王",
		Post:    "尽快发货",
	}

	ctx := context.Background()
	orderInfoResponse, err := grpclient.CreateOrder(ctx, orderRequest)
	if err != nil {
		log.Fatal("调用[CreateOrder]失败:", err)
	}

	fmt.Printf("ID=%v, OrderSn=%v, Total=%v, Status=%v\n", orderInfoResponse.Id, orderInfoResponse.OrderSn, orderInfoResponse.Total, orderInfoResponse.Status)

	log.Println("===================== TestCreateOrder End =====================")
}

//功能：演示通过go-redis库+lua实现一个顺序自增的id发号器
//通常用在文件存储的目录计算等方面
func TestRedis() {

	log.Println("===================== Go-Redis Begin =====================")
	fmt.Println("初始化Redis")

	//var ctx = context.Background()
	redisDb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:6379",
		Password:     "123456", // no password set
		DB:           0,        // use default DB
		PoolSize:     10,
		MinIdleConns: 5,
	})

	_, err := redisDb.Ping().Result()
	if err != nil {
		log.Fatal("Redis连接失败：", err)
		return
	}

	//得到一个id
	key := "id:logid"
	luaId := redis.NewScript(`
local id_key = KEYS[1]
local current = redis.call('get',id_key)
if current == false then
    redis.call('set',id_key,1)
    return '1'
end
redis.log(redis.LOG_NOTICE,' current:'..current..':')
local result =  tonumber(current)+1
redis.log(redis.LOG_NOTICE,' result:'..result..':')
redis.call('set',id_key,result)
return tostring(result)
	`)

	n, err := luaId.Run(redisDb, []string{key}, 2).Result()
	if err != nil {
		log.Fatal("Redis调用Lua脚本失败：", err)
		return
	}

	var ret string = n.(string)
	retint, err := strconv.ParseInt(ret, 10, 64)
	if err != nil {
		log.Fatal("Redis调用Lua脚本失败：", err)
		return
	}

	fmt.Println("ID为", retint)

	log.Println("===================== Go-Redis End =====================")
}

/*
https://github.com/liuhongdi/digv17
lua代码的说明：
id_key变量作为存储的kv对的key
如果变量不存在，设置id_key值为1并返回
如果变量存在，值加1后返回
注意转为字符串形式后返回，方便java代码接收
*/
