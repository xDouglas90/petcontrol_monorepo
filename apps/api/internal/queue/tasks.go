package queue

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeNotificationDummy       = "notifications:dummy"
	TypeScheduleConfirmed       = "schedules:confirmed"
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

func NewScheduleConfirmationTask(payload ScheduleConfirmationPayload, queueName string) (*asynq.Task, error) {
	if payload.Version == 0 {
		payload.Version = 1
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TypeScheduleConfirmed,
		body,
		asynq.Queue(queueName),
		asynq.MaxRetry(5),
		asynq.Timeout(45*time.Second),
	), nil
}

func NewPersonAccessCredentialsTask(payload PersonAccessCredentialsPayload, queueName string) (*asynq.Task, error) {
	if payload.Version == 0 {
		payload.Version = 1
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(
		TypePersonAccessCredentials,
		body,
		asynq.Queue(queueName),
		asynq.MaxRetry(5),
		asynq.Timeout(45*time.Second),
	), nil
}
