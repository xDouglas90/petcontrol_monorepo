package apperror

import (
	"errors"
	"net/http"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrNotFound            = errors.New("resource not found")
	ErrInternal            = errors.New("internal error")
	ErrServiceUnavailable  = errors.New("service unavailable")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("conflict")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrAccountInactive     = errors.New("account inactive")
	ErrAccountLocked       = errors.New("account locked")
	ErrEmailNotVerified    = errors.New("email not verified")
)

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrAccountInactive), errors.Is(err, ErrAccountLocked), errors.Is(err, ErrEmailNotVerified):
		return http.StatusForbidden
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrUnprocessableEntity):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrServiceUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
