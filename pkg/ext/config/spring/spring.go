package spring

import (
	"fmt"
	"strings"

	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/source"

	"skeleton/pkg/ext/config/url"
)

const (
	kConfProfile, dConfProfile = "profile", "default"
	kConfUrl, dConfUrl         = "conf_host", "http://localhost:8080"
	kConfLabel, dConfLabel     = "conf_label", "go"
	kConfName, dConfName       = "conf_name", ""
)

func NewSource(opts ...source.Option) source.Source {
	options := source.NewOptions(opts...)
	address := fmt.Sprintf("%v/%v/%v-%v.json",
		getValue(options, kConfUrl, dConfUrl),
		getValue(options, kConfLabel, dConfLabel),
		getValue(options, kConfName, dConfName),
		getValue(options, kConfProfile, dConfProfile))
	fmt.Println("Config URL: ", address)
	return url.NewSource(url.WithURL(address))
}

func getValue(opts source.Options, key string, value string) string {
	if v, ok := opts.Context.Value(key).(string); ok {
		return v
	}
	return config.Get(strings.Split(key, "_")...).String(value)
}
