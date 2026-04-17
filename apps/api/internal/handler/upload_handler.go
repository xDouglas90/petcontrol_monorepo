package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

type uploadIntentCreator interface {
	CreateIntent(ctx context.Context, input service.CreateUploadIntentInput) (service.CreateUploadIntentOutput, error)
	Complete(ctx context.Context, input service.CompleteUploadInput) (service.CompleteUploadOutput, error)
}

type UploadHandler struct {
	service uploadIntentCreator
}

type createUploadIntentRequest struct {
	Resource    string `json:"resource"`
	Field       string `json:"field"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type completeUploadRequest struct {
	Resource  string `json:"resource"`
	Field     string `json:"field"`
	ObjectKey string `json:"object_key"`
}

type uploadIntentResponse struct {
	UploadURL string            `json:"upload_url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	ObjectKey string            `json:"object_key"`
	PublicURL string            `json:"public_url"`
	ExpiresAt string            `json:"expires_at"`
}

type completeUploadResponse struct {
	ObjectKey string `json:"object_key"`
	PublicURL string `json:"public_url"`
}

func NewUploadHandler(service uploadIntentCreator) *UploadHandler {
	return &UploadHandler{service: service}
}

func (h *UploadHandler) CreateIntent(c *gin.Context) {
	var req createUploadIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	result, err := h.service.CreateIntent(c.Request.Context(), service.CreateUploadIntentInput{
		Resource:    strings.TrimSpace(req.Resource),
		Field:       strings.TrimSpace(req.Field),
		FileName:    strings.TrimSpace(req.FileName),
		ContentType: strings.TrimSpace(req.ContentType),
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "create_upload_intent_failed", "failed to create upload intent")
		return
	}

	middleware.JSONData(c, http.StatusCreated, uploadIntentResponse{
		UploadURL: result.UploadURL,
		Method:    result.Method,
		Headers:   flattenHeaders(result.Headers),
		ObjectKey: result.ObjectKey,
		PublicURL: result.PublicURL,
		ExpiresAt: result.ExpiresAt.UTC().Format(time.RFC3339),
	})
}

func (h *UploadHandler) Complete(c *gin.Context) {
	var req completeUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, http.StatusUnprocessableEntity, "invalid_request_body", "invalid request body")
		return
	}

	result, err := h.service.Complete(c.Request.Context(), service.CompleteUploadInput{
		Resource:  strings.TrimSpace(req.Resource),
		Field:     strings.TrimSpace(req.Field),
		ObjectKey: strings.TrimSpace(req.ObjectKey),
	})
	if err != nil {
		middleware.JSONError(c, apperror.HTTPStatus(err), "complete_upload_failed", "failed to complete upload")
		return
	}

	middleware.JSONData(c, http.StatusOK, completeUploadResponse{
		ObjectKey: result.ObjectKey,
		PublicURL: result.PublicURL,
	})
}

func flattenHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string, len(headers))
	for key, values := range headers {
		out[key] = strings.Join(values, ", ")
	}
	return out
}
