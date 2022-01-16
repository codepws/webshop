package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	pb "webshop-service/goods-srv/proto"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
)

//http://10.0.101.68:8500/v1/agent/services
func main() {

	//命令行参数
	args := os.Args

	// 连接服务
	consul_host := "127.0.0.1"
	consul_port := 8500
	server_name := "goods-srv"
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
	grpcClient := pb.NewGoodsClient(grpcClientConnect)

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
		TestCategoryList(grpcClient)
	case "2":
		TestGetGoodsListByIds(grpcClient)
	case "3":
		TestGetGoodsDetailById(grpcClient)
	default:
		TestCategoryList(grpcClient)
		TestGetGoodsListByIds(grpcClient)
		TestGetGoodsDetailById(grpcClient)
	}

	//

}

func TestCategoryList(grpclient pb.GoodsClient) {

	log.Println("===================== TestCategoryList Begin =====================")

	ctx := context.Background()
	categoryListResponse, err := grpclient.GetAllCategorysList(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatal("调用[GetAllCategorysList]失败:", err)
	}

	fmt.Println("Total->", categoryListResponse.Total)

	fmt.Println("JsonData->", categoryListResponse.JsonData)

	jsonDataBytes, err := json.Marshal(categoryListResponse.Data)
	if err != nil {
		panic(err)
	}
	jsonDataString := string(jsonDataBytes)
	fmt.Println("Data->", jsonDataString)

	log.Println("===================== TestCategoryList End =====================")

}

func TestGetGoodsListByIds(grpclient pb.GoodsClient) {

	log.Println("===================== TestGetGoodsListByIds Begin =====================")

	ids := []uint32{1, 2, 5}

	goodsInfoByIdsRequest := &pb.GoodsInfoByIdsRequest{
		Ids: ids,
	}

	ctx := context.Background()
	goodsListResponse, err := grpclient.GetGoodsListByIds(ctx, goodsInfoByIdsRequest)
	if err != nil {
		log.Fatal("调用[GetGoodsListByIds]失败:", err)
	}

	fmt.Println("Total->", goodsListResponse.Total)

	jsonDataBytes, err := json.Marshal(goodsListResponse.Data)
	if err != nil {
		panic(err)
	}
	jsonDataString := string(jsonDataBytes)
	fmt.Println("Data->", jsonDataString)

	log.Println("===================== TestGetGoodsListByIds End =====================")

}

func TestGetGoodsDetailById(grpclient pb.GoodsClient) {

	log.Println("===================== TestGetGoodsDetailById Begin =====================")

	goodsId := 1

	goodInfoByIdRequest := &pb.GoodInfoByIdRequest{
		Id: uint32(goodsId),
	}

	ctx := context.Background()
	goodsResponse, err := grpclient.GetGoodsDetailById(ctx, goodInfoByIdRequest)
	if err != nil {
		log.Fatal("调用[GetGoodsDetailById]失败:", err)
	}

	jsonDataBytes, err := json.Marshal(goodsResponse)
	if err != nil {
		panic(err)
	}
	jsonDataString := string(jsonDataBytes)
	fmt.Println("Data->", jsonDataString)

	log.Println("===================== TestGetGoodsDetailById Begin =====================")

}
