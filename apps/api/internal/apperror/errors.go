package apperror

import "errors"

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("resource not found")
	ErrInternal   = errors.New("internal error")
)
