package errors

import (
	"context"
	"errors"

	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type GrpcDetail interface {
	Detail() proto.Message
}

func Grpc(requestId string, err error) error {
	code := Code(err)
	if code == Unknown {
		switch {
		case errors.Is(err, context.Canceled):
			code = Canceled
		case errors.Is(err, context.DeadlineExceeded):
			code = DeadlineExceeded
		}
	}

	details := []proto.Message{
		&errdetails.DebugInfo{StackEntries: Stack(err)},
		&errdetails.RequestInfo{RequestId: requestId},
	}

	for _, detail := range Detail(err) {
		if isAcceptDetail(detail) {
			details = append(details, detail)
		}
	}

	st := status.New(code, Desc(err))
	st, _ = st.WithDetails(details...)
	return st.Err()
}

func FromGrpc(ctx context.Context, err error) error {
	if s := status.FromContextError(ctx.Err()); s != nil {
		return E(s.Code(), err)
	}
	if s, ok := status.FromError(err); ok {
		return E(s.Code(), err)
	}
	return err
}
