package whatsapp

import (
	"context"
	"log/slog"

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
