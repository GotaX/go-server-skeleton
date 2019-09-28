package factory

import (
	"fmt"

	"github.com/sirupsen/logrus"
	driver "github.com/streadway/amqp"
)

var AMQP = Option{
	Name:      "AMQP",
	OnCreate:  newAmqp,
	OnCreated: logAmqpError,
	OnDestroy: func(v interface{}) { _ = v.(*driver.Connection).Close() },
}

func newAmqp(source Scanner) (interface{}, error) {
	var c struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		VHost    string `json:"vhost"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("amqp://%v:%v@%v:%v/%v",
		c.Username, c.Password, c.Host, c.Port, c.VHost)
	return driver.Dial(url)
}

func logAmqpError(name string, v interface{}) {
	conn := v.(*driver.Connection)
	ch := conn.NotifyClose(make(chan *driver.Error))
	go func() {
		for err := range ch {
			logrus.WithFields(logrus.Fields{
				"name":    "amqp-" + name,
				"code":    err.Code,
				"server":  err.Server,
				"recover": err.Recover,
			}).Warn(err.Reason)
		}
	}()
}
