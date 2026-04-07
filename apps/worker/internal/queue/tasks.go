package queue

import (
	"encoding/json"
	"fmt"
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

func ParseDummyNotificationTask(task *asynq.Task) (DummyNotificationPayload, error) {
	if task.Type() != TypeNotificationDummy {
		return DummyNotificationPayload{}, fmt.Errorf("unexpected task type: %s", task.Type())
	}

	var payload DummyNotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return DummyNotificationPayload{}, err
	}

	return payload, nil
}
