package servicereg

import (
	utils "common/Utils"
	"math/rand"
	"strconv"
	"time"
)

/**
服务注册
在网络编程中，一般会提供项目的 IP、PORT、PROTOCOL，在服务治理中，我们还需要知道对应的服务名、实例名以及一些自定义的扩展信息
在这里使用ServiceInstance 接口来规定注册服务时必须的一些信息，同时用 DefaultServiceInstance 实现

定义接口
Represents an instance of a service in a discovery system.
*/
type ServiceInstance interface {

	// return The unique instance ID as registered.
	GetInstanceId() string

	// return The service name as registered.
	GetServiceName() string

	// return The hostname of the registered service instance.
	GetHost() string

	// return The port of the registered service instance.
	GetPort() int

	// return Whether the port of the registered service instance uses HTTPS.
	IsSecure() bool

	// return The service tags.
	GetTags() []string

	// return The key / value pair metadata associated with the service instance.
	GetMetadata() map[string]string

	// 健康检查方式：flase为GRPC方式检查，true为Http方式检查
	IsHttpCheck() bool
}

/**
 *
 */
type DefaultServiceInstance struct {
	instanceId  string //服务ID（唯一）
	serviceName string //服务名称
	host        string
	port        int
	secure      bool
	tags        []string
	metadata    map[string]string
	isHttpCheck bool //健康检查方式：flase为GRPC方式检查，true为Http方式检查，若为true，则 http路径为 actuator/health
}

/**
 * default
 */
func NewDefaultServiceInstance(serviceName string, host string, port int, secure bool, tags []string, metadata map[string]string, isHttpCheck bool, instanceId string) (*DefaultServiceInstance, error) {
	//metadata map[string]string

	if len(host) == 0 {
		localIP, err := utils.FindFirstNonLoopbackIP()
		if err != nil {
			return nil, err
		}
		host = localIP
	}

	if len(instanceId) == 0 {
		rand.Seed(time.Now().UnixNano())
		instanceId = serviceName + "-" + strconv.FormatInt(time.Now().Unix(), 10) + "-" + strconv.Itoa(rand.Intn(9000)+1000)
	}

	return &DefaultServiceInstance{instanceId: instanceId, serviceName: serviceName, host: host, port: port, secure: secure, tags: tags, isHttpCheck: isHttpCheck}, nil
}

func (serviceInstance DefaultServiceInstance) GetInstanceId() string {
	return serviceInstance.instanceId
}

func (serviceInstance DefaultServiceInstance) GetServiceName() string {
	return serviceInstance.serviceName
}

func (serviceInstance DefaultServiceInstance) GetHost() string {
	return serviceInstance.host
}

func (serviceInstance DefaultServiceInstance) GetPort() int {
	return serviceInstance.port
}

func (serviceInstance DefaultServiceInstance) IsSecure() bool {
	return serviceInstance.secure
}

func (serviceInstance DefaultServiceInstance) GetTags() []string {
	return serviceInstance.tags
}

func (serviceInstance DefaultServiceInstance) GetMetadata() map[string]string {
	return serviceInstance.metadata
}

func (serviceInstance DefaultServiceInstance) IsHttpCheck() bool {
	return serviceInstance.isHttpCheck
}
