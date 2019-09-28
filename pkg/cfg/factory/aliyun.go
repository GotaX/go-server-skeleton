package factory

import (
	driver "github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/auth/credentials"
)

var AliYun = Option{
	Name:      "AliYun",
	OnCreate:  newAliYun,
	OnDestroy: func(v interface{}) { v.(*driver.Client).Shutdown() },
}

func newAliYun(source Scanner) (interface{}, error) {
	var c struct {
		Region       string `json:"region"`
		AccessKey    string `json:"accessKey"`
		AccessSecret string `json:"accessSecret"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	cred := credentials.NewAccessKeyCredential(c.AccessKey, c.AccessSecret)
	return driver.NewClientWithOptions(c.Region, &driver.Config{}, cred)
}
