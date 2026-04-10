package whatsapp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type Client struct {
	logger *slog.Logger
}

func NewClient(logger *slog.Logger) *Client {
	return &Client{logger: logger}
}

func (c *Client) SendDummyNotification(_ context.Context, payload queue.DummyNotificationPayload) error {
	c.logger.Info("dummy notification sent", "company_id", payload.CompanyID, "user_id", payload.UserID, "message", payload.Message)
	return nil
}

func (c *Client) SendScheduleConfirmation(_ context.Context, payload queue.ScheduleConfirmationPayload) error {
	c.logger.Info("schedule confirmation sent",
		"schedule_id", payload.ScheduleID,
		"company_id", payload.CompanyID,
		"changed_by", payload.ChangedBy,
		"status", payload.Status,
		"version", payload.Version,
	)
	return nil
}

func NewWebhookHandler(verifyToken string, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/whatsapp", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			challenge := r.URL.Query().Get("hub.challenge")
			token := r.URL.Query().Get("hub.verify_token")
			if verifyToken == "" || token != verifyToken {
				logger.Warn("whatsapp webhook verification failed", "result", "forbidden")
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			logger.Info("whatsapp webhook verified")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, challenge)
		case http.MethodPost:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "invalid body", http.StatusBadRequest)
				return
			}
			logger.Info("whatsapp webhook received", "bytes", len(body))
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, "ok")
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}
