package rpc

import (
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	grpc2 "github.com/GotaX/go-server-skeleton/pkg/ext/grpc"
)

func init() {
	grpcPrometheus.EnableHandlingTimeHistogram()
}

type Service interface {
	Register(server *grpc.Server)
}

func registerHealthServer(s *grpc.Server) {
	hsrv := health.NewServer()
	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hsrv)
}

type GrpcConfiguration struct {
	LogEntry     *logrus.Entry
	LogExtractor grpcCtxTags.RequestFieldExtractorFunc
	LogDecider   func(fullMethodName string, err error) bool
	services     []Service
}

func (c *GrpcConfiguration) Register(service Service) {
	c.services = append(c.services, service)
}

func NewGrpcServer(configure func(c *GrpcConfiguration)) *grpc.Server {
	const mHealth = "/grpc.health.v1.Health/Check"
	c := &GrpcConfiguration{
		LogEntry:     logrus.NewEntry(logrus.StandardLogger()),
		LogDecider:   func(fullMethodName string, err error) bool { return fullMethodName != mHealth },
		LogExtractor: func(fullMethod string, req interface{}) map[string]interface{} { return nil },
	}
	configure(c)

	s := grpc.NewServer(
		grpc.StatsHandler(grpc2.TraceHandler()),
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(
			grpcCtxTags.StreamServerInterceptor(grpcCtxTags.WithFieldExtractor(c.LogExtractor)),
			grpcLogrus.StreamServerInterceptor(c.LogEntry, grpc2.LogDecider()),
			grpcPrometheus.StreamServerInterceptor,
			grpcRecovery.StreamServerInterceptor(grpc2.RecoveryHandler()),
			grpc2.StreamServerErrorHandler(),
		)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
			grpcCtxTags.UnaryServerInterceptor(grpcCtxTags.WithFieldExtractor(c.LogExtractor)),
			grpcLogrus.UnaryServerInterceptor(c.LogEntry, grpc2.LogDecider()),
			grpcPrometheus.UnaryServerInterceptor,
			grpcRecovery.UnaryServerInterceptor(grpc2.RecoveryHandler()),
			grpc2.UnaryServerErrorHandler(),
		)))

	for _, srv := range c.services {
		srv.Register(s)
	}

	registerHealthServer(s)
	reflection.Register(s)
	return s
}
