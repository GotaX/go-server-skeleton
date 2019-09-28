package endpoint

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/sync/errgroup"

	"github.com/GotaX/go-server-skeleton/pkg/ext"
)

type Endpoint interface {
	Name() string
	Run() error
	Stop() error
}

func Run(endpoints ...Endpoint) error {
	group := &errgroup.Group{}
	for _, endpoint := range endpoints {
		e := endpoint
		entry := logrus.WithField("name", e.Name())

		group.Go(func() error {
			entry.Debug("Listening ...")
			return e.Run()
		})

		ext.OnShutdown(func() {
			entry.Debug("Shutdown...")
			if err := e.Stop(); err != nil {
				entry.WithError(err).Warn("Fail to shutdown")
			} else {
				entry.Debug("Stopped")
			}
		})
	}
	return group.Wait()
}

type rest struct {
	name string
	srv  *http.Server
}

func Http(name, addr string, handler http.Handler) Endpoint {
	return &rest{
		name: name,
		srv: &http.Server{
			Addr: addr,
			Handler: &ochttp.Handler{
				Handler:     handler,
				Propagation: ext.Propagation,
			},
		},
	}
}

func (e rest) Name() string {
	return fmt.Sprintf("Http (%s), %s", e.srv.Addr, e.name)
}

func (e rest) Run() error {
	return e.srv.ListenAndServe()
}

func (e rest) Stop() error {
	return e.srv.Shutdown(context.Background())
}

type GrpcServer interface {
	Serve(lis net.Listener) error
	Stop()
}

type grpc struct {
	name string
	addr string
	srv  GrpcServer
}

func Grpc(name, address string, server GrpcServer) Endpoint {
	return &grpc{
		name: name,
		addr: address,
		srv:  server,
	}
}

func (e grpc) Name() string {
	return fmt.Sprintf("GRPC (%s), %s", e.addr, e.name)
}

func (e grpc) Run() error {
	if lis, err := net.Listen("tcp", e.addr); err != nil {
		return err
	} else {
		return e.srv.Serve(lis)
	}
}

func (e grpc) Stop() error {
	e.srv.Stop()
	return nil
}
