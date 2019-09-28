package errors

import (
	"errors"
	"fmt"

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
	Field       string
	Description string
}

func (e BadRequestError) Error() string {
	return fmt.Sprintf(
		"%s, field = %q, desc = %q",
		ErrBadRequest, e.Field, e.Description)
}

func (e BadRequestError) Unwrap() error {
	return ErrBadRequest
}

func (e BadRequestError) Detail() proto.Message {
	return &errdetails.BadRequest_FieldViolation{
		Field:       e.Field,
		Description: e.Description,
	}
}
