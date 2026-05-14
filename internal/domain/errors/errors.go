package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a structured application error.
type AppError struct {
	Code       int
	Message    string
	Detail     string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Sentinel errors
var (
	ErrNotFound           = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
	ErrUnauthorized       = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden          = &AppError{Code: http.StatusForbidden, Message: "forbidden"}
	ErrConflict           = &AppError{Code: http.StatusConflict, Message: "resource already exists"}
	ErrBadRequest         = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrInternalServer     = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
	ErrTokenExpired       = &AppError{Code: http.StatusUnauthorized, Message: "token has expired"}
	ErrTokenInvalid       = &AppError{Code: http.StatusUnauthorized, Message: "token is invalid"}
	ErrTokenRevoked       = &AppError{Code: http.StatusUnauthorized, Message: "token has been revoked"}
	ErrInvalidCredentials = &AppError{Code: http.StatusUnauthorized, Message: "invalid credentials"}
	ErrAccountInactive    = &AppError{Code: http.StatusForbidden, Message: "account is inactive"}
)

// New creates a new AppError with a wrapped cause.
func New(code int, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// Wrap wraps an existing AppError with additional detail.
func Wrap(appErr *AppError, detail string, err error) *AppError {
	return &AppError{
		Code:    appErr.Code,
		Message: appErr.Message,
		Detail:  detail,
		Err:     err,
	}
}

// IsNotFound checks if error is a not-found error.
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

// IsConflict checks if error is a conflict error.
func IsConflict(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusConflict
	}
	return false
}