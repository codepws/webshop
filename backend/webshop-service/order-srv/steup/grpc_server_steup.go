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
	pb "webshop-service/order-srv/proto"
	"webshop-service/order-srv/servicer"

	rls "github.com/envoyproxy/go-control-plane/envoy/service/ratelimit/v2"
	"github.com/juju/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func SteupGrpcReg() (*grpc.Server, error) {
	fmt.Println("======GRPC 订单服务创建======")

	//server相对来说启动比较简单,一般都会加拦截器来获取matedata或者去recover() panic，又或者打印一些日志
	//matedata: matedata是一个map[string][]string的结构，用来在客户端和服务器之间传输数据。其中的一个作用是可以传递分布式调用环境中的链路id，方便跟踪调试。另外也可以传一些业务相关的数据

	// create server
	var serverOptions []grpc.ServerOption
	//serverOptions = append(serverOptions, grpc.UnaryInterceptor(unaryServerInterceptor)) //一元拦截器
	//serverOptions = append(serverOptions, grpc.())

	//实例化grpc Server
	grpcServer := grpc.NewServer(serverOptions...)
	// 将服务描述(server)及其具体实现(greeterServer)注册到 gRPC 中去.
	// 内部使用的是一个 map 结构存储，类似 HTTP server。
	//注册grpc方法
	pb.RegisterOrderServer(grpcServer, &servicer.OrderServicer{})

	// Consul Grpc健康检测 默认检测策略 服务端注册一个叫xxxx.Health的服务，用于grpc的健康检查
	//hsrv := health.NewServer()
	//hsrv.SetServingStatus(register.Service, grpc_health_v1.HealthCheckResponse_SERVING)
	//grpc_health_v1.RegisterHealthServer(server, hsrv)
	// Consul Grpc健康检测 自定义检测策略
	healthServer := &HealthService{}                              //health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer) // Server HealthService

	// GRPC限流
	rls.RegisterRateLimitServiceServer(grpcServer, &ServerLimit{
		//默认令牌桶，fillInterval 每过多⻓时间向桶⾥放⼀个令牌，capacity 是桶的容量，超过桶容量的部分会被直接丢弃。
		bucket: ratelimit.NewBucket(time.Second*5, 6),

		//和默认方式一样，唯一不同是每次填充的令牌数是 quantum,而不是 1 个。
		//bucket: ratelimit.NewBucketWithQuantum(time.Second*5, 6, 2),
		//
		//bucket: ratelimit.NewBucketWithClock(time.Second*5, 6, nil),
		//默认一个间隔周期内就产生一个token，如果是高并发情况下，可以通过参数quantum控制产生多个。
		//第三个参数是一个clock  interface，主要是方便mock测试，如果传nil用的就是realClock{}
	})
	reflection.Register(grpcServer)

	return grpcServer, nil
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
