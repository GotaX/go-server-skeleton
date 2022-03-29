package errors

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type code = codes.Code

const (
	Ignore             = codes.OK
	OK                 = codes.OK
	Canceled           = codes.Canceled
	Unknown            = codes.Unknown
	InvalidArgument    = codes.InvalidArgument
	DeadlineExceeded   = codes.DeadlineExceeded
	NotFound           = codes.NotFound
	AlreadyExists      = codes.AlreadyExists
	PermissionDenied   = codes.PermissionDenied
	ResourceExhausted  = codes.ResourceExhausted
	FailedPrecondition = codes.FailedPrecondition
	Aborted            = codes.Aborted
	OutOfRange         = codes.OutOfRange
	Unimplemented      = codes.Unimplemented
	Internal           = codes.Internal
	Unavailable        = codes.Unavailable
	DataLoss           = codes.DataLoss
	Unauthenticated    = codes.Unauthenticated
)

type Op string

type Error struct {
	Op   Op
	Code code
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"op = %s, code = %s, desc = %s",
		Ops(e), Code(e), Cause(e))
}

func (e *Error) Unwrap() error {
	return e.Err
}

func Is(err, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}

func E(args ...interface{}) error {
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			e.Op = Op(arg)
		case Op:
			e.Op = arg
		case code:
			e.Code = arg
		case error:
			e.Err = arg
		case nil:
			continue
		default:
			log.Panicf("invalid argument: %T", arg)
		}
	}
	if e.Err != nil {
		return e
	} else {
		return nil
	}
}

func Ops(err error) (ops []Op) {
	if err, ok := err.(*Error); ok && err.Op != "" {
		ops = append(ops, err.Op)
	}
	if subErr := errors.Unwrap(err); subErr != nil {
		ops = append(ops, Ops(subErr)...)
	}
	return
}

func Stack(err error) []string {
	ops := Ops(err)
	stack := make([]string, len(ops))
	for i, op := range ops {
		stack[i] = string(op)
	}
	return stack
}

func Code(err error) (code codes.Code) {
	if err, ok := err.(*Error); ok && err.Code != Ignore {
		return err.Code
	}
	if sub := errors.Unwrap(err); sub != nil {
		return Code(sub)
	}
	return Unknown
}

func Desc(err error) string {
	if err := Cause(err); err != nil {
		return err.Error()
	}
	return ""
}

// Return first error not *errors.Error
func Cause(err error) error {
	if _, ok := err.(*Error); !ok {
		return err
	}
	if sub := errors.Unwrap(err); sub != nil {
		return Cause(sub)
	}
	return err
}

func Detail(err error) []proto.Message {
	if err, ok := err.(GrpcDetail); ok {
		return []proto.Message{err.Detail()}
	}

	if err, ok := status.FromError(err); ok {
		details := err.Details()
		messages := make([]proto.Message, 0, len(details))
		for _, detail := range details {
			if msg, ok := detail.(proto.Message); ok {
				messages = append(messages, msg)
			}
		}
		return messages
	}

	if sub := errors.Unwrap(err); sub != nil {
		return Detail(sub)
	} else {
		messages := make([]proto.Message, 0, 1)
		messages = append(messages, &errdetails.ErrorInfo{Reason: err.Error()})
		return messages
	}
}
