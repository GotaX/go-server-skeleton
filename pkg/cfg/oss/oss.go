package oss

import (
	"net/http"

	driver "github.com/aliyun/aliyun-oss-go-sdk/oss"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
)

var Option = cfg.Option{
	Name:     "OSS",
	OnCreate: newOss,
}

func newOss(source cfg.Scanner) (interface{}, error) {
	var c struct {
		Id       string `json:"id"`
		Secret   string `json:"secret"`
		Endpoint string `json:"endpoint"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}
	return driver.New(c.Endpoint, c.Id, c.Secret,
		driver.HTTPClient(http.DefaultClient))
}
