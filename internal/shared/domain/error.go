package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// Error is the unified error value carried through application and adapter layers.
type Error struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	Status    int            `json:"-"`
	RequestID string         `json:"-"`
	Cause     error          `json:"-"`
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) WithRequestID(id string) *Error {
	clone := *e
	clone.RequestID = id
	return &clone
}

func (e *Error) WithDetail(key string, value any) *Error {
	clone := *e
	if clone.Details == nil {
		clone.Details = map[string]any{}
	}
	clone.Details[key] = value
	return &clone
}

func New(code, message string, status int) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

func Wrap(code, message string, status int, cause error) *Error {
	return &Error{Code: code, Message: message, Status: status, Cause: cause}
}

func IsCode(err error, code string) bool {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Code == code
}

var (
	ErrBadRequest      = New("bad_request", "the request is invalid", http.StatusBadRequest)
	ErrUnauthorized    = New("unauthorized", "authentication is required", http.StatusUnauthorized)
	ErrForbidden       = New("forbidden", "the caller is not allowed to perform this action", http.StatusForbidden)
	ErrNotFound        = New("not_found", "the requested resource does not exist", http.StatusNotFound)
	ErrConflict        = New("conflict", "the request conflicts with the current resource state", http.StatusConflict)
	ErrPrecondition    = New("precondition_failed", "the resource version does not match", http.StatusPreconditionFailed)
	ErrTooManyRequests = New("too_many_requests", "rate limit exceeded", http.StatusTooManyRequests)
	ErrInternal        = New("internal_error", "an unexpected error occurred", http.StatusInternalServerError)
	ErrUnavailable     = New("service_unavailable", "the service is temporarily unavailable", http.StatusServiceUnavailable)
)

func Validation(details map[string]any) *Error {
	return &Error{
		Code:    "validation_failed",
		Message: "one or more fields failed validation",
		Status:  http.StatusBadRequest,
		Details: details,
	}
}
