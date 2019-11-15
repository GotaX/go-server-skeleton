package factory

import (
	"database/sql"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/GotaX/go-server-skeleton/pkg/ext"
)

const (
	EnvProfile     = "APP_PROFILE"
	DefaultProfile = "default"
)

type Scanner interface {
	Scan(val interface{}) error
}

type Option struct {
	Name      string
	OnCreate  func(source Scanner) (interface{}, error)
	OnCreated func(name string, v interface{})
	OnDestroy func(v interface{})
}

type ProviderMethod func(target interface{})

// Tool function

func IsDefaultEnv() bool {
	profile := os.Getenv(EnvProfile)
	return profile == "" || profile == DefaultProfile
}

func sliceToMap(slice []string, sep string) map[string]interface{} {
	m := make(map[string]interface{})
	for _, v := range slice {
		kv := strings.SplitN(v, sep, 2)
		m[kv[0]] = kv[1]
	}
	return m
}

func sliceToValues(slice []string, sep string) url.Values {
	m := url.Values{}
	for _, v := range slice {
		kv := strings.SplitN(v, sep, 2)
		m.Set(kv[0], kv[1])
	}
	return m
}

func registerDBStats(name string, v interface{}) {
	ext.RegisterDbStats(5*time.Second, v.(*sql.DB), name)
}
