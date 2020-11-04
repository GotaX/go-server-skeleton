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
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"time"

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

	// Ref: https://github.com/grpc/grpc-go/blob/master/examples/features/keepalive/server/main.go
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}

	kasp := keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 5 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout:               1 * time.Second,  // Wait 1 second for the ping ack before assuming the connection is dead
	}

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
		)),
		grpc.KeepaliveEnforcementPolicy(kaep),
		grpc.KeepaliveParams(kasp))

	for _, srv := range c.services {
		srv.Register(s)
	}

	registerHealthServer(s)
	reflection.Register(s)
	return s
}
