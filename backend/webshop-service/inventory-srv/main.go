// 包注释
// @Title  请填写文件名称（需要改）
// @Description  请填写文件描述（需要改）
// @Author  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
// @Update  请填写自己的真是姓名（需要改）  ${DATE} ${TIME}
package main

import (
	configmgr "common/ConfigMgr"
	utils "common/Utils"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"
	"webshop-service/inventory-srv/common/global"
	"webshop-service/inventory-srv/config"
	"webshop-service/inventory-srv/steup"

	"gopkg.in/yaml.v3"
	//"github.com/bronze1man/yaml2json/y2jLib"
)

func onChange(namespace, group, dataId, data string) {
	fmt.Println("config changed group:" + group + ", dataId:" + dataId + ", content:\n" + data)

	///////////////////////////////////////////////////////////////////////////
	//fmt.Println(content) //字符串 - yaml
	//想要将一个json字符串转换成struct，需要去设置这个struct的tag
	// json
	appConfig := config.AppConfig{}
	err := yaml.Unmarshal([]byte(data), &appConfig)
	if err != nil {
		//zap.S().Fatalf("读取nacos配置失败： %s", err.Error())
		fmt.Println("读取nacos配置失败：", err)
	}
	global.AppConfig = &appConfig
}

// 函数（方法）注释
// @title    函数名称
// @description   函数的详细描述
// @auth      作者             时间（2019/6/18   10:57 ）
// @param     输入参数名        参数类型         "解释"
// @return    返回参数名        参数类型         "解释"
func main() {
	fmt.Println("======inventory-srv======")

	//Chdir()
	//
	isDebug := utils.AppIsDebugMode()

	appConfig := config.AppConfig{}
	// 用户服务 - 配置中心初始化
	configmgr.InitConfig("inventory-srv", isDebug, &appConfig, onChange)

	fmt.Println(appConfig)
	fmt.Println("LogConfig:", appConfig.LogConfig)
	fmt.Println("DBs:", appConfig.DBs)
	fmt.Printf("Caches.MasterRedis:%v\n", appConfig.Caches.MasterRedis)
	fmt.Printf("Caches.SlaveRedis:%v\n", appConfig.Caches.SlaveRedis)
	fmt.Printf("Caches.RedisCluster.MaxRedirects:%v\n", appConfig.Caches.RedisCluster)

	fmt.Printf("Caches.LockRedis 个数:%v\n", len(appConfig.Caches.LockRedis))
	for i := 0; i < len(appConfig.Caches.LockRedis); i++ {
		fmt.Printf("LockRedis[%v]: %v\n", i, *(appConfig.Caches.LockRedis[i]))
	}

	fmt.Println("RocketMQ.NameServers:", appConfig.RocketMQ.NameServers)
	fmt.Println("RocketMQ.OrderInvReback:", *(appConfig.RocketMQ.OrderInvReback))

	fmt.Println("JWTConfig:", *(appConfig.JWTConfig))
	fmt.Println("AliSmsConfig:", *(appConfig.AliSmsConfig))
	fmt.Println("ConsulConfig:", *(appConfig.ConsulConfig))

	// 全局变量
	global.AppConfig = &appConfig

	//数据库初始化
	if err := global.DBMgr.Init(appConfig.DBs); err != nil {
		panic(err)
	}

	//RocketMQ
	if err := steup.SteupRocketMQReg(appConfig.RocketMQ); err != nil {
		panic(err)
	}

	// GRPC 方法注册
	grpcServer, err := steup.SteupGrpcReg()
	if err != nil {
		panic(err)
	}
	defer grpcServer.Stop()

	// Consul 服务注册，将用户服务注册到Consul中心
	consulSegister, err := steup.SteupServiceReg()
	if err != nil {
		panic(err)
	}
	defer consulSegister.Deregister()

	////////////////////////////////////////////////////////
	service_ip, service_port, err := consulSegister.GetServiceAddr()
	if err != nil {
		panic(err)
	}

	//监听网络并绑定到grpc
	list, err := net.Listen("tcp", fmt.Sprintf("%v:%v", service_ip, service_port)) // 169.254.116.41  0.0.0.0
	if err != nil {
		log.Fatal("net listen 失败")
	}
	//defer list.Close()

	////////////////////////////////////////////////////////////////////
	//// 优雅退出go守护进程 接收终止信号
	//创建监听退出chan
	quit := make(chan os.Signal, 1)
	//监听指定信号 ctrl+c kill
	//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //, syscall.SIGUSR1, syscall.SIGUSR2)

	//
	go func() {
		log.Println("Grpc 监听启动")
		if err := grpcServer.Serve(list); err != nil {
			//log.Fatal("grpc serve 失败: ", err)
			log.Println("GRPC Serve 失败：", err)

			time.Sleep(time.Second * 2)
			log.Println("发送退出信号")
			quit <- syscall.SIGQUIT
		}

	}()
	//
	time.Sleep(time.Second * 1)
	//
	log.Printf("[%s] - [%s:%d] 服务启动成功!\n", global.AppConfig.Name, service_ip, service_port)
	/////////////////////////////////////
	//等待退出信号
	<-quit

	log.Println("Shutdown User Server ...")

	//
	time.Sleep(time.Second * 1)
	// 5秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过5秒就超时退出
	//if err := list.Close(); err != nil {
	//	log.Fatal("Server Shutdown error: ", err)
	//}

	log.Println("User Server exiting")

}

// chdir 将程序工作路径修改成程序所在位置
func Chdir() (err error) {
	curUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("用户目录", curUser.HomeDir)

	//dir, err := filepath.Abs("views/list.html") //filepath.Dir(os.Args[0])
	//dir, err := exec.LookPath(os.Args[0])

	olddir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("当前工作目录", olddir)

	dir, _ := os.Executable()
	exPath := filepath.Dir(dir)
	println(exPath)

	/*
		newdir, err := filepath.Abs("src")
		if err != nil {
			log.Fatal(err)
		}

		err = os.Chdir(newdir)
		if err != nil {
			log.Fatal(err)
		}
		// or like this
		// cmd := exec.Command("cd", curUser.HomeDir)
		dir, err := os.Getwd() // ok in application
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("new工作目录", dir)
	*/
	return
}
