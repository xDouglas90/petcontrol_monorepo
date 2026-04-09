package processor

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type ScheduleConfirmationSender interface {
	SendScheduleConfirmation(ctx context.Context, payload queue.ScheduleConfirmationPayload) error
}

type ScheduleConfirmationProcessor struct {
	logger *slog.Logger
	sender ScheduleConfirmationSender
}

func NewScheduleConfirmationProcessor(logger *slog.Logger, sender ScheduleConfirmationSender) *ScheduleConfirmationProcessor {
	return &ScheduleConfirmationProcessor{logger: logger, sender: sender}
}

func (p *ScheduleConfirmationProcessor) HandleScheduleConfirmation(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.ParseScheduleConfirmationTask(task)
	if err != nil {
		p.logger.Error("worker task failed",
			"operation", "schedule_confirmation",
			"result", "invalid_payload",
			"error", err.Error(),
		)
		return err
	}

	p.logger.Info("worker task started",
		"operation", "schedule_confirmation",
		"schedule_id", payload.ScheduleID,
		"company_id", payload.CompanyID,
		"status", payload.Status,
		"version", payload.Version,
	)

	if err := p.sender.SendScheduleConfirmation(ctx, payload); err != nil {
		p.logger.Error("worker task failed",
			"operation", "schedule_confirmation",
			"schedule_id", payload.ScheduleID,
			"company_id", payload.CompanyID,
			"status", payload.Status,
			"result", "failed",
			"error", err.Error(),
		)
		return err
	}

	p.logger.Info("worker task completed",
		"operation", "schedule_confirmation",
		"schedule_id", payload.ScheduleID,
		"company_id", payload.CompanyID,
		"status", payload.Status,
		"result", "success",
	)

	return nil
}
