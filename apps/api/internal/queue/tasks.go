package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const TypeNotificationDummy = "notifications:dummy"

type DummyNotificationPayload struct {
	CompanyID  string    `json:"company_id"`
	UserID     string    `json:"user_id"`
	Message    string    `json:"message"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

func NewDummyNotificationTask(payload DummyNotificationPayload, queueName string) (*asynq.Task, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TypeNotificationDummy,
		body,
		asynq.Queue(queueName),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
	), nil
}
