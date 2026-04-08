package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	appjwt "github.com/xdouglas90/petcontrol_monorepo/internal/jwt"
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
)

type publisherStub struct {
	err      error
	called   bool
	captured queue.DummyNotificationPayload
}

func (p *publisherStub) EnqueueDummyNotification(_ context.Context, payload queue.DummyNotificationPayload) error {
	p.called = true
	p.captured = payload
	return p.err
}

func (p *publisherStub) Close() error { return nil }

func TestWorkerHandler_EnqueueDummyNotification(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pub := &publisherStub{}
	h := NewWorkerHandler(pub)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_claims", appjwt.Claims{UserID: "user-1", CompanyID: "11111111-1111-1111-1111-111111111111"})
		c.Set("company_id", pgtype.UUID{Valid: true})
		c.Next()
	})
	router.POST("/enqueue", h.EnqueueDummyNotification)

	body, err := json.Marshal(map[string]string{"message": "hello"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusAccepted, res.Code)
	require.True(t, pub.called)
	require.Equal(t, "user-1", pub.captured.UserID)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", pub.captured.CompanyID)
	require.Equal(t, "hello", pub.captured.Message)
}

func TestWorkerHandler_EnqueueDummyNotification_MissingContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pub := &publisherStub{}
	h := NewWorkerHandler(pub)

	router := gin.New()
	router.POST("/enqueue", h.EnqueueDummyNotification)

	req := httptest.NewRequest(http.MethodPost, "/enqueue", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusForbidden, res.Code)
	require.False(t, pub.called)
}

func TestWorkerHandler_EnqueueDummyNotification_EnqueueError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pub := &publisherStub{err: context.DeadlineExceeded}
	h := NewWorkerHandler(pub)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_claims", appjwt.Claims{UserID: "user-1", CompanyID: "11111111-1111-1111-1111-111111111111"})
		c.Set("company_id", pgtype.UUID{Valid: true})
		c.Next()
	})
	router.POST("/enqueue", h.EnqueueDummyNotification)

	body, err := json.Marshal(map[string]string{"message": "retry"})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.True(t, pub.called)
}

func TestWorkerHandler_EnqueueDummyNotification_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	pub := &publisherStub{}
	h := NewWorkerHandler(pub)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("auth_claims", appjwt.Claims{UserID: "user-1", CompanyID: "11111111-1111-1111-1111-111111111111"})
		c.Set("company_id", pgtype.UUID{Valid: true})
		c.Next()
	})
	router.POST("/enqueue", h.EnqueueDummyNotification)

	req := httptest.NewRequest(http.MethodPost, "/enqueue", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnprocessableEntity, res.Code)
	require.False(t, pub.called)
	require.Contains(t, res.Body.String(), "invalid request body")
}
