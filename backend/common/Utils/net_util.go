package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
)

// 当前APP是否为开发模式
func AppIsDebugMode() bool {

	fmt.Println("WEBSHOP_DEBUG:", os.Getenv("WEBSHOP_DEBUG"))
	//如果是本地开发环境端口号固定，线上环境启动获取端口号
	debug := os.Getenv("WEBSHOP_DEBUG") //viper.GetBool("WEBSHOP_DEBUG")
	if debug != "" {
		fmt.Println("当前为开发模式")
	} else {
		fmt.Println("当前为生产模式")
	}

	return debug != ""
}

// 动态获取一个空闲可用的端口号
func GetFreePort() (int, error) {

	//ResolveTCPAddr用于获取一个TCPAddr
	//net参数是"tcp4"、"tcp6"、"tcp"
	//addr表示域名或IP地址加端口号
	tcpaddr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	//监听端口
	tcplisten, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return 0, err
	}
	defer tcplisten.Close()

	return tcplisten.Addr().(*net.TCPAddr).Port, nil
}

// 获取多个空闲可用端口
func GetFreePorts(count int) ([]int, error) {
	var ports []int
	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		defer l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
}

func FindFirstNonLoopbackIP() (ipv4 string, err error) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			adders, _ := netInterfaces[i].Addrs()

			for _, address := range adders {
				if inet, ok := address.(*net.IPNet); ok && !inet.IP.IsLoopback() {
					fmt.Println(inet)
					if inet.IP.To4() != nil {
						return inet.IP.String(), nil
					}
				}
			}
		}
	}

	return "", errors.New("no find")
}

//如果是服务主机多网卡的情况，默认获取第一个地址，所以有可能返回的是局域网IP，而不是外网IP
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "unknow"
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "unknow"
}
