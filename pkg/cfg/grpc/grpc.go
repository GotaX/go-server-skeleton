package grpc

import (
	"time"

	retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	driver "google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/GotaX/go-server-skeleton/pkg/cfg"
	"github.com/GotaX/go-server-skeleton/pkg/ext/grpc"
)

var Option = cfg.Option{
	Name:     "GRPC",
	OnCreate: newGrpc,
}

func newGrpc(source cfg.Scanner) (interface{}, error) {
	var target string
	if err := source.Scan(&target); err != nil {
		return nil, err
	}

	opts := []retry.CallOption{
		retry.WithBackoff(retry.BackoffExponential(100 * time.Millisecond)),
		retry.WithCodes(codes.Unavailable, codes.ResourceExhausted, codes.Aborted),
	}
	return driver.Dial(target, driver.WithInsecure(),
		driver.WithStatsHandler(&ocgrpc.ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler: trace.AlwaysSample(),
			},
		}),
		driver.WithChainStreamInterceptor(
			retry.StreamClientInterceptor(opts...),
			grpc.StreamClientErrorHandler(),
		),
		driver.WithChainUnaryInterceptor(
			retry.UnaryClientInterceptor(opts...),
			grpc.UnaryClientErrorHandler(),
		))
}
