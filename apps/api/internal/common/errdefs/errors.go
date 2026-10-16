package errdefs

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type ResponseError struct {
	Type      string   `json:"type" doc:"Stable, machine-readable error code" enum:"bad_request,unauthorized,forbidden,not_found,conflict,validation_error,rate_limited,internal_error,error" example:"not_found"`
	Status    int      `json:"status" doc:"HTTP status code" example:"404"`
	Message   string   `json:"message" doc:"Human-readable summary of what went wrong" example:"Project not found"`
	Details   []string `json:"details,omitempty" doc:"Optional actionable details, e.g. which field failed validation"`
	RequestID string   `json:"request_id,omitempty" doc:"Identifier of the request, quote it when reporting the failure so it can be found in the server logs" example:"api-7f9c4d2b-000042"`

	cause error
}

func (e *ResponseError) Error() string {
	return e.Message
}

func (e *ResponseError) GetStatus() int {
	return e.Status
}

func (e *ResponseError) Cause() error {
	return e.cause
}

func (e *ResponseError) Unwrap() error {
	return e.cause
}

func errorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusUnprocessableEntity:
		return "validation_error"
	case http.StatusTooManyRequests:
		return "rate_limited"
	case http.StatusInternalServerError:
		return "internal_error"
	default:
		return "error"
	}
}

var HumaErrorFunc = func(status int, message string, errs ...error) huma.StatusError {
	details := make([]string, 0, len(errs))
	for _, err := range errs {
		if detail := DetailMessage(err); detail != "" {
			details = append(details, detail)
		}
	}
	return &ResponseError{
		Type:    errorCode(status),
		Status:  status,
		Message: message,
		Details: details,
	}
}

func WithCause(err huma.StatusError, cause error) huma.StatusError {
	resp, ok := err.(*ResponseError)
	if !ok {
		return err
	}
	resp.cause = cause
	return resp
}

func DetailMessage(err error) string {
	if err == nil {
		return ""
	}
	var custom *CustomError
	if errors.As(err, &custom) {
		return custom.Message
	}
	return err.Error()
}

type ErrorType int

var (
	ErrAlreadyBootstrapped = errors.New("already bootstrapped")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidInput        = NewCustomError(ErrTypeInvalidInput, "")
	ErrNotFound            = NewCustomError(ErrTypeNotFound, "")
	ErrConflict            = NewCustomError(ErrTypeConflict, "")
)

const (
	ErrTypeInvalidInput ErrorType = iota
	ErrTypeNotFound
	ErrTypeConflict
)

var errorTypeStrings = map[ErrorType]string{
	ErrTypeInvalidInput: "ErrInvalidInput",
	ErrTypeNotFound:     "ErrNotFound",
	ErrTypeConflict:     "ErrConflict",
}

func (e ErrorType) String() string {
	if s, ok := errorTypeStrings[e]; ok {
		return s
	}
	return "ErrUnknown"
}

func MaskAsNotFound(err error, message string) error {
	if errors.Is(err, ErrUnauthorized) {
		return NewCustomError(ErrTypeNotFound, message)
	}
	return err
}

type InternalError struct {
	Summary string
	cause   error
}

func (e *InternalError) Error() string {
	if e.cause == nil {
		return e.Summary
	}
	return fmt.Sprintf("%s: %v", e.Summary, e.cause)
}

func (e *InternalError) Unwrap() error {
	return e.cause
}

func NewInternalError(cause error, summary string) error {
	return &InternalError{Summary: summary, cause: cause}
}

type CustomError struct {
	Type    ErrorType
	Message string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type.String(), e.Message)
}

func (e *CustomError) Is(target error) bool {
	t, ok := target.(*CustomError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

func NewCustomError(t ErrorType, message string) *CustomError {
	return &CustomError{
		Type:    t,
		Message: message,
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
