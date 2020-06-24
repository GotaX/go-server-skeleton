package redis

import (
	"fmt"

	driver "github.com/go-redis/redis/v8"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
)

var Option = cfg.Option{
	Name:     "Redis",
	OnCreate: newRedis,
}

func newRedis(source cfg.Scanner) (interface{}, error) {
	var c struct {
		Db       int    `json:"db"`
		Host     string `json:"host"`
		Password string `json:"password"`
		Port     string `json:"port"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	return driver.NewClient(&driver.Options{
		DB:       c.Db,
		Addr:     fmt.Sprintf("%v:%v", c.Host, c.Port),
		Password: c.Password,
	}), nil
}
