package srvrpc

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"

	"github.com/GotaX/go-server-skeleton/internal/example/pkg/cfg"
	"github.com/GotaX/go-server-skeleton/internal/example/pkg/rpc"
	"github.com/GotaX/go-server-skeleton/pkg/endpoint"
	rpc2 "github.com/GotaX/go-server-skeleton/pkg/endpoint/rpc"
	"github.com/GotaX/go-server-skeleton/pkg/errors"
)

type mServer struct{ srv *grpc.Server }

func (s mServer) Serve(lis net.Listener) error { return s.srv.Serve(lis) }
func (s mServer) Stop()                        { s.srv.GracefulStop() }

func Server() endpoint.GrpcServer {
	srv := rpc2.NewGrpcServer(func(c *rpc2.GrpcConfiguration) {
		c.LogEntry = accessLogger()
		c.LogExtractor = extractor

		c.Register(newService())
	})
	return mServer{srv: srv}
}

func accessLogger() *logrus.Entry {
	var logger *logrus.Logger
	cfg.LogGrpc(&logger)
	return logrus.NewEntry(logger)
}

func extractor(fullMethod string, req interface{}) map[string]interface{} {
	// TODO: Do access log
	if req, ok := req.(*rpc.HelloRequest); ok {
		return map[string]interface{}{"name": req.Greeting}
	}
	return nil
}

type Service struct {
}

func newService() rpc2.Service {
	return &Service{}
}

func (s *Service) Register(gs *grpc.Server) {
	// TODO: Register services
	rpc.RegisterHelloServiceServer(gs, s)
}

func (s *Service) SayHello(ctx context.Context, req *rpc.HelloRequest) (*rpc.HelloResponse, error) {
	const op errors.Op = "server.SayHello"

	if num := rand.Int() % 2; num == 0 {
		return nil, errors.E(op, errors.Internal, xerrors.New("crash"))
	}

	dur := time.Duration(rand.Int63n(int64(2 * time.Second)))
	logrus.Debugf("Sleep %v...", dur)
	time.Sleep(dur)

	reply := fmt.Sprintf("Welcome %s", req.Greeting)
	return &rpc.HelloResponse{Reply: reply}, nil
}
