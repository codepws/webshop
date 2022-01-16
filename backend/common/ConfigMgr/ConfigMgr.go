package configmgr

import (
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type nacosConfig struct {
	Host      string `mapstructure:"host"`
	Port      uint64 `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	User      string `mapstructure:"user"`
	Password  string `mapstructure:"password"`
	DataId    string `mapstructure:"dataid"`
	Group     string `mapstructure:"group"`
}

// Nacos配置初始化
func InitConfig(dirname string, isDebug bool, appConfig interface{} /*out*/, onChange func(namespace, group, dataId, data string) /*out*/) {

	configFilePrefix := "config"
	configFileName := fmt.Sprintf("%s/%s-pro.yaml", dirname, configFilePrefix)
	if isDebug {
		configFileName = fmt.Sprintf("%s/%s-debug.yaml", dirname, configFilePrefix)
	}

	v := viper.New()
	//文件的路径如何设置
	v.SetConfigFile(configFileName)
	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}
	//这个对象如何在其他文件中使用 - 全局变量
	var ConfigMgr = &nacosConfig{}
	if err := v.Unmarshal(ConfigMgr); err != nil {
		panic(err)
	}
	//zap.S().Infof("配置信息: &v", ConfigMgr)

	///////////////////////////////////////////////////////
	//从nacos中读取配置信息
	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: ConfigMgr.Host,
			Port:   ConfigMgr.Port,
		},
	}
	// 创建clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         ConfigMgr.Namespace, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              fmt.Sprintf("tmp/%s/nacos/log", dirname),
		CacheDir:            fmt.Sprintf("tmp/%s/nacos/cache", dirname),
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	/*
		// 创建动态配置客户端
		configClient, err := clients.CreateConfigClient(map[string]interface{}{
			"serverConfigs": serverConfigs,
			"clientConfig":  clientConfig,
		})
		if err != nil {
			panic(err)
		}
	*/

	// 创建动态配置客户端的另一种方式 (推荐)
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		panic(err)
	}

	// 获取配置
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: ConfigMgr.DataId,
		Group:  ConfigMgr.Group})

	if err != nil {
		panic(err)
	}

	// 监听配置变化
	//go func() {
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId:   ConfigMgr.DataId,
		Group:    ConfigMgr.Group,
		OnChange: onChange, /*func(namespace, group, dataId, data string) {
			fmt.Println("config changed group:" + group + ", dataId:" + dataId + ", content:\n" + data)
		}*/
	})
	if err != nil {
		panic(err)
	}
	//}()

	fmt.Println(content)

	///////////////////////////////////////////////////////////////////////////
	//fmt.Println(content) //字符串 - yaml
	//想要将一个json字符串转换成struct，需要去设置这个struct的tag
	// json
	err = yaml.Unmarshal([]byte(content), appConfig)
	if err != nil {
		//zap.S().Fatalf("读取nacos配置失败： %s", err.Error())
		fmt.Println("读取nacos配置失败：", err)
	}

}
