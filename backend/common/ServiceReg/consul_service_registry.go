package servicereg

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/consul/api"
)

// ConsulRegister consul service register
type ConsulServiceRegistry struct {
	/*
		Address                        string        // Consul地址（格式为 IP:Port，示例：127.0.0.1:8080）
		Service                        string        // 服务名称
		Tags                           []string      // 服务Tags
		IP                             string        // 服务IP（注意：默认为空串，自动获取IP）
		Port                           int           // 服务端口号
		BalanceFactor                  int           //
		Timeout                        time.Duration //Check超时时间
		Interval                       time.Duration //Check健康检查时间间隔
		DeregisterCriticalServiceAfter time.Duration //Check故障检查失败默认30s后 consul自动将注册服务删除	// 注销时间，相当于过期时间
	*/

	serviceInstances     map[string]map[string]ServiceInstance
	client               api.Client
	localServiceInstance ServiceInstance
}

// new a consulServiceRegistry instance
// token is optional  NewConsulRegister
func NewConsulServiceRegistry(host string, port int, token string) (*ConsulServiceRegistry, error) {
	if len(host) < 3 {
		return nil, errors.New("check host")
	}

	if port <= 0 || port > 65535 {
		return nil, errors.New("check port, port should between 1 and 65535")
	}

	config := api.DefaultConfig()
	config.Address = host + ":" + strconv.Itoa(port)
	config.Token = token
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	client.Agent().Services()

	return &ConsulServiceRegistry{client: *client}, nil
}

///////////////////////////////////////////////////////////////////////////////////////
// 服务注册
//

// Register register service
func (csr *ConsulServiceRegistry) Register(serviceInstance ServiceInstance) error {

	var tags []string //tag，可以为空
	if serviceInstance.IsSecure() {
		tags = append(tags, "secure=true")
	} else {
		tags = append(tags, "secure=false")
	}

	//if serviceInstance.GetMetadata() != nil {
	//	for key, value := range serviceInstance.GetMetadata() {
	//		tags = append(tags, key+"="+value)
	//	}
	//}

	//服务IP
	//if csr.IP == "" {
	//	csr.IP = utils.FindFirstNonLoopbackIP()()
	//}

	// 生成注册对象 创建注册到consul的服务
	registration := new(api.AgentServiceRegistration)
	registration.ID = serviceInstance.GetInstanceId()    //服务实例唯一标识
	registration.Name = serviceInstance.GetServiceName() //服务名称

	registration.Address = serviceInstance.GetHost() //服务IP
	registration.Port = serviceInstance.GetPort()    //服务端口 因为要运行多个实例，端口不能在appsettings.json里配置，在docker容器运行时传入

	registration.Tags = tags                          //tag，可以为空
	registration.Meta = serviceInstance.GetMetadata() //Meta信息

	// 生成对应的检查对象 增加consul健康检查回调函数
	check := new(api.AgentServiceCheck)

	if serviceInstance.IsHttpCheck() {
		// http 健康检测支持
		schema := "http"
		if serviceInstance.IsSecure() {
			schema = "https"
		}
		http_check := fmt.Sprintf("%s://%s:%d/actuator/health", schema, registration.Address, registration.Port) // HTTP健康检查地址
		fmt.Println("Http健康检测", http_check)
		check.HTTP = http_check // HTTP健康检查地址
	} else {
		// grpc 健康检查支持，执行健康检查的地址，service 会传到 Health.Check 函数中
		//check.GRPC = fmt.Sprintf("%s:%d", registration.Address, registration.Port) // GRPC健康检查地址
		grpc_check := fmt.Sprintf("%s:%d/%s", registration.Address, registration.Port, registration.Name)
		fmt.Println("GRPC健康检测", grpc_check)
		check.GRPC = grpc_check
		check.GRPCUseTLS = serviceInstance.IsSecure()
	}

	check.Timeout = "5s"                         //超时时间
	check.Interval = "3s"                        //健康检查时间间隔，健康检查间隔
	check.DeregisterCriticalServiceAfter = "10s" //故障检查失败30s后 consul自动将注册服务删除	// 注销时间，相当于过期时间
	//
	registration.Check = check //健康检查

	//服务注册 注册服务到Consul
	err := csr.client.Agent().ServiceRegister(registration)
	if err != nil {
		fmt.Println(err)
		//panic(err)
		return err
	}

	if csr.serviceInstances == nil {
		csr.serviceInstances = map[string]map[string]ServiceInstance{}
	}
	services := csr.serviceInstances[serviceInstance.GetServiceName()]
	if services == nil {
		services = map[string]ServiceInstance{}
	}
	//同一服务多个实例
	services[serviceInstance.GetInstanceId()] = serviceInstance //service实例的service id 必须唯一
	csr.serviceInstances[serviceInstance.GetServiceName()] = services
	csr.localServiceInstance = serviceInstance

	/*
		config := api.DefaultConfig()
		config.Address = r.Address
		client, err := api.NewClient(config)
		if err != nil {
			return err
		}
		agent := client.Agent()

		if r.IP == "" {
			r.IP = localIP()
		}

		checkHealthAddr := fmt.Sprintf("%v:%v", r.IP, r.Port) //fmt.Sprintf("%v:%v/%v", IP, r.Port, r.Service)
		fmt.Println("Health.Check地址", checkHealthAddr)
		reg := &api.AgentServiceRegistration{
			ID:      fmt.Sprintf("%v-%v-%v", r.Service, r.IP, r.Port), // 服务节点的名称
			Name:    fmt.Sprintf("%v", r.Service),                     // 服务名称
			Tags:    r.Tags,                                           // tag，可以为空
			Port:    r.Port,                                           // 服务端口
			Address: r.IP,                                             // 服务 IP
			Meta: map[string]string{
				"balanceFactor": strconv.Itoa(r.BalanceFactor),
			},
			Check: &api.AgentServiceCheck{ // 健康检查
				GRPC:                           checkHealthAddr, // grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
				GRPCUseTLS:                     false,
				Interval:                       r.Interval.String(), // 健康检查间隔
				Timeout:                        (time.Second * 10).String(),
				DeregisterCriticalServiceAfter: r.DeregisterCriticalServiceAfter.String(), // 注销时间，相当于过期时间
			},
		}

		if err := agent.ServiceRegister(reg); err != nil {
			return err
		}
	*/

	return nil
}

