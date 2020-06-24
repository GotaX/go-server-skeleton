package factory

import (
	"fmt"

	v7 "github.com/go-redis/redis/v7"
	v8 "github.com/go-redis/redis/v8"
)

const (
	_ = iota
	redisV7
	redisV8
)

var (
	Redis = Option{
		Name:     "Redis",
		OnCreate: newRedis(redisV7),
	}
	RedisV8 = Option{
		Name:     "Redis",
		OnCreate: newRedis(redisV8),
	}
)

func newRedis(version int) func(source Scanner) (interface{}, error) {
	return func(source Scanner) (obj interface{}, err error) {
		var c struct {
			Db       int    `json:"db"`
			Host     string `json:"host"`
			Password string `json:"password"`
			Port     string `json:"port"`
		}
		if err := source.Scan(&c); err != nil {
			return nil, err
		}

		switch version {
		case redisV7:
			obj = v7.NewClient(&v7.Options{
				DB:       c.Db,
				Addr:     fmt.Sprintf("%v:%v", c.Host, c.Port),
				Password: c.Password,
			})
		case redisV8:
			obj = v8.NewClient(&v8.Options{
				DB:       c.Db,
				Addr:     fmt.Sprintf("%v:%v", c.Host, c.Port),
				Password: c.Password,
			})
		}
		return
	}
}
