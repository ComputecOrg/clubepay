package domain

import "net/http"

// ServiceError is a typed error that carries an HTTP status code.
type ServiceError struct {
	Code    int
	Message string
	Err     error
}

func NewServiceError(code int, message string, err error) *ServiceError {
	return &ServiceError{Code: code, Message: message, Err: err}
}

func (e *ServiceError) Error() string { return e.Message }

func (e *ServiceError) Unwrap() error { return e.Err }

// Constructors for common error types.
// Note: prefixed with New to avoid conflict with sentinel vars in errors.go.

func NewErrNotFound(msg string) *ServiceError {
	return NewServiceError(http.StatusNotFound, msg, nil)
}

func NewErrConflict(msg string) *ServiceError {
	return NewServiceError(http.StatusConflict, msg, nil)
}

func NewErrForbidden(msg string) *ServiceError {
	return NewServiceError(http.StatusForbidden, msg, nil)
}

func NewErrBadRequest(msg string) *ServiceError {
	return NewServiceError(http.StatusBadRequest, msg, nil)
}

func NewErrInternal(msg string, err error) *ServiceError {
	return NewServiceError(http.StatusInternalServerError, msg, err)
}
