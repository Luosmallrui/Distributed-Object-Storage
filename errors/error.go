package errors

import (
	"errors"
	"net/http"
)

// common errs
var (
	ErrNotFound   = errors.New("resource not found")
	ErrBadRequest = errors.New("bad request")
)

// ErrorToHTTPCode ..
func ErrorToHTTPCode(err error) int {
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}
	if errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
