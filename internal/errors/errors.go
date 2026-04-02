package errors

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInternal     = errors.New("internal server error")
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
	}
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewAPIError(status int, message string, err error) *APIError {
	if message == "" && err != nil {
		message = err.Error()
	}
	return &APIError{Status: status, Message: message, Err: err}
}

func NewNotFound(message string, err error) error {
	return NewAPIError(http.StatusNotFound, message, err)
}

func NewConflict(message string, err error) error {
	return NewAPIError(http.StatusConflict, message, err)
}

func NewBadRequest(message string, err error) error {
	return NewAPIError(http.StatusBadRequest, message, err)
}

func NewUnauthorized(message string, err error) error {
	return NewAPIError(http.StatusUnauthorized, message, err)
}

func NewForbidden(message string, err error) error {
	return NewAPIError(http.StatusForbidden, message, err)
}

func NewInternal(message string, err error) error {
	if message == "" {
		message = "internal server error"
	}
	return NewAPIError(http.StatusInternalServerError, message, err)
}

func HTTPError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	if apiErr, ok := errors.AsType[*APIError](err); ok {
		if apiErr.Message == "" {
			return apiErr.Status, http.StatusText(apiErr.Status)
		}
		return apiErr.Status, apiErr.Message
	}

	if errors.Is(err, context.Canceled) {
		return 499, "client closed request"
	}
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound, http.StatusText(http.StatusNotFound)
	}
	if errors.Is(err, ErrConflict) {
		return http.StatusConflict, http.StatusText(http.StatusConflict)
	}
	if errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest, http.StatusText(http.StatusBadRequest)
	}
	if errors.Is(err, ErrUnauthorized) {
		return http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)
	}
	if errors.Is(err, ErrForbidden) {
		return http.StatusForbidden, http.StatusText(http.StatusForbidden)
	}

	return http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)
}
