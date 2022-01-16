package config

import (
	settings "common/settings"
)

//var AppConfig = new(AppConfig)

type AppConfig struct {
	//Mode    string `mapstructure:"mode" json:"mode" yaml:"mode"`
	Version string   `mapstructure:"version" json:"version" yaml:"version"`
	Name    string   `mapstructure:"name" json:"name" yaml:"name"` //当前服务名称
	Host    string   `mapstructure:"host" json:"host" yaml:"host"` //当前服务Host
	Port    int      `mapstructure:"port" json:"port" yaml:"port"` //当前服务端口号
	Tags    []string `mapstructure:"tags" json:"tags" yaml:"tags"` //当前服务端Tags

	*settings.LogConfig `mapstructure:"log" json:"log" yaml:"log"`
	*settings.DBs       `mapstructure:"dbs" json:"dbs" yaml:"dbs"`
	settings.Caches     `mapstructure:"caches" json:"caches" yaml:"caches"`

	*settings.JWTConfig    `mapstructure:"jwt" json:"jwt" yaml:"jwt"`                //JWT
	*settings.AliSmsConfig `mapstructure:"alisms" json:"alisms" yaml:"alisms"`       //阿里云短信服务
	*settings.ConsulConfig `mapstructure:"consul" json:"consul" yaml:"consul"`       //服务发现、配置管理中心服务
	*settings.RocketMQ     `mapstructure:"rocketmq" json:"rocketmq" yaml:"rocketmq"` //RocketMQ配置

	GoodsSrv     *settings.ServerConfig `mapstructure:"goods_srv" json:"goods_srv" yaml:"goods_srv"`             //商品服务配置
	InventorySrv *settings.ServerConfig `mapstructure:"inventory_srv" json:"inventory_srv" yaml:"inventory_srv"` //库存服务配置
}

// moduleConfig could be in a module specific package

/*
type RedisConfig struct {
	Host   string `mapstructure:"host" json:"host"`
	Port   int    `mapstructure:"port" json:"port"`
	Expire int    `mapstructure:"expire" json:"expire"`
}
*/
/*
type ServerConfig struct {
	Name        string        `mapstructure:"name" json:"name"`
	Host        string        `mapstructure:"host" json:"host"`
	Tags        []string      `mapstructure:"tags" json:"tags"`
	Port        int           `mapstructure:"port" json:"port"`
	UserSrvInfo UserSrvConfig `mapstructure:"user_srv" json:"user_srv"`
	JWTInfo     JWTConfig     `mapstructure:"jwt" json:"jwt"`
	AliSmsInfo  AliSmsConfig  `mapstructure:"sms" json:"sms"`
	RedisInfo   RedisConfig   `mapstructure:"redis" json:"redis"`
	ConsulInfo  ConsulConfig  `mapstructure:"consul" json:"consul"`
}
*/
