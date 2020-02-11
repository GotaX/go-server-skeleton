package grpc

import (
	"context"
	"runtime/debug"

	grpcLogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"

	"github.com/GotaX/go-server-skeleton/pkg/errors"
	"github.com/GotaX/go-server-skeleton/pkg/ext"
)

func StreamClientErrorHandler() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		if stream, err = streamer(ctx, desc, cc, method, opts...); err != nil {
			err = errors.FromGrpc(ctx, err)
		}
		return
	}
}

func UnaryClientErrorHandler() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
			err = errors.FromGrpc(ctx, err)
		}
		return
	}
}

func StreamServerErrorHandler() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		if err = handler(srv, ss); err != nil {
			ri := ext.GetRequestInfo(ss.Context())
			err = errors.Grpc(ri.String(), err)
		}
		return
	}
}

func UnaryServerErrorHandler() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if resp, err = handler(ctx, req); err != nil {
			ri := ext.GetRequestInfo(ctx)
			err = errors.Grpc(ri.String(), err)
		}
		return
	}
}

func RecoveryHandler() grpcRecovery.Option {
	return grpcRecovery.WithRecoveryHandler(func(p interface{}) (err error) {
		err = xerrors.Errorf("%+v", p)
		logrus.WithError(err).Error(string(debug.Stack()))
		return
	})
}

func LogDecider() grpcLogrus.Option {
	const mHealth = "/grpc.health.v1.Health/Check"
	return grpcLogrus.WithDecider(func(fullMethodName string, err error) bool {
		return fullMethodName != mHealth
	})
}

func TraceHandler() stats.Handler {
	const mHealth = "grpc.health.v1.Health.Check"
	return &ocgrpc.ServerHandler{
		IsPublicEndpoint: false,
		StartOptions: trace.StartOptions{
			Sampler: func(parameters trace.SamplingParameters) trace.SamplingDecision {
				return trace.SamplingDecision{Sample: parameters.Name != mHealth}
			},
		},
	}
}
