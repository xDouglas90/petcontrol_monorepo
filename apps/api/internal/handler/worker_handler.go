package handler

import (
	"net/http"
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
		c.JSON(http.StatusForbidden, gin.H{"error": "user context required"})
		return
	}

	companyID, ok := middleware.GetCompanyID(c)
	if !ok || !companyID.Valid {
		c.JSON(http.StatusForbidden, gin.H{"error": "company context required"})
		return
	}

	var req enqueueDummyRequest
	_ = c.ShouldBindJSON(&req)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"data": gin.H{"enqueued": true, "task": queue.TypeNotificationDummy}})
}
