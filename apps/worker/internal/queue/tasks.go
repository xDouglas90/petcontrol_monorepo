package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeScheduleConfirmed       = "schedules:confirmed"
	TypeNotificationDummy       = "notifications:dummy"
	TypePersonAccessCredentials = "people:access_credentials"
)

type DummyNotificationPayload struct {
	CompanyID  string    `json:"company_id"`
	UserID     string    `json:"user_id"`
	Message    string    `json:"message"`
	EnqueuedAt time.Time `json:"enqueued_at"`
}

type ScheduleConfirmationPayload struct {
	Version       int       `json:"version"`
	ScheduleID    string    `json:"schedule_id"`
	CompanyID     string    `json:"company_id"`
	ChangedBy     string    `json:"changed_by"`
	ClientName    string    `json:"client_name,omitempty"`
	PetName       string    `json:"pet_name,omitempty"`
	ServiceTitles []string  `json:"service_titles,omitempty"`
	Status        string    `json:"status"`
	StatusNotes   string    `json:"status_notes"`
	OccurredAt    time.Time `json:"occurred_at"`
}

type PersonAccessCredentialsPayload struct {
	Version           int       `json:"version"`
	CompanyID         string    `json:"company_id"`
	PersonID          string    `json:"person_id"`
	UserID            string    `json:"user_id"`
	RecipientName     string    `json:"recipient_name"`
	RecipientEmail    string    `json:"recipient_email"`
	TemporaryPassword string    `json:"temporary_password"`
	SystemURL         string    `json:"system_url"`
	Role              string    `json:"role"`
	OccurredAt        time.Time `json:"occurred_at"`
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

func ParsePersonAccessCredentialsTask(task *asynq.Task) (PersonAccessCredentialsPayload, error) {
	if task.Type() != TypePersonAccessCredentials {
		return PersonAccessCredentialsPayload{}, fmt.Errorf("unexpected task type: %s", task.Type())
	}

	var payload PersonAccessCredentialsPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return PersonAccessCredentialsPayload{}, err
	}
	if payload.Version == 0 {
		return PersonAccessCredentialsPayload{}, fmt.Errorf("missing payload version")
	}
	if payload.CompanyID == "" || payload.PersonID == "" || payload.UserID == "" || payload.RecipientEmail == "" || payload.TemporaryPassword == "" || payload.OccurredAt.IsZero() {
		return PersonAccessCredentialsPayload{}, fmt.Errorf("invalid person access credentials payload")
	}

	return payload, nil
}
