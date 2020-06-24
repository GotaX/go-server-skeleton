package factory

import (
	"database/sql"
	"fmt"
	"math"

	driver "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/pkg/ext"
)

var MySQL = Option{
	Name:      "MySQL",
	OnCreate:  newMySQL,
	OnCreated: registerDBStats,
}

type mysqlLogger struct{}

func (l *mysqlLogger) Print(v ...interface{}) { logrus.Debug(v...) }

func newMySQL(source Scanner) (v interface{}, err error) {
	if err := driver.SetLogger(&mysqlLogger{}); err != nil {
		return nil, err
	}

	var c struct {
		Host     string   `json:"host"`
		Port     string   `json:"port"`
		Database string   `json:"database"`
		Username string   `json:"username"`
		Password string   `json:"password"`
		Params   []string `json:"params"`
		MaxOpen  int      `json:"maxOpen"`
		MaxIdle  int      `json:"maxIdle"`
		Tracing  bool     `json:"tracing"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	mu := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v",
		c.Username, c.Password, c.Host, c.Port, c.Database)
	if len(c.Params) > 0 {
		values := sliceToValues(c.Params, "=")
		mu += "?" + values.Encode()
	}

	name := "mysql"
	if c.Tracing {
		if name, err = ext.RegisterTracingDriver(name); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open(name, mu)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(c.MaxOpen)
	db.SetMaxIdleConns(int(math.Max(float64(c.MaxIdle), 1)))
	return db, nil
}