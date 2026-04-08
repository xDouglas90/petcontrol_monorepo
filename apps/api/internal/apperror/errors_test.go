package apperror

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "bad request", err: ErrBadRequest, want: http.StatusBadRequest},
		{name: "unauthorized", err: ErrUnauthorized, want: http.StatusUnauthorized},
		{name: "invalid credentials", err: ErrInvalidCredentials, want: http.StatusUnauthorized},
		{name: "forbidden", err: ErrForbidden, want: http.StatusForbidden},
		{name: "account inactive", err: ErrAccountInactive, want: http.StatusForbidden},
		{name: "account locked", err: ErrAccountLocked, want: http.StatusForbidden},
		{name: "email not verified", err: ErrEmailNotVerified, want: http.StatusForbidden},
		{name: "not found", err: ErrNotFound, want: http.StatusNotFound},
		{name: "conflict", err: ErrConflict, want: http.StatusConflict},
		{name: "unprocessable entity", err: ErrUnprocessableEntity, want: http.StatusUnprocessableEntity},
		{name: "fallback", err: ErrInternal, want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, HTTPStatus(tt.err))
		})
	}
}
