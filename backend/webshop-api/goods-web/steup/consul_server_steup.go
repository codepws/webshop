// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package steup

import (
	servicereg "common/ServiceReg"
	utils "common/Utils"
	"fmt"
	"strconv"
	"time"
	"webshop-api/user-web/common/global"
)

//
func SteupServiceReg() (*servicereg.ConsulServiceRegistry, error) {
	fmt.Println("======Consul user-web 服务注册======")

	// Consul 客户端实例
	// Consul 服务注册
	consul_host := global.AppConfig.ConsulConfig.Host
	consul_port := global.AppConfig.ConsulConfig.Port
	consul_token := ""
	register, err := servicereg.NewConsulServiceRegistry(consul_host, consul_port, consul_token)
	if err != nil {
		return nil, err
	}

	service_name := global.AppConfig.Name //"user-web"
	service_ip := global.AppConfig.Host
	service_port := global.AppConfig.Port
	if service_port <= 0 || service_port > 65535 {
		service_port, err = utils.GetFreePort() //动态获取空闲服务端口
		if err != nil {
			return nil, err
		}
	}

	service_tags := global.AppConfig.Tags // []string{"srv", "user"}
	//service_metadata := nil

	//服务实例ID
	service_id := fmt.Sprintf("%s-%s:%d", "user-srv", service_ip, service_port)

	serviceInstance, err := servicereg.NewDefaultServiceInstance(service_name, service_ip, service_port, false, service_tags, nil, true, service_id)
	if err != nil {
		return nil, err
	}

	// 将用户服务注册到Consul中心
	if err := register.Register(serviceInstance); err != nil {
		return nil, err
	}

	//健康检查异步处理
	go checkHealth(register)

	fmt.Println("Consul user-web 服务注册成功")

	return register, nil
}

var _userServiceUrls []string

func checkHealth(register *servicereg.ConsulServiceRegistry) {

	fmt.Println("======Consul  user-web 服务实例健康检测======")

	var waitIndex uint64 = 0
	for {
		// 阻塞查询 Blocking Qurey
		serviceEntry, queryMeta, err := register.HealthService(global.AppConfig.Name, waitIndex)
		if err != nil {
			fmt.Println("HealthService 错误：", err)
			time.Sleep(time.Second * 1)
			continue
		}

		//控制台打印一下获取服务列表的响应时间等信息  Mon Jan 2 15:04:05 -0700 MST 2006
		fmt.Printf("%s 获取服务[%s]：queryOptions.WaitIndex：%d  LastIndex：%d\n", time.Now().Format("2006-01-02 15:04:05"), global.AppConfig.Name, waitIndex, queryMeta.LastIndex)

		//版本号不一致 说明服务列表发生了变化
		if waitIndex != queryMeta.LastIndex {
			//
			waitIndex = queryMeta.LastIndex

			fmt.Println(queryMeta.CacheHit, queryMeta.CacheAge, queryMeta.LastIndex, queryMeta.RequestTime)
			var serviceUrls []string = make([]string, 0, len(serviceEntry))
			var addr string
			for idx, item := range serviceEntry {
				addr = item.Service.Address + ":" + strconv.Itoa(item.Service.Port)
				serviceUrls = append(serviceUrls, addr)

				fmt.Println(idx, item.Service.ID, item.Service.Service, item.Service.Address, item.Service.Port, item.Service.Namespace, item.Service.Datacenter, item.Service.CreateIndex, item.Service.ModifyIndex)

				fmt.Println(" ", item.Node.ID, item.Node.Node, item.Node.Address, item.Node.Datacenter, item.Node.CreateIndex, item.Node.ModifyIndex)
				fmt.Println(" ", item.Checks.AggregatedStatus())
			}

			//最新服务地址列表
			_userServiceUrls = serviceUrls //res.Response.Select(p => $"http://{p.Service.Address + ":" + p.Service.Port}").ToArray();

			fmt.Println("最新服务地址列表: ", _userServiceUrls)
		}

	}

}
