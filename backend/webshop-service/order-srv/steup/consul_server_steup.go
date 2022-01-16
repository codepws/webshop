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
	"webshop-service/order-srv/common/global"
)

func SteupServiceReg() (*servicereg.ConsulServiceRegistry, error) {
	fmt.Println("======Consul 订单服务注册======")

	//isDebug := utils.AppIsDebugMode()

	// GRPC 方法注册

	// Consul 服务注册
	consul_host := global.AppConfig.ConsulConfig.Host
	consul_port := global.AppConfig.ConsulConfig.Port
	consul_token := ""
	register, err := servicereg.NewConsulServiceRegistry(consul_host, consul_port, consul_token)
	if err != nil {
		return nil, err
	}

	service_name := global.AppConfig.Name //"order-srv"
	service_ip := global.AppConfig.Host
	service_port, err := utils.GetFreePort() //动态获取空闲服务端口
	if err != nil {
		return nil, err
	}
	service_tags := global.AppConfig.Tags // []string{"srv", "user"}
	//service_metadata := nil

	//服务实例ID
	service_id := fmt.Sprintf("%s-%s:%d", "order-srv", service_ip, service_port)

	serviceInstance, err := servicereg.NewDefaultServiceInstance(service_name, service_ip, service_port, false, service_tags, nil, false, service_id)
	if err != nil {
		return nil, err
	}

	// 将订单服务注册到Consul中心
	if err := register.Register(serviceInstance); err != nil {
		return nil, err
	}

	return register, nil
}
