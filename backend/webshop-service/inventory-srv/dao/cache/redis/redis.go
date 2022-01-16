package redis

import (
	"fmt"
	"webshop-service/user-srv/config"

	"github.com/go-redis/redis"
)

var (
	client *redis.Client
	Nil    = redis.Nil
)

type SliceCmd = redis.SliceCmd
type StringStringMapCmd = redis.StringStringMapCmd

func Init(cacheCfg []*config.RedisConfig) error {
	fmt.Println("初始化Redis")
	var (
		redis_master *config.RedisConfig = cacheCfg[0]
		//redis_slave_1 *settings.RedisConfig = cacheCfg[1]
		//redis_slave_2 *settings.RedisConfig = cacheCfg[1]
	)

	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redis_master.Host, redis_master.Port),
		Password:     redis_master.Password, // no password set
		DB:           redis_master.DB,       // use default DB
		PoolSize:     redis_master.PoolSize,
		MinIdleConns: redis_master.MinIdleConns,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	return nil
}

// Close 关闭Redis连接
func Close() {
	fmt.Println("关闭Redis")
	_ = client.Close()
}
