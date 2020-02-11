package rds

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/trace"

	"github.com/GotaX/go-server-skeleton/pkg/ext/app"
)

var (
	dbUp = NewGaugeVec(GaugeOpts{
		Name: "db_up",
		Help: "数据库运行状态",
	}, []string{"name"})
	dbIdle = NewGaugeVec(GaugeOpts{
		Name: "db_idle",
		Help: "数据库闲置连接数",
	}, []string{"name"})
	dbInUse = NewGaugeVec(GaugeOpts{
		Name: "db_in_use",
		Help: "数据库活跃连接数",
	}, []string{"name"})
	dbOpenConnections = NewGaugeVec(GaugeOpts{
		Name: "db_open_connections",
		Help: "数据库开启的连接总数",
	}, []string{"name"})
	dbWaitCount = NewGaugeVec(GaugeOpts{
		Name: "db_wait_count",
		Help: "数据库等待中连接数",
	}, []string{"name"})
	dbWaitDuration = NewGaugeVec(GaugeOpts{
		Name: "db_wait_duration",
		Help: "数据库累计等待连接时间",
	}, []string{"name"})

	registerDBMetrics = &sync.Once{}
)

func RegisterDbStats(interval time.Duration, db *sql.DB, name string) {
	registerDBMetrics.Do(func() {
		MustRegister(dbUp, dbIdle, dbInUse, dbOpenConnections, dbWaitCount, dbWaitDuration)
	})

	name = regexp.MustCompile(`\w+ \((\w+)\)`).FindStringSubmatch(name)[1]
	labels := Labels{"name": name}
	app.RunTicker(fmt.Sprintf("DB stats checker %s", name), interval, func() {
		stats := db.Stats()
		dbIdle.With(labels).Set(float64(stats.Idle))
		dbInUse.With(labels).Set(float64(stats.InUse))
		dbOpenConnections.With(labels).Set(float64(stats.OpenConnections))
		dbWaitCount.With(labels).Set(float64(stats.WaitCount))
		dbWaitDuration.With(labels).Set(float64(stats.WaitDuration))

		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		if err := db.PingContext(ctx); err != nil {
			dbUp.With(labels).Set(0)
		} else {
			dbUp.With(labels).Set(1)
		}
	})
}

func RegisterTracingDriver(driverName string) (string, error) {
	// retrieve the driver implementation we need to wrap with instrumentation
	db, err := sql.Open(driverName, "")
	if err != nil {
		return "", err
	}
	dri := db.Driver()
	if err = db.Close(); err != nil {
		return "", err
	}

	// Since we might want to register multiple ocsql drivers to have different
	// TraceOptions, but potentially the same underlying database driver, we
	// cycle through to find available driver names.
	driverName = driverName + "-mbxc-"
	for i := int64(0); i < 100; i++ {
		var (
			found   = false
			regName = driverName + strconv.FormatInt(i, 10)
		)
		for _, name := range sql.Drivers() {
			if name == regName {
				found = true
			}
		}
		if !found {
			sql.Register(regName, wDriver{dri})
			return regName, nil
		}
	}
	return "", errors.New("unable to register driver, all slots have been taken")
}

type wDriver struct{ driver.Driver }

func (w wDriver) Open(name string) (driver.Conn, error) {
	conn, err := w.Driver.Open(name)
	return &wConn{conn}, err
}

type wConn struct{ driver.Conn }

func (w wConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if v, ok := w.Conn.(driver.ConnBeginTx); ok {
		return v.BeginTx(ctx, opts)
	}
	return nil, errors.New("conn not driver.ConnBeginTx")
}

func (w wConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (row driver.Rows, err error) {
	ctx, span := startSpan(ctx, "sql:query", query, args)
	defer func() {
		setSpanStatus(span, err)
		span.End()
	}()
	return w.Conn.(driver.QueryerContext).QueryContext(ctx, query, args)
}

func (w wConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (r driver.Result, err error) {
	ctx, span := startSpan(ctx, "sql:exec", query, args)
	defer func() {
		setSpanStatus(span, err)
		span.End()
	}()
	return w.Conn.(driver.ExecerContext).ExecContext(ctx, query, args)
}

func startSpan(ctx context.Context, name string, query string, args []driver.NamedValue) (context.Context, *trace.Span) {
	ctx, span := trace.StartSpan(ctx, name)
	span.AddAttributes(attributes(query, args)...)
	return ctx, span
}

func attributes(query string, args []driver.NamedValue) []trace.Attribute {
	return []trace.Attribute{
		trace.StringAttribute("db.statement", query),
		trace.StringAttribute("db.params", argsToString(args)),
	}
}

func argsToString(args []driver.NamedValue) string {
	if size := len(args); size > 0 {
		values := make([]string, len(args))
		for i, nv := range args {
			switch v := nv.Value.(type) {
			case int64:
				values[i] = strconv.FormatInt(v, 10)
			case float64:
				values[i] = strconv.FormatFloat(v, 'f', -1, 64)
			case bool:
				values[i] = strconv.FormatBool(v)
			case []byte:
				values[i] = string(v)
			case string:
				values[i] = v
			case time.Time:
				values[i] = v.Format(time.RFC3339)
			case nil:
				values[i] = "NULL"
			default:
				values[i] = fmt.Sprintf("%v", v)
			}
		}
		return strings.Join(values, ", ")
	}
	return ""
}

func setSpanStatus(span *trace.Span, err error) {
	var status trace.Status
	switch err {
	case nil:
		status.Code = trace.StatusCodeOK
		span.SetStatus(status)
		return
	case context.Canceled:
		status.Code = trace.StatusCodeCancelled
	case context.DeadlineExceeded:
		status.Code = trace.StatusCodeDeadlineExceeded
	case sql.ErrNoRows:
		status.Code = trace.StatusCodeNotFound
	case sql.ErrTxDone:
		status.Code = trace.StatusCodeFailedPrecondition
	default:
		status.Code = trace.StatusCodeUnknown
	}
	status.Message = err.Error()
	span.SetStatus(status)
}
