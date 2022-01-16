// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package steup

import (
	"context"
	"fmt"
	"log"
	"time"
	"webshop-api/user-web/common/global"
	"webshop-api/user-web/common/otgrpc"

	pb "webshop-api/user-web/proto"

	rls "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v2"
	//"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/juju/ratelimit"
	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
//consul_host string
//consul_port int
)

// GRPC连接user-srv
func SteupGrpcConnet() error {

	// 连接用户服务user-srv
	consulInfo := global.AppConfig.ConsulConfig
	//参数target: grpc.Dial:的第一个参数，这个参数的主要作用的通过它来找到对应的服务端地址，target传入是一个字符串，统一格式为scheme://authority/endpoint，然后通过以下方式解析为Target struct
	grpcClientConnect, err := grpc.Dial(
		fmt.Sprintf("consul://%s:%d/%s?healthy=true&wait=14s&insecure=true", consulInfo.Host, consulInfo.Port, global.AppConfig.UserSrvConfig.Name),
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		//一元拦截器，适用于普通rpc连接，相应的还有流拦截器。拦截器只有第一个生效，所以一般设置一个
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),

		grpc.WithBlock(), //grpc.Dial默认建立连接是异步的，加了这个参数后会等待所有连接建立成功后再返回
		//grpc.UseCompressor(gzip.Name), //在grpc.Dial参数中设置压缩的方式，PS：压缩方式客户端应该和服务端对应
	)
	if err != nil {
		//zap.S().Fatal("[SteupGrpcConnet] 连接 【用户服务】 失败")
		log.Println("[SteupGrpcConnet] 连接 【用户服务】 失败")
		return err
	}

	rateLimitServiceClient := rls.NewRateLimitServiceClient(grpcClientConnect)
	// Send a request to the server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rateLimitResponse, err := rateLimitServiceClient.ShouldRateLimit(ctx, &rls.RateLimitRequest{Domain: "envoy"})
	if err != nil {
		//log.Fatalf("could not call service: %v", err)
		log.Println("[SteupGrpcConnet] 调用 【用户服务】失败")
		return err
	}
	//log.Printf("response[%d]: %v", idx, r)
	if rateLimitResponse.OverallCode != rls.RateLimitResponse_OK {
		log.Printf("[SteupGrpcConnet] 限流 【用户服务】状态：%d\n", rateLimitResponse.OverallCode)
		return err
	}

	// 实例化GPRC客户端
	global.UserSrvGrpcClient = pb.NewUserClient(grpcClientConnect)

	return nil
}

// Service 健康检查实现
type HealthService struct{}

// Check 实现健康检查接口，这里直接返回健康状态，这里也可以有更复杂的健康检查策略，比如根据服务器负载来返回
func (h *HealthService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	fmt.Println("HealthService Check：", time.Now().Format("2006-01-02 13:04:05"))
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthService) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	fmt.Println("HealthService Watch：", time.Now().Format("2006-01-02 13:04:05"))
	return nil
}

type ServerLimit struct {
	bucket *ratelimit.Bucket
}

func (s *ServerLimit) ShouldRateLimit(ctx context.Context, request *rls.RateLimitRequest) (*rls.RateLimitResponse, error) {
	log.Printf("request begin: %v\n", request)
	// logic to rate limit every second request
	var overallCode rls.RateLimitResponse_Code
	if s.bucket.TakeAvailable(1) == 0 {
		overallCode = rls.RateLimitResponse_OVER_LIMIT
	} else {
		overallCode = rls.RateLimitResponse_OK
	}

	response := &rls.RateLimitResponse{OverallCode: overallCode}

	log.Printf("response[%v]: %v\n", request, response)
	return response, nil
}
