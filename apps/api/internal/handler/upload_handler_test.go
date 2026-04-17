package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type fakeUploadService struct {
	createIntentResult service.CreateUploadIntentOutput
	createIntentErr    error
	completeResult     service.CompleteUploadOutput
	completeErr        error
	lastCreateInput    service.CreateUploadIntentInput
	lastCompleteInput  service.CompleteUploadInput
}

func (f *fakeUploadService) CreateIntent(_ context.Context, input service.CreateUploadIntentInput) (service.CreateUploadIntentOutput, error) {
	f.lastCreateInput = input
	return f.createIntentResult, f.createIntentErr
}

func (f *fakeUploadService) Complete(_ context.Context, input service.CompleteUploadInput) (service.CompleteUploadOutput, error) {
	f.lastCompleteInput = input
	return f.completeResult, f.completeErr
}

func TestUploadHandlerCreateIntent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceStub := &fakeUploadService{
		createIntentResult: service.CreateUploadIntentOutput{
			UploadURL: "https://signed.example.com/upload",
			Method:    http.MethodPut,
			Headers:   http.Header{"Content-Type": []string{"image/png"}},
			ObjectKey: "uploads/pets/image_url/2026/04/test.png",
			PublicURL: "https://cdn.example.com/uploads/pets/image_url/2026/04/test.png",
			ExpiresAt: time.Date(2026, 4, 17, 19, 0, 0, 0, time.UTC),
		},
	}
	handlerUnderTest := NewUploadHandler(serviceStub)

	router := gin.New()
	router.POST("/uploads/intents", handlerUnderTest.CreateIntent)

	body, err := json.Marshal(map[string]any{
		"resource":     " pets ",
		"field":        " image_url ",
		"file_name":    " thor.png ",
		"content_type": " image/png ",
		"size_bytes":   1024,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/uploads/intents", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Equal(t, "pets", serviceStub.lastCreateInput.Resource)
	require.Equal(t, "image_url", serviceStub.lastCreateInput.Field)
	require.Equal(t, "thor.png", serviceStub.lastCreateInput.FileName)
	require.Equal(t, "image/png", serviceStub.lastCreateInput.ContentType)
	require.JSONEq(t, `{
		"data": {
			"upload_url": "https://signed.example.com/upload",
			"method": "PUT",
			"headers": {"Content-Type": "image/png"},
			"object_key": "uploads/pets/image_url/2026/04/test.png",
			"public_url": "https://cdn.example.com/uploads/pets/image_url/2026/04/test.png",
			"expires_at": "2026-04-17T19:00:00Z"
		}
	}`, res.Body.String())
}

func TestUploadHandlerCreateIntentMapsErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handlerUnderTest := NewUploadHandler(&fakeUploadService{createIntentErr: apperror.ErrServiceUnavailable})

	router := gin.New()
	router.POST("/uploads/intents", handlerUnderTest.CreateIntent)

	req := httptest.NewRequest(http.MethodPost, "/uploads/intents", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusServiceUnavailable, res.Code)
	require.Contains(t, res.Body.String(), "create_upload_intent_failed")
}

func TestUploadHandlerComplete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serviceStub := &fakeUploadService{
		completeResult: service.CompleteUploadOutput{
			ObjectKey: "uploads/pets/image_url/2026/04/test.png",
			PublicURL: "https://cdn.example.com/uploads/pets/image_url/2026/04/test.png",
		},
	}
	handlerUnderTest := NewUploadHandler(serviceStub)

	router := gin.New()
	router.POST("/uploads/complete", handlerUnderTest.Complete)

	req := httptest.NewRequest(http.MethodPost, "/uploads/complete", bytes.NewBufferString(`{
		"resource":"pets",
		"field":"image_url",
		"object_key":"uploads/pets/image_url/2026/04/test.png"
	}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "uploads/pets/image_url/2026/04/test.png", serviceStub.lastCompleteInput.ObjectKey)
	require.JSONEq(t, `{
		"data": {
			"object_key":"uploads/pets/image_url/2026/04/test.png",
			"public_url":"https://cdn.example.com/uploads/pets/image_url/2026/04/test.png"
		}
	}`, res.Body.String())
}
