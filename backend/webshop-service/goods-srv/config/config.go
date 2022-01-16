package config

import (
	settings "common/settings"
)

//var AppConfig = new(AppConfig)

type AppConfig struct {
	//Mode                string   `mapstructure:"mode" json:"mode" yaml:"mode"`
	Version string   `mapstructure:"version" json:"version" yaml:"version"`
	Host    string   `mapstructure:"host" json:"host" yaml:"host"`
	Port    int      `mapstructure:"port" json:"port" yaml:"port"`
	Name    string   `mapstructure:"name" json:"name" yaml:"name"`
	Tags    []string `mapstructure:"tags" json:"tags" yaml:"tags"` //Tags

	*settings.LogConfig `mapstructure:"log" json:"log" yaml:"log"`
	*settings.DBs       `mapstructure:"dbs" json:"dbs" yaml:"dbs"`
	settings.Caches     `mapstructure:"caches" json:"caches" yaml:"caches"`

	*settings.JWTConfig    `mapstructure:"jwt" json:"jwt" yaml:"jwt"`          //JWT
	*settings.AliSmsConfig `mapstructure:"alisms" json:"alisms" yaml:"alisms"` //阿里云短信服务
	*settings.ConsulConfig `mapstructure:"consul" json:"consul" yaml:"consul"` //服务发现、配置管理中心服务
	/*
		Host   string `yaml:"host"`
		User   string `yaml:"user"`
		Pwd    string `yaml:"pwd"`
		Dbname string `yaml:"dbname"`
	*/
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
