package processor

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type DummySender interface {
	SendDummyNotification(ctx context.Context, payload queue.DummyNotificationPayload) error
}

type NotificationsProcessor struct {
	logger *slog.Logger
	sender DummySender
}

func NewNotificationsProcessor(logger *slog.Logger, sender DummySender) *NotificationsProcessor {
	return &NotificationsProcessor{logger: logger, sender: sender}
}

func (p *NotificationsProcessor) HandleDummyNotification(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.ParseDummyNotificationTask(task)
	if err != nil {
		return err
	}

	p.logger.Info("processing dummy notification", "company_id", payload.CompanyID, "user_id", payload.UserID)
	return p.sender.SendDummyNotification(ctx, payload)
}
