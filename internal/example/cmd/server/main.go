package main

import (
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"

	"skeleton/internal/example/pkg/srvrest"
	"skeleton/internal/example/pkg/srvrpc"
	"skeleton/pkg/endpoint"
	"skeleton/pkg/endpoint/metrics"
	"skeleton/pkg/ext"
)

func main() {
	err := endpoint.Run(
		endpoint.Http("rest", ":8080", srvrest.Router()),
		endpoint.Http("metrics", ":8081", metrics.Router()),
		endpoint.Grpc("app", ":8082", srvrpc.Server()),
	)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.WithError(err).Error("Shutdown")
	}
	ext.WaitShutdown()
}
