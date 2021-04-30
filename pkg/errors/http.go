package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type HttpDetail struct {
	Value proto.Message
}

func (d HttpDetail) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(d.Value)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Type  string `json:"@type"`
		Value string `json:"value"`
	}{
		Type:  proto.MessageName(d.Value),
		Value: string(data),
	})
}

type HttpResponse struct {
	Error struct {
		Code    int          `json:"code"`
		Message string       `json:"message"`
		Status  string       `json:"status"`
		Details []HttpDetail `json:"details"`
	} `json:"error"`
}

func FromHttp(resp *http.Response) error {
	if resp.StatusCode < http.StatusBadRequest {
		return nil
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return E(Internal, err)
	}

	var hr HttpResponse
	if err = json.Unmarshal(data, &hr); err != nil {
		return E(Internal, err)
	}

	if hr.Error.Code > 0 {
		req := resp.Request
		return E(
			StrToCode(hr.Error.Status),
			errors.New(hr.Error.Message),
			fmt.Sprintf("%s %s", req.Method, req.URL),
		)
	}

	return E(Unknown, string(data))
}

func Http(requestId string, err error) (r HttpResponse) {
	defer func() {
		switch {
		case errors.Is(err, context.Canceled):
			r.Error.Status = CodeToStr(Canceled)
			r.Error.Code = CodeToHttp(Canceled)
		case errors.Is(err, context.DeadlineExceeded):
			r.Error.Status = CodeToStr(DeadlineExceeded)
			r.Error.Code = CodeToHttp(DeadlineExceeded)
		}
	}()

	if err, ok := err.(*Error); ok {
		code := Code(err)
		r.Error.Code = CodeToHttp(code)
		r.Error.Message = err.Error()
		r.Error.Status = CodeToStr(code)
		r.Error.Details = details(
			&errdetails.DebugInfo{StackEntries: Stack(err)},
			&errdetails.RequestInfo{RequestId: requestId},
			Detail(err))
		r.Error.Details = merge(r.Error.Details)
		return
	}

	s := status.Convert(err)
	c := s.Code()

	r.Error.Code = CodeToHttp(c)
	r.Error.Message = s.Message()
	r.Error.Status = CodeToStr(c)
	r.Error.Details = details(s.Details()...)
	return
}

func details(arr ...interface{}) (values []HttpDetail) {
	for _, value := range arr {
		switch v := value.(type) {
		case HttpDetail:
			values = append(values, v)
		case proto.Message:
			values = append(values, HttpDetail{Value: v})
		case []proto.Message:
			temp := make([]interface{}, len(v))
			for i := range v {
				temp[i] = v[i]
			}
			values = append(values, details(temp...)...)
		default:
			logrus.Warn("Unsupported detail type %T", v)
			continue
		}
	}
	return
}

func merge(values []HttpDetail) (details []HttpDetail) {
	details = []HttpDetail{
		{Value: &errdetails.RequestInfo{}},
		{Value: &errdetails.DebugInfo{}},
	}

	for _, detail := range values {
		switch message := detail.Value.(type) {
		case *errdetails.RequestInfo:
			proto.Merge(details[0].Value, message)
		case *errdetails.DebugInfo:
			proto.Merge(details[1].Value, message)
		default:
			details = append(details, detail)
		}
	}
	return
}

var codeToHttp = map[code]int{
	OK:                 http.StatusOK,
	Canceled:           499,
	Unknown:            http.StatusInternalServerError,
	InvalidArgument:    http.StatusBadRequest,
	DeadlineExceeded:   http.StatusGatewayTimeout,
	NotFound:           http.StatusNotFound,
	AlreadyExists:      http.StatusConflict,
	PermissionDenied:   http.StatusForbidden,
	ResourceExhausted:  http.StatusTooManyRequests,
	FailedPrecondition: http.StatusBadRequest,
	Aborted:            http.StatusConflict,
	OutOfRange:         http.StatusBadRequest,
	Unimplemented:      http.StatusNotImplemented,
	Internal:           http.StatusInternalServerError,
	Unavailable:        http.StatusServiceUnavailable,
	DataLoss:           http.StatusInternalServerError,
	Unauthenticated:    http.StatusUnauthorized,
}

func CodeToHttp(c code) int {
	return codeToHttp[c]
}

var codeToStr = map[code]string{
	OK:                 "OK",
	Canceled:           "CANCELLED",
	Unknown:            "UNKNOWN",
	InvalidArgument:    "INVALID_ARGUMENT",
	DeadlineExceeded:   "DEADLINE_EXCEEDED",
	NotFound:           "NOT_FOUND",
	AlreadyExists:      "ALREADY_EXISTS",
	PermissionDenied:   "PERMISSION_DENIED",
	ResourceExhausted:  "RESOURCE_EXHAUSTED",
	FailedPrecondition: "FAILED_PRECONDITION",
	Aborted:            "ABORTED",
	OutOfRange:         "OUT_OF_RANGE",
	Unimplemented:      "UNIMPLEMENTED",
	Internal:           "INTERNAL",
	Unavailable:        "UNAVAILABLE",
	DataLoss:           "DATA_LOSS",
	Unauthenticated:    "UNAUTHENTICATED",
}

func CodeToStr(c code) string {
	return codeToStr[c]
}

var strToCode = map[string]code{
	"OK":                  OK,
	"CANCELLED":           Canceled,
	"UNKNOWN":             Unknown,
	"INVALID_ARGUMENT":    InvalidArgument,
	"DEADLINE_EXCEEDED":   DeadlineExceeded,
	"NOT_FOUND":           NotFound,
	"ALREADY_EXISTS":      AlreadyExists,
	"PERMISSION_DENIED":   PermissionDenied,
	"RESOURCE_EXHAUSTED":  ResourceExhausted,
	"FAILED_PRECONDITION": FailedPrecondition,
	"ABORTED":             Aborted,
	"OUT_OF_RANGE":        OutOfRange,
	"UNIMPLEMENTED":       Unimplemented,
	"INTERNAL":            Internal,
	"UNAVAILABLE":         Unavailable,
	"DATA_LOSS":           DataLoss,
	"UNAUTHENTICATED":     Unauthenticated,
}

func StrToCode(str string) code {
	return strToCode[str]
}
