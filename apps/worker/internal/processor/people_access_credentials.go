package processor

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type PersonAccessCredentialsSender interface {
	SendPersonAccessCredentials(ctx context.Context, payload queue.PersonAccessCredentialsPayload) error
}

type PersonAccessCredentialsProcessor struct {
	logger *slog.Logger
	sender PersonAccessCredentialsSender
}

func NewPersonAccessCredentialsProcessor(logger *slog.Logger, sender PersonAccessCredentialsSender) *PersonAccessCredentialsProcessor {
	return &PersonAccessCredentialsProcessor{logger: logger, sender: sender}
}

func (p *PersonAccessCredentialsProcessor) HandlePersonAccessCredentials(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.ParsePersonAccessCredentialsTask(task)
	if err != nil {
		p.logger.Error("worker task failed",
			"operation", "person_access_credentials",
			"result", "invalid_payload",
			"error", err.Error(),
		)
		return err
	}

	p.logger.Info("worker task started",
		"operation", "person_access_credentials",
		"user_id", payload.UserID,
		"person_id", payload.PersonID,
		"recipient_email", payload.RecipientEmail,
	)

	if err := p.sender.SendPersonAccessCredentials(ctx, payload); err != nil {
		p.logger.Error("worker task failed",
			"operation", "person_access_credentials",
			"user_id", payload.UserID,
			"person_id", payload.PersonID,
			"recipient_email", payload.RecipientEmail,
			"result", "failed",
			"error", err.Error(),
		)
		return err
	}

	p.logger.Info("worker task completed",
		"operation", "person_access_credentials",
		"user_id", payload.UserID,
		"person_id", payload.PersonID,
		"recipient_email", payload.RecipientEmail,
		"result", "success",
	)

	return nil
}
