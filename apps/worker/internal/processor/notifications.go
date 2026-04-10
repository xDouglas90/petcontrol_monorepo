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
		p.logger.Error("worker task failed",
			"operation", "dummy_notification",
			"result", "invalid_payload",
			"error", err.Error(),
		)
		return err
	}

	p.logger.Info("worker task started",
		"operation", "dummy_notification",
		"company_id", payload.CompanyID,
		"user_id", payload.UserID,
	)

	if err := p.sender.SendDummyNotification(ctx, payload); err != nil {
		p.logger.Error("worker task failed",
			"operation", "dummy_notification",
			"company_id", payload.CompanyID,
			"user_id", payload.UserID,
			"result", "failed",
			"error", err.Error(),
		)
		return err
	}

	p.logger.Info("worker task completed",
		"operation", "dummy_notification",
		"company_id", payload.CompanyID,
		"user_id", payload.UserID,
		"result", "success",
	)

	return nil
}
