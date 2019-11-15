package factory

import (
	"database/sql"
	"fmt"
	"math"

	_ "github.com/lib/pq"

	"github.com/GotaX/go-server-skeleton/pkg/ext"
)

var Postgres = Option{
	Name:      "Postgres",
	OnCreate:  newPostgres,
	OnCreated: registerDBStats,
}

func newPostgres(source Scanner) (interface{}, error) {
	var c struct {
		Host     string   `json:"host"`
		Port     string   `json:"port"`
		Database string   `json:"database"`
		Username string   `json:"username"`
		Password string   `json:"password"`
		Params   []string `json:"params"`
		MaxOpen  int      `json:"maxOpen"`
		MaxIdle  int      `json:"maxIdle"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	mu := fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
		c.Username, c.Password, c.Host, c.Port, c.Database)
	if len(c.Params) > 0 {
		values := sliceToValues(c.Params, "=")
		mu += "?" + values.Encode()
	}

	name, err := ext.RegisterTracingDriver("postgres")
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(name, mu)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(c.MaxOpen)
	db.SetMaxIdleConns(int(math.Max(float64(c.MaxIdle), 1)))
	return db, nil
}
