package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

type AppError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Status  int          `json:"status"`
	Details []FieldError `json:"details,omitempty"`
	cause   error
}

func (e *AppError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error { return e.cause }

func New(code string, status int, message string) *AppError {
	return &AppError{Code: code, Status: status, Message: message}
}

func Wrap(err error, code string, status int, message string) *AppError {
	return &AppError{Code: code, Status: status, Message: message, cause: err}
}

func (e *AppError) WithDetails(details ...FieldError) *AppError {
	e.Details = append(e.Details, details...)
	return e
}

func (e *AppError) WithCause(err error) *AppError {
	e.cause = err
	return e
}

func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}

var (
	ErrInternal           = New("INTERNAL_ERROR", http.StatusInternalServerError, "internal server error")
	ErrBadRequest         = New("BAD_REQUEST", http.StatusBadRequest, "bad request")
	ErrUnauthorized       = New("AUTH_TOKEN_MISSING", http.StatusUnauthorized, "authentication required")
	ErrTokenExpired       = New("AUTH_TOKEN_EXPIRED", http.StatusUnauthorized, "token expired")
	ErrTokenInvalid       = New("AUTH_TOKEN_INVALID", http.StatusUnauthorized, "token invalid")
	ErrForbidden          = New("AUTH_FORBIDDEN", http.StatusForbidden, "insufficient permissions")
	ErrNotFound           = New("NOT_FOUND", http.StatusNotFound, "resource not found")
	ErrConflict           = New("CONFLICT", http.StatusConflict, "resource conflict")
	ErrValidation         = New("VALIDATION_ERROR", http.StatusUnprocessableEntity, "validation failed")
	ErrRateLimited        = New("RATE_LIMITED", http.StatusTooManyRequests, "rate limit exceeded")
	ErrServiceUnavailable = New("SERVICE_UNAVAILABLE", http.StatusServiceUnavailable, "service unavailable")
)
