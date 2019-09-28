package server

import (
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"

	"skeleton/pkg/ext"
)

func Grpc(address string) (*grpc.ClientConn, error) {
	return grpc.Dial(address,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithUnaryInterceptor(ext.UnaryClientErrorHandler()),
		grpc.WithStreamInterceptor(ext.StreamClientErrorHandler()))
}
