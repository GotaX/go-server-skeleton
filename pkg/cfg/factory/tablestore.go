package factory

import (
	driver "github.com/aliyun/aliyun-tablestore-go-sdk/tablestore"
)

var OTS = Option{
	Name:     "OTS",
	OnCreate: newOts,
}

func newOts(source Scanner) (interface{}, error) {
	var c struct {
		Id       string `json:"id"`
		Secret   string `json:"secret"`
		Endpoint string `json:"endpoint"`
		Instance string `json:"instance"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}
	client := driver.NewClient(c.Endpoint, c.Instance, c.Id, c.Secret)
	return client, nil
}
