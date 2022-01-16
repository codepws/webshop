package main

import (
	"context"
	"fmt"
	"log"
	"os"
	pb "webshop-service/inventory-srv/proto"

	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

//http://10.0.101.68:8500/v1/agent/services
func main() {

	//命令行参数
	args := os.Args

	// 连接服务
	consul_host := "127.0.0.1"
	consul_port := 8500
	server_name := "inventory-srv"
	//参数target: grpc.Dial:的第一个参数，这个参数的主要作用的通过它来找到对应的服务端地址，target传入是一个字符串，统一格式为scheme://authority/endpoint，然后通过以下方式解析为Target struct
	grpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", consul_host, consul_port, server_name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
	)
	if err != nil {
		log.Fatal("[grpc.Dial] 连接失败", err)
	}
	// 实例化GPRC客户端
	grpcClient := pb.NewInventoryClient(grpcClientConnect)

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
		TestGetInvDetailByGoodsId(grpcClient)
	case "2":
		TestSetInv(grpcClient)
	case "3":
		TestSellInv(grpcClient)
	case "4":
		TestRebackInv(grpcClient)
	default:
		TestGetInvDetailByGoodsId(grpcClient)
		TestSetInv(grpcClient)
		TestSellInv(grpcClient)
		TestRebackInv(grpcClient)
	}
	//TestGetInvDetailByGoodsId(grpcClient)

	//TestSetInv(grpcClient)

}

func TestGetInvDetailByGoodsId(grpclient pb.InventoryClient) {

	log.Println("===================== TestGetInvDetailByGoodsId Begin =====================")

	goodsId := 3

	goodsIdRequest := &pb.GoodsId{
		GoodsId: uint32(goodsId),
	}

	ctx := context.Background()
	goodsInvResponse, err := grpclient.GetInvDetailByGoodsId(ctx, goodsIdRequest)
	if err != nil {
		log.Fatal("调用[GetInvDetailByGoodsId]失败:", err)
	}

	fmt.Printf("GoodsId=%d, Stocks=%d\n", goodsInvResponse.GoodsId, goodsInvResponse.Nums)

	log.Println("===================== TestGetInvDetailByGoodsId End =====================")
}

func TestSetInv(grpclient pb.InventoryClient) {

	log.Println("===================== TestSetInv Begin =====================")

	goodsId := 3
	stocks := 300

	goodsInvInfo := &pb.GoodsInvInfo{
		GoodsId: uint32(goodsId),
		Nums:    uint32(stocks),
	}

	ctx := context.Background()
	_, err := grpclient.SetInv(ctx, goodsInvInfo)
	if err != nil {
		log.Fatal("调用[SellInv]失败:", err)
	}

	log.Println("===================== TestSetInv End =====================")
}

func TestSellInv(grpclient pb.InventoryClient) {

	log.Println("===================== TestSellInv Begin =====================")

	orderSn := "1111111111111"

	count := 6
	goodsInfoList := make([]*pb.GoodsInvInfo, count)
	for i := 0; i < len(goodsInfoList); i++ {
		goodsInvInfo := &pb.GoodsInvInfo{
			GoodsId: uint32(i + 1),
			Nums:    uint32((i + 1) * 100),
		}
		goodsInfoList[i] = goodsInvInfo
	}

	sellInfo := &pb.SellInfo{
		GoodsInfo: goodsInfoList,
		OrderSn:   orderSn,
	}

	ctx := context.Background()
	_, err := grpclient.SellInv(ctx, sellInfo)
	if err != nil {
		log.Fatal("调用[SellInv]失败:", err)
	}

	log.Println("===================== TestSellInv End =====================")
}

func TestRInv(grpclient pb.InventoryClient) {

	log.Println("===================== TestSellInv Begin =====================")

	orderSn := "1111111111111"

	count := 6
	goodsInfoList := make([]*pb.GoodsInvInfo, count)
	for i := 0; i < len(goodsInfoList); i++ {
		goodsInvInfo := &pb.GoodsInvInfo{
			GoodsId: uint32(i + 1),
			Nums:    uint32((i + 1) * 100),
		}
		goodsInfoList[i] = goodsInvInfo
	}

	sellInfo := &pb.SellInfo{
		GoodsInfo: goodsInfoList,
		OrderSn:   orderSn,
	}

	ctx := context.Background()
	_, err := grpclient.SellInv(ctx, sellInfo)
	if err != nil {
		log.Fatal("调用[SellInv]失败:", err)
	}

	log.Println("===================== TestSellInv End =====================")
}

func TestRebackInv(grpclient pb.InventoryClient) {

	log.Println("===================== TestRebackInv Begin =====================")

	orderSn := "1111111111111"

	orderSnInfo := &pb.OrderSnInfo{
		OrderSn: orderSn,
	}

	ctx := context.Background()
	_, err := grpclient.RebackInv(ctx, orderSnInfo)
	if err != nil {
		//log.Fatal("调用[RebackInv]失败:", err)
		log.Printf("调用[RebackInv]失败: %T %v\n", err, err)

		status := status.Convert(err)
		if status.Code() == 111111 {
			log.Printf("Desc: %v\n", status.Message())
		}

		log.Printf("Status: %v      %v\n", status.Code(), status.Message())
	}

	log.Println("===================== TestRebackInv End =====================")
}
