package queue

import (
	"context"

	"github.com/hibiken/asynq"
)

type Publisher interface {
	EnqueueDummyNotification(ctx context.Context, payload DummyNotificationPayload) error
	EnqueueScheduleConfirmation(ctx context.Context, payload ScheduleConfirmationPayload) error
	EnqueuePersonAccessCredentials(ctx context.Context, payload PersonAccessCredentialsPayload) error
	Close() error
}

type AsynqPublisher struct {
	client    *asynq.Client
	queueName string
}

func NewAsynqPublisher(redisAddr string, queueName string) *AsynqPublisher {
	return &AsynqPublisher{
		client:    asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}),
		queueName: queueName,
	}
}

func (p *AsynqPublisher) EnqueueDummyNotification(ctx context.Context, payload DummyNotificationPayload) error {
	task, err := NewDummyNotificationTask(payload, p.queueName)
	if err != nil {
		return err
	}

	_, err = p.client.EnqueueContext(ctx, task)
	return err
}

func (p *AsynqPublisher) EnqueueScheduleConfirmation(ctx context.Context, payload ScheduleConfirmationPayload) error {
	task, err := NewScheduleConfirmationTask(payload, p.queueName)
	if err != nil {
		return err
	}

	_, err = p.client.EnqueueContext(ctx, task)
	return err
}

func (p *AsynqPublisher) EnqueuePersonAccessCredentials(ctx context.Context, payload PersonAccessCredentialsPayload) error {
	task, err := NewPersonAccessCredentialsTask(payload, p.queueName)
	if err != nil {
		return err
	}

	_, err = p.client.EnqueueContext(ctx, task)
	return err
}

func (p *AsynqPublisher) Close() error {
	return p.client.Close()
}
