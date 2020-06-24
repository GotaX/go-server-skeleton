package postgres

import (
	"database/sql"
	"fmt"
	"math"

	_ "github.com/lib/pq"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
	"github.com/GotaX/go-server-skeleton/pkg/cfg/rds"
)

var Option = cfg.Option{
	Name:      "Postgres",
	OnCreate:  newPostgres,
	OnCreated: cfg.RegisterDBStats,
}

func newPostgres(source cfg.Scanner) (v interface{}, err error) {
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

	mu := fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
		c.Username, c.Password, c.Host, c.Port, c.Database)
	if len(c.Params) > 0 {
		values := cfg.SliceToValues(c.Params, "=")
		mu += "?" + values.Encode()
	}

	name := "postgres"
	if c.Tracing {
		if name, err = rds.RegisterTracingDriver(name); err != nil {
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
