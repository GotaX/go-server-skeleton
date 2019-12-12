package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrResourceExhausted  = errors.New("resource exhausted")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrBadRequest         = errors.New("bad request")
)

type NotFoundError struct {
	Type string
	ID   string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf(
		"%s, type = %q, id = %q",
		ErrNotFound, e.Type, e.ID)
}

func (e NotFoundError) Unwrap() error {
	return ErrNotFound
}

func (e NotFoundError) Detail() proto.Message {
	return &errdetails.ResourceInfo{
		ResourceType: e.Type,
		ResourceName: e.ID,
	}
}

type ResourceExhaustedError struct {
	Subject     string
	Description string
}

func (e ResourceExhaustedError) Error() string {
	return fmt.Sprintf(
		"%s, subject = %q, desc = %q",
		ErrResourceExhausted, e.Subject, e.Description)
}

func (e ResourceExhaustedError) Unwrap() error {
	return ErrResourceExhausted
}

func (e ResourceExhaustedError) Detail() proto.Message {
	return &errdetails.QuotaFailure{
		Violations: []*errdetails.QuotaFailure_Violation{{
			Subject:     e.Subject,
			Description: e.Description,
		}},
	}
}

type FailedPreconditionError struct {
	Type        string
	Subject     string
	Description string
}

func (e FailedPreconditionError) Error() string {
	return fmt.Sprintf(
		"%s, type = %q, subject = %q, desc = %q",
		ErrFailedPrecondition, e.Type, e.Subject, e.Description)
}

func (e FailedPreconditionError) Unwrap() error {
	return ErrFailedPrecondition
}

func (e FailedPreconditionError) Detail() proto.Message {
	return &errdetails.PreconditionFailure_Violation{
		Type:        e.Type,
		Subject:     e.Subject,
		Description: e.Description,
	}
}

type BadRequestError struct {
	FieldViolations []*errdetails.BadRequest_FieldViolation `json:"field_violations"`
}

func NewBadRequestError(violations ...errdetails.BadRequest_FieldViolation) *BadRequestError {
	arr := make([]*errdetails.BadRequest_FieldViolation, len(violations))
	for i := range violations {
		arr[i] = &violations[i]
	}
	return &BadRequestError{FieldViolations: arr}
}

func (e BadRequestError) Error() string {
	sb := &strings.Builder{}
	_, _ = fmt.Fprint(sb, ErrBadRequest, ";")
	for _, v := range e.FieldViolations {
		_, _ = fmt.Fprintf(sb, "field = %q, desc = %q; ", v.Field, v.Description)
	}
	return sb.String()
}

func (e BadRequestError) Unwrap() error {
	return ErrBadRequest
}

func (e BadRequestError) Detail() proto.Message {
	return &errdetails.BadRequest{FieldViolations: e.FieldViolations}
}
