package proxy

import (
	"net"
	"time"

	mysql0 "github.com/go-sql-driver/mysql"
	driver "golang.org/x/net/proxy"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
)

var Option = cfg.Option{
	Name:     "Proxy",
	OnCreate: newProxy,
	OnCreated: func(name string, v interface{}) {
		dialer := v.(driver.Dialer)
		mysql0.RegisterDial("socks5",
			func(addr string) (conn net.Conn, e error) { return dialer.Dial("tcp", addr) })
	},
}

func newProxy(source cfg.Scanner) (interface{}, error) {
	var c struct {
		Protocol string `json:"protocol"`
		Address  string `json:"address"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}
	return driver.SOCKS5(c.Protocol, c.Address,
		&driver.Auth{
			User:     c.Username,
			Password: c.Password,
		},
		&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		},
	)
}
