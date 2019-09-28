package main

import (
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"

	"github.com/GotaX/go-server-skeleton/internal/example/pkg/srvrest"
	"github.com/GotaX/go-server-skeleton/internal/example/pkg/srvrpc"
	"github.com/GotaX/go-server-skeleton/pkg/endpoint"
	"github.com/GotaX/go-server-skeleton/pkg/endpoint/metrics"
	"github.com/GotaX/go-server-skeleton/pkg/ext"
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
