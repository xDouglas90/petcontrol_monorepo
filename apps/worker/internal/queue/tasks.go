package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeScheduleConfirmed = "schedules:confirmed"
	TypeNotificationDummy = "notifications:dummy"
)

type DummyNotificationPayload struct {
	CompanyID  string    `json:"company_id"`
	UserID     string    `json:"user_id"`
	Message    string    `json:"message"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

type ScheduleConfirmationPayload struct {
	Version     int       `json:"version"`
	ScheduleID  string    `json:"schedule_id"`
	CompanyID   string    `json:"company_id"`
	ChangedBy   string    `json:"changed_by"`
	Status      string    `json:"status"`
	StatusNotes string    `json:"status_notes"`
	OccurredAt  time.Time `json:"occurred_at"`
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

func ParseScheduleConfirmationTask(task *asynq.Task) (ScheduleConfirmationPayload, error) {
	if task.Type() != TypeScheduleConfirmed {
		return ScheduleConfirmationPayload{}, fmt.Errorf("unexpected task type: %s", task.Type())
	}

	var payload ScheduleConfirmationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return ScheduleConfirmationPayload{}, err
	}
	if payload.Version == 0 {
		return ScheduleConfirmationPayload{}, fmt.Errorf("missing payload version")
	}
	if payload.ScheduleID == "" || payload.CompanyID == "" || payload.ChangedBy == "" || payload.Status == "" || payload.OccurredAt.IsZero() {
		return ScheduleConfirmationPayload{}, fmt.Errorf("invalid schedule confirmation payload")
	}

	return payload, nil
}
