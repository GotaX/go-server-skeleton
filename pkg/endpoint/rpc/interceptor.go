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

	"skeleton/pkg/ext"
)

func init() {
	grpcPrometheus.EnableHandlingTimeHistogram()
}

type Service interface {
	Register(server *grpc.Server)
}

/*
NewGrpcServer create a basic server with: access log, health check, prometheus, recovery

entry is logger for access log, extractor is used to extract fields from request to log
*/
func NewGrpcServer(entry *logrus.Entry, extractor grpcCtxTags.RequestFieldExtractorFunc, services ...Service) *grpc.Server {
	s := grpc.NewServer(
		grpc.StatsHandler(ext.TraceHandler()),
		grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(
			grpcCtxTags.StreamServerInterceptor(grpcCtxTags.WithFieldExtractor(extractor)),
			grpcLogrus.StreamServerInterceptor(entry, ext.LogDecider()),
			grpcPrometheus.StreamServerInterceptor,
			grpcRecovery.StreamServerInterceptor(ext.RecoveryHandler()),
			ext.StreamServerErrorHandler(),
		)),
		grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
			grpcCtxTags.UnaryServerInterceptor(grpcCtxTags.WithFieldExtractor(extractor)),
			grpcLogrus.UnaryServerInterceptor(entry, ext.LogDecider()),
			grpcPrometheus.UnaryServerInterceptor,
			grpcRecovery.UnaryServerInterceptor(ext.RecoveryHandler()),
			ext.UnaryServerErrorHandler(),
		)))

	registerHealthServer(s)
	reflection.Register(s)

	for _, srv := range services {
		srv.Register(s)
	}
	return s
}

func registerHealthServer(s *grpc.Server) {
	hsrv := health.NewServer()
	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hsrv)
}
