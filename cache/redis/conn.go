package redis

import (
	"github.com/gomodule/redigo/redis"
	"log"
	"time"
)

var (
	pool *redis.Pool
	redisHost = "127.0.0.1:6379"
	redisPassword = "123456"
)

// newRedisPool : 创建redis连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:         50,
		MaxActive:       30,
		IdleTimeout:     300 * time.Second,
		Dial: func() (redis.Conn, error) {
			// 1 打开连接
			c, err := redis.Dial("tcp", redisHost)
			if err != nil {
				log.Println(err.Error())
				return nil, err
			}

			// 2 访问认证
			if _, err = c.Do("AUTH", redisPassword); err != nil {
				c.Close()
				log.Println(err.Error())
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},

	}
}

func init() {
	pool = newRedisPool()
}

func Pool() *redis.Pool {
	return pool
}

