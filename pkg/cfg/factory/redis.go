package factory

import (
	"fmt"

	driver "github.com/go-redis/redis/v7"
)

var Redis = Option{
	Name:     "Redis",
	OnCreate: newRedis,
}

func newRedis(source Scanner) (interface{}, error) {
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
