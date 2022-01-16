package settings

type LogConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	Filename   string `mapstructure:"filename" json:"filename" yaml:"filename"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
}

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

/*
lettuce:
      pool:
        max-active: 1000 #连接池最大连接数（使用负值表示没有限制）
        max-idle: 10 # 连接池中的最大空闲连接
        min-idle: 5 # 连接池中的最小空闲连接
        max-wait: -1 # 连接池最大阻塞等待时间（使用负值表示没有限制）
*/
type Caches struct {
	MasterRedis   RedisConfig   `mapstructure:"master_redis" json:"master_redis" yaml:"master_redis"`       //主Redis（哨兵模式）
	SlaveRedis    RedisConfig   `mapstructure:"slave_redis" json:"slave_redis" yaml:"slave_redis"`          //从Redis（哨兵模式）
	RedisCluster  RedisCluster  `mapstructure:"redis_cluster" json:"redis_cluster" yaml:"redis_cluster"`    // 分布式Redis锁-集群模式
	RedisSentinel RedisSentinel `mapstructure:"redis_sentinel" json:"redis_sentinel" yaml:"redis_sentinel"` //Redis（哨兵模式）

	LockRedis []*RedisConfig `mapstructure:"lock_redis" json:"lock_redis" yaml:"lock_redis"` //分布式Redis锁-多台独立Redis
}

// Redis集群
type RedisCluster struct {
	MaxRedirects int            `mapstructure:"max_redirects" json:"max_redirects" yaml:"max_redirects"` //获取失败 最大重定向次数
	LockRedis    []*RedisConfig `mapstructure:"lock_redis" json:"lock_redis" yaml:"lock_redis"`          //分布式Redis锁-集群模式
}

// Redis哨兵
type RedisSentinel struct {
	MasterRedis RedisConfig   `mapstructure:"master_redis" json:"master_redis" yaml:"master_redis"` //主Redis
	SlaveRedis  RedisConfig   `mapstructure:"slave_redis" json:"slave_redis" yaml:"slave_redis"`    //从Redis
	Sentinels   []RedisConfig `mapstructure:"sentinels" json:"sentinels" yaml:"sentinels"`          //哨兵
}

type RedisConfig struct {
	Host         string `mapstructure:"host" json:"host" yaml:"host"`
	Port         int    `mapstructure:"port" json:"port" yaml:"port"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	DB           int    `mapstructure:"db" json:"db" yaml:"db"`
	PoolSize     int    `mapstructure:"pool_size" json:"pool_size" yaml:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns" json:"min_idle_conns" yaml:"min_idle_conns"`
}

//商品服务地址信息
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

type RocketMQConfig struct {
	PGroupName string `mapstructure:"p_group_name" json:"p_group_name" yaml:"p_group_name"` //生产者组
	CGroupName string `mapstructure:"c_group_name" json:"c_group_name" yaml:"c_group_name"` //消费者组
	Topic      string `mapstructure:"topic" json:"topic" yaml:"topic"`
	Tag        string `mapstructure:"tag" json:"tag" yaml:"tag"`
}

// RocketMQ配置
type RocketMQ struct {
	NameServers          []string        `mapstructure:"name_servers" json:"name_servers" yaml:"name_servers"`
	OrderCreateInvReback *RocketMQConfig `mapstructure:"order_create_inv_reback" json:"order_create_inv_reback" yaml:"order_create_inv_reback"` //订单创建失败库存归还（生产者）
	OrderPayInvReback    *RocketMQConfig `mapstructure:"order_pay_inv_reback" json:"order_pay_inv_reback" yaml:"order_pay_inv_reback"`          //订单支付超时库存归还（生产者）
	OrderInvReback       *RocketMQConfig `mapstructure:"order_inv_reback" json:"order_inv_reback" yaml:"order_inv_reback"`                      //订单库存归还（消费者）
	OrderPayTimeout      *RocketMQConfig `mapstructure:"order_pay_timeout" json:"order_pay_timeout" yaml:"order_pay_timeout"`                   //订单支付时间超时
}
