package alifc

import (
	driver "github.com/aliyun/fc-go-sdk"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
)

var Option = cfg.Option{
	Name:     "AliFc",
	OnCreate: newFc,
}

func newFc(source cfg.Scanner) (interface{}, error) {
	var c struct {
		Id       string `json:"id"`
		Secret   string `json:"secret"`
		Endpoint string `json:"endpoint"`
		Account  string `json:"account"`
		Region   string `json:"region"`
		Version  string `json:"version"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	return driver.NewClient(
		c.Endpoint, c.Version, c.Id, c.Secret,
		driver.WithRetryCount(10))
}
