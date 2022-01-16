package config

//var AppConfig = new(AppConfig)

type AppConfig struct {
	Name       string `mapstructure:"name" json:"name" yaml:"name"`
	Mode       string `mapstructure:"mode" json:"mode" yaml:"mode"`
	Version    string `mapstructure:"version" json:"version" yaml:"version"`
	Host       string `mapstructure:"host" json:"host" yaml:"host"`
	Port       int    `mapstructure:"port" json:"port" yaml:"port"`
	*LogConfig `mapstructure:"log" json:"log" yaml:"log"`
	*DBs       `mapstructure:"dbs" json:"dbs" yaml:"dbs"`
	Caches     []*RedisConfig `mapstructure:"caches" json:"caches" yaml:"caches"`
	//Tags       []string       `mapstructure:"tags" json:"tags" yaml:"tags"` //Tags

	*JWTConfig    `mapstructure:"jwt" json:"jwt" yaml:"jwt"`          //JWT
	*AliSmsConfig `mapstructure:"alisms" json:"alisms" yaml:"alisms"` //阿里云短信服务
	*ConsulConfig `mapstructure:"consul" json:"consul" yaml:"consul"` //服务发现、配置管理中心服务
	*ServerConfig `mapstructure:"server" json:"server" yaml:"server"` //用户服务配置
	/*
		Host   string `yaml:"host"`
		User   string `yaml:"user"`
		Pwd    string `yaml:"pwd"`
		Dbname string `yaml:"dbname"`
	*/
}

// moduleConfig could be in a module specific package

type DBs struct {
	MasterDB DBConfig `mapstructure:"master_db" json:"master_db" yaml:"master_db"`
	SlaveDB  DBConfig `mapstructure:"slave_db" json:"slave_db" yaml:"slave_db"`
}

type DBConfig struct {
	Type         string `mapstructure:"type" json:"type" yaml:"type"`
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         int    `mapstructure:"port" json:"port" yaml:"port"`
	User         string `mapstructure:"user" json:"user" yaml:"user"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	Database     string `mapstructure:"database" json:"database" yaml:"database"`
	MaxOpenConns int    `mapstructure:"max_open_conns" json:"name" yaml:"name"`
	MaxIdleConns int    `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
}

type LogConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	Filename   string `mapstructure:"filename" json:"filename" yaml:"filename"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         int    `mapstructure:"port" json:"port" yaml:"port"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	DB           int    `mapstructure:"db" json:"db" yaml:"db"`
	PoolSize     int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" json:"min_idle_conns" yaml:"min_idle_conns"`
}

//用户服务地址信息
type ServerConfig struct {
	Host string   `mapstructure:"host" json:"host" yaml:"host"` //服务IP
	Port int      `mapstructure:"port" json:"port" yaml:"port"` //服务Port
	Name string   `mapstructure:"name" json:"name" yaml:"name"` //服务名称
	Tags []string `mapstructure:"tags" json:"tags" yaml:"tags"` //服务Tags
}

type JWTConfig struct {
	SigningKey string `mapstructure:"key" json:"key" yaml:"key"`
	Expire     int    `mapstructure:"expire" json:"expire" yaml:"expire"`
}

type AliSmsConfig struct {
	ApiKey     string `mapstructure:"key" json:"key" yaml:"key"`
	ApiSecrect string `mapstructure:"secrect" json:"secrect" yaml:"secrect"`
	Expire     int    `mapstructure:"expire" json:"name" yaml:"name"` //短信超时时间，单位：秒
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
	Port int    `mapstructure:"port" json:"port" yaml:"port"`
}

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
