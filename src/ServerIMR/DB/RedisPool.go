package DB

import (
	"ServerIMR/config"
	"time"

	"github.com/gomodule/redigo/redis"
)

var Redis_pool *redis.Pool

func InitRedis() {
	Redis_pool = NewRedisPool(config.Config.Redis_address, config.Config.Redis_password,
		config.Config.Redis_db)
}

func NewRedisPool(server, password string, db int) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     100,
		MaxActive:   500,
		IdleTimeout: 480 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialTimeout("tcp", server,
				time.Duration(2)*time.Second, 0, 0)
			if err != nil {
				return nil, err
			}
			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if db > 0 && db < 16 {
				if _, err := c.Do("SELECT", db); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
	}
}
