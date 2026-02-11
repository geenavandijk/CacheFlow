package httpx

import (
	"errors"
	"net/http"
)

type Error struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func (e *Error) WithErr(err error) *Error {
	e.Err = err
	return e
}

func BadRequest(message string, fields map[string]string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: "bad_request", Message: message, Fields: fields}
}

func Unauthorized(message string) *Error {
	return &Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Status: http.StatusForbidden, Code: "forbidden", Message: message}
}

func NotFound(message string) *Error {
	return &Error{Status: http.StatusNotFound, Code: "not_found", Message: message}
}

func Conflict(message string, fields map[string]string) *Error {
	return &Error{Status: http.StatusConflict, Code: "conflict", Message: message, Fields: fields}
}

func Internal(message string) *Error {
	return &Error{Status: http.StatusInternalServerError, Code: "internal", Message: message}
}

func ToAPIError(err error, requestID string) (int, *APIError) {
	var he *Error
	if errors.As(err, &he) {
		return he.Status, &APIError{
			Code:      he.Code,
			Message:   he.Message,
			Fields:    he.Fields,
			RequestID: requestID,
		}
	}

	// Unknown error => do not leak internals
	return http.StatusInternalServerError, &APIError{
		Code:      "internal",
		Message:   "Internal server error",
		RequestID: requestID,
	}
}
