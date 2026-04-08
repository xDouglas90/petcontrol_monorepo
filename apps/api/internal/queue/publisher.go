package queue

import (
	"context"

	"github.com/hibiken/asynq"
)

type Publisher interface {
	EnqueueDummyNotification(ctx context.Context, payload DummyNotificationPayload) error
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

func (p *AsynqPublisher) Close() error {
	return p.client.Close()
}
