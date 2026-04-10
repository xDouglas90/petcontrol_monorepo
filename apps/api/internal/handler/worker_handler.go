package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
)

type WorkerHandler struct {
	publisher queue.Publisher
}

func NewWorkerHandler(publisher queue.Publisher) *WorkerHandler {
	return &WorkerHandler{publisher: publisher}
}

type enqueueDummyRequest struct {
	Message string `json:"message"`
}

func (h *WorkerHandler) EnqueueDummyNotification(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims.UserID == "" {
		middleware.JSONError(c, 403, "user_context_required", "user context required")
		return
	}

	companyID, ok := middleware.GetCompanyID(c)
	if !ok || !companyID.Valid {
		middleware.JSONError(c, 403, "company_context_required", "company context required")
		return
	}

	var req enqueueDummyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.JSONError(c, 422, "invalid_request_body", "invalid request body")
		return
	}
	if req.Message == "" {
		req.Message = "Dummy notification from API"
	}

	payload := queue.DummyNotificationPayload{
		CompanyID:  claims.CompanyID,
		UserID:     claims.UserID,
		Message:    req.Message,
		EnqueuedAt: time.Now().UTC(),
	}

	if err := h.publisher.EnqueueDummyNotification(c.Request.Context(), payload); err != nil {
		middleware.JSONError(c, 500, "enqueue_task_failed", "failed to enqueue task")
		return
	}

	middleware.JSONData(c, 202, gin.H{"enqueued": true, "task": queue.TypeNotificationDummy})
}
