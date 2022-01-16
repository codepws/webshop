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
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"
	"webshop-api/user-web/common/global"
	"webshop-api/user-web/config"
	"webshop-api/user-web/steup"

	"gopkg.in/yaml.v2"
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
	fmt.Println("======user-web======")
	//Chdir()

	isDebug := utils.AppIsDebugMode()

	appConfig := config.AppConfig{}
	// 用户服务 - 配置中心初始化
	configmgr.InitConfig("user-web", isDebug, &appConfig, onChange)
	// 全局变量
	global.AppConfig = &appConfig

	fmt.Println(appConfig)
	fmt.Println("LogConfig:", appConfig.LogConfig)
	fmt.Printf("Caches[%d]:%v\n", len(appConfig.Caches), appConfig.Caches)
	for i := 0; i < len(appConfig.Caches); i++ {
		fmt.Println("Caches:", *(appConfig.Caches[i]))
	}
	fmt.Println("UserSrvConfig:", *(appConfig.UserSrvConfig))
	fmt.Println("JWTConfig:", *(appConfig.JWTConfig))
	fmt.Println("AliSmsConfig:", *(appConfig.AliSmsConfig))
	fmt.Println("ConsulConfig:", *(appConfig.ConsulConfig))

	if err := steup.SteupLogger(global.AppConfig.LogConfig, isDebug); err != nil {
		panic(err)
	}

	if err := steup.SteupGrpcConnet(); err != nil {
		panic(err)
	}

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

	// Web 服务路由注册
	webRouter, err := steup.SteupWebRouter()
	if err != nil {
		panic(err)
	}

	//
	log.Printf("[%s] - [%s:%d] 服务启动成功!\n", global.AppConfig.Name, service_ip, service_port)
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

		err := webRouter.Run(fmt.Sprintf("%s:%d", service_ip, service_port))
		if err != nil {
			//logger.MainLogger.Fatal(fmt.Sprintf("run server failed, err:%v\n", err))
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

	/*
		userConn, err := grpc.Dial(
			fmt.Sprintf("consul://%s:%d/%s?wait=14s", "", 11111, ""),
			grpc.WithInsecure(),
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		)
		if err != nil {
			zap.S().Fatal("[InitSrvConn] 连接 【用户服务失败】")
		}
		fmt.Println(userConn)

		//Step 1:
		if err := SetupConfig(); err != nil {
			log.Fatalln(err.Error())
			return
		}
		defer logger.Close()

		//
		routers.SetupValidator()

		//Step 2:
		router := routers.SetupRouter()
		if err := router.Run(fmt.Sprintf(":%d", settings.WebConf.Port)); err != nil {
			zap.L().Fatal(fmt.Sprintf("run server failed, err:%v\n", err))
			return
		}

		////////////////////////////////////////////////////////////////////
		//// 优雅退出go守护进程 接收终止信号
		//创建监听退出chan
		quit := make(chan os.Signal, 1)

		//监听指定信号 ctrl+c kill
		//signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //, syscall.SIGUSR1, syscall.SIGUSR2)

		<-quit
		/*
			for s := range quit {
				switch s {
				case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
					fmt.Println("退出", s)
					//ExitFunc()
				//case syscall.SIGUSR1:
				//	fmt.Println("usr1", s)
				//case syscall.SIGUSR2:
				//	fmt.Println("usr2", s)
				default:
					fmt.Println("other 信号", s)
				}
			}
	*/

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