// deregister a service
func (c ConsulServiceRegistry) Deregister() {
	if c.serviceInstances == nil {
		return
	}

	services := c.serviceInstances[c.localServiceInstance.GetServiceName()]

	if services == nil {
		return
	}

	delete(services, c.localServiceInstance.GetInstanceId())

	if len(services) == 0 {
		delete(c.serviceInstances, c.localServiceInstance.GetServiceName())
	}

	_ = c.client.Agent().ServiceDeregister(c.localServiceInstance.GetInstanceId())

	c.localServiceInstance = nil
}

func (c ConsulServiceRegistry) GetServiceAddr() (ip string, port int, err error) {
	if c.serviceInstances == nil {
		return "", 0, errors.New("not instaance")
	}
	return c.localServiceInstance.GetHost(), c.localServiceInstance.GetPort(), nil
}

///////////////////////////////////////////////////////////////////////////////////////
// 服务发现
//

//返回指定服务的健康状态 List Service Instances for Service:   阻塞查询 Blocking Qurey
func (c ConsulServiceRegistry) HealthService(serviceName string, waitIndex uint64) ([]*api.ServiceEntry, *api.QueryMeta, error) {

	health := c.client.Health()
	/*
		healthcheck, _, err := health.Checks(serviceName, nil)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Println(healthcheck.AggregatedStatus())
	*/
	queryOptions := &api.QueryOptions{
		UseCache:  true,
		WaitIndex: waitIndex,
		WaitTime:  time.Minute * 5, //WaitTime默认为10分钟
	}

	return health.Service(serviceName, "", false, queryOptions)
}
