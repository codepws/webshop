// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package steup

import (
	"webshop-api/user-web/routers"

	//"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"

	"github.com/gin-gonic/gin"
	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
)

var (
//consul_host string
//consul_port int
)

// GRPC连接user-srv
func SteupWebRouter() (*gin.Engine, error) {

	//注册自定义表单验证规则
	routers.SetupValidator()

	//Step 2:
	router := routers.SetupRouter()

	return router, nil
}

/*
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
*/
