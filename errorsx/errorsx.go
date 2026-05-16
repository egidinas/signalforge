package errorsx

import (
	"errors"
	"net/http"
)

// Code is a machine-readable error category.
type Code string

const (
	CodeInvalidArgument Code = "invalid_argument"
	CodeNotFound        Code = "not_found"
	CodeConflict        Code = "conflict"
	CodeUnauthorized    Code = "unauthorized"
	CodeForbidden       Code = "forbidden"
	CodeDeadline        Code = "deadline"
	CodeUnavailable     Code = "unavailable"
	CodeInternal        Code = "internal"
)

type codedError struct {
	code Code
	err  error
}

func (e *codedError) Error() string { return e.err.Error() }
func (e *codedError) Unwrap() error { return e.err }

// WithCode wraps err with a Code. Preserves errors.Is / errors.As chains.
func WithCode(err error, code Code) error {
	if err == nil {
		return nil
	}
	return &codedError{code: code, err: err}
}

// CodeOf returns the Code attached to err. Unknown errors return CodeInternal.
func CodeOf(err error) Code {
	var ce *codedError
	if errors.As(err, &ce) {
		return ce.code
	}
	return CodeInternal
}

// IsCode reports whether err carries the given Code.
func IsCode(err error, code Code) bool {
	return CodeOf(err) == code
}

// HTTPStatus maps a Code to its canonical HTTP status code.
func HTTPStatus(code Code) int {
	switch code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeDeadline:
		return http.StatusGatewayTimeout
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
