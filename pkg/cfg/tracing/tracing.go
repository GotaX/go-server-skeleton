package tracing

import (
	"net/http"

	"contrib.go.opencensus.io/exporter/jaeger"
	"contrib.go.opencensus.io/exporter/zipkin"
	. "github.com/openzipkin/zipkin-go"
	. "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
	"github.com/GotaX/go-server-skeleton/pkg/ext/app"
	"github.com/GotaX/go-server-skeleton/pkg/ext/tracing"
)

var Option = cfg.Option{
	Name:     "Tracing",
	OnCreate: newTracing,
}

func newTracing(source cfg.Scanner) (interface{}, error) {
	var c struct {
		Enable      bool   `json:"enable"`
		ServiceName string `json:"serviceName"`
		Endpoint    string `json:"endpoint"`
		Type        string `json:"type"`
		Jaeger      string `json:"jaeger"`
	}
	if err := source.Scan(&c); err != nil {
		return nil, err
	}

	if !c.Enable {
		return nil, nil
	}

	var (
		exporter trace.Exporter
		err      error
	)
	if !cfg.IsDefaultEnv() {
		switch c.Type {
		case "jaeger":
			exporter, err = newJaegerExporter(c.ServiceName, c.Jaeger)
			tracing.Propagation = &tracing.JaegerFormat{}
		case "zipkin":
			exporter, _, err = newZipkinExporter(c.ServiceName, c.Endpoint)
			tracing.Propagation = &tracecontext.HTTPFormat{}
		}
		if err != nil {
			logrus.WithError(err).Warn("Fail to init tracing")
		}
	}

	if exporter == nil {
		exporter = &PrintExporter{logger: logrus.StandardLogger()}
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	http.DefaultClient.Transport = &ochttp.Transport{
		Propagation:    tracing.Propagation,
		FormatSpanName: formatSpanName,
	}
	return exporter, nil
}

func formatSpanName(request *http.Request) string {
	if name := request.URL.Path; name != "" {
		return name
	}
	return "/"
}

func newZipkinExporter(serviceName, endpoint string) (trace.Exporter, func(), error) {
	onError := func(err error) (trace.Exporter, func(), error) {
		return nil, nil, xerrors.Errorf("while newZipkinExporter: %w", err)
	}
	ip, err := app.HostIP()
	if err != nil {
		return onError(err)
	}
	host := ip.String() + ":8080"

	localEndpoint, err := NewEndpoint(serviceName, host)
	if err != nil {
		return onError(err)
	}
	reporter := NewReporter(endpoint)
	exporter := zipkin.NewExporter(reporter, localEndpoint)
	return exporter, func() { _ = reporter.Close() }, nil
}

func newJaegerExporter(serviceName, endpoint string) (trace.Exporter, error) {
	return jaeger.NewExporter(jaeger.Options{
		AgentEndpoint: endpoint + ":6831",
		Process: jaeger.Process{
			ServiceName: serviceName,
			Tags: []jaeger.Tag{
				jaeger.StringTag("hostname", app.HostName()),
				jaeger.StringTag("ip", app.HostIPStr()),
				jaeger.StringTag("version", app.Version()),
			},
		},
	})
}
