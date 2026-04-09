package whatsapp

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWebhookHandler_VerifiesToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler("secret-token", logger)

	req := httptest.NewRequest(http.MethodGet, "/webhook/whatsapp?hub.verify_token=secret-token&hub.challenge=12345", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "12345", res.Body.String())
}

func TestNewWebhookHandler_RejectsInvalidToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler("secret-token", logger)

	req := httptest.NewRequest(http.MethodGet, "/webhook/whatsapp?hub.verify_token=wrong&hub.challenge=12345", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestNewWebhookHandler_AcceptsPost(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewWebhookHandler("secret-token", logger)

	req := httptest.NewRequest(http.MethodPost, "/webhook/whatsapp", http.NoBody)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "ok", res.Body.String())
}
