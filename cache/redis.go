package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rickslab/ares/config"
)

var (
	redisClients = map[string]*RedisClient{}
	redisMutex   = sync.RWMutex{}
)

type RedisClient struct {
	*redis.Pool
}

func NewRedisClient(addr, password, db string) *RedisClient {
	return &RedisClient{
		Pool: &redis.Pool{
			MaxIdle:     10,
			IdleTimeout: 240 * time.Second,
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			Dial: func() (redis.Conn, error) {
				return dial("tcp", addr, password, db)
			},
		},
	}
}

func (cli *RedisClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	conn := cli.Pool.Get()
	defer conn.Close()

	return conn.Do(cmd, args...)
}

func dial(network, address, password, db string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	if db != "" {
		if _, err := c.Do("SELECT", db); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func Redis(name string) *RedisClient {
	cli := getRedisCli(name)
	if cli != nil {
		return cli
	}
	return initRedisCli(name)
}

func SetRedis(name string, cli *RedisClient) {
	redisClients[name] = cli
}

func initRedisCli(name string) *RedisClient {
	redisMutex.Lock()
	defer redisMutex.Unlock()

	cli := redisClients[name]
	if cli != nil {
		return cli
	}

	conf := config.YamlEnv().Sub(fmt.Sprintf("redis.%s", name))
	if conf == nil {
		return nil
	}

	cli = NewRedisClient(conf.GetString("address"), conf.GetString("auth"), conf.GetString("db"))
	SetRedis(name, cli)
	return cli
}

func getRedisCli(name string) *RedisClient {
	redisMutex.RLock()
	defer redisMutex.RUnlock()

	return redisClients[name]
}
