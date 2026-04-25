package queue

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func TestParseDummyNotificationTask(t *testing.T) {
	task := asynq.NewTask(TypeNotificationDummy, []byte(`{"company_id":"company-1","user_id":"user-1","message":"hello","enqueued_at":"2026-01-01T00:00:00Z"}`))

	payload, err := ParseDummyNotificationTask(task)
	require.NoError(t, err)
	require.Equal(t, "company-1", payload.CompanyID)
	require.Equal(t, "user-1", payload.UserID)
	require.Equal(t, "hello", payload.Message)
	require.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), payload.EnqueuedAt)
}

func TestParseDummyNotificationTask_UnexpectedType(t *testing.T) {
	task := asynq.NewTask(TypeScheduleConfirmed, []byte(`{}`))

	_, err := ParseDummyNotificationTask(task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected task type")
}

func TestParseDummyNotificationTask_InvalidJSON(t *testing.T) {
	task := asynq.NewTask(TypeNotificationDummy, []byte(`{`))

	_, err := ParseDummyNotificationTask(task)
	require.Error(t, err)
}

func TestParseScheduleConfirmationTask(t *testing.T) {
	task := asynq.NewTask(TypeScheduleConfirmed, []byte(`{"version":1,"schedule_id":"schedule-1","company_id":"company-1","changed_by":"user-1","status":"confirmed","status_notes":"ok","occurred_at":"2026-01-01T00:00:00Z"}`))

	payload, err := ParseScheduleConfirmationTask(task)
	require.NoError(t, err)
	require.Equal(t, 1, payload.Version)
	require.Equal(t, "schedule-1", payload.ScheduleID)
	require.Equal(t, "company-1", payload.CompanyID)
	require.Equal(t, "user-1", payload.ChangedBy)
	require.Equal(t, "confirmed", payload.Status)
	require.Equal(t, "ok", payload.StatusNotes)
	require.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), payload.OccurredAt)
}

func TestParseScheduleConfirmationTask_InvalidPayload(t *testing.T) {
	task := asynq.NewTask(TypeScheduleConfirmed, []byte(`{"version":1}`))

	_, err := ParseScheduleConfirmationTask(task)
	require.Error(t, err)
}

func TestParseScheduleConfirmationTask_Version2Context(t *testing.T) {
	task := asynq.NewTask(TypeScheduleConfirmed, []byte(`{"version":2,"schedule_id":"schedule-1","company_id":"company-1","changed_by":"user-1","client_name":"Maria Silva","pet_name":"Thor","service_titles":["Banho completo"],"status":"confirmed","status_notes":"ok","occurred_at":"2026-01-01T00:00:00Z"}`))

	payload, err := ParseScheduleConfirmationTask(task)
	require.NoError(t, err)
	require.Equal(t, 2, payload.Version)
	require.Equal(t, "Maria Silva", payload.ClientName)
	require.Equal(t, "Thor", payload.PetName)
	require.Equal(t, []string{"Banho completo"}, payload.ServiceTitles)
}

func TestParseScheduleConfirmationTask_UnexpectedType(t *testing.T) {
	task := asynq.NewTask(TypeNotificationDummy, []byte(`{}`))

	_, err := ParseScheduleConfirmationTask(task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected task type")
}

func TestParseScheduleConfirmationTask_MissingVersion(t *testing.T) {
	task := asynq.NewTask(TypeScheduleConfirmed, []byte(`{"schedule_id":"schedule-1","company_id":"company-1","changed_by":"user-1","status":"confirmed","occurred_at":"2026-01-01T00:00:00Z"}`))

	_, err := ParseScheduleConfirmationTask(task)
	require.Error(t, err)
	require.EqualError(t, err, "missing payload version")
}

func TestParseScheduleConfirmationTask_InvalidJSON(t *testing.T) {
	task := asynq.NewTask(TypeScheduleConfirmed, []byte(`{`))

	_, err := ParseScheduleConfirmationTask(task)
	require.Error(t, err)
}

func TestParsePersonAccessCredentialsTask(t *testing.T) {
	task := asynq.NewTask(TypePersonAccessCredentials, []byte(`{"version":1,"company_id":"company-1","person_id":"person-1","user_id":"user-1","recipient_name":"Maria Silva","recipient_email":"maria@example.com","temporary_password":"secret","system_url":"https://app.example.com","role":"admin","occurred_at":"2026-01-01T00:00:00Z"}`))

	payload, err := ParsePersonAccessCredentialsTask(task)
	require.NoError(t, err)
	require.Equal(t, 1, payload.Version)
	require.Equal(t, "company-1", payload.CompanyID)
	require.Equal(t, "person-1", payload.PersonID)
	require.Equal(t, "user-1", payload.UserID)
	require.Equal(t, "Maria Silva", payload.RecipientName)
	require.Equal(t, "maria@example.com", payload.RecipientEmail)
	require.Equal(t, "secret", payload.TemporaryPassword)
	require.Equal(t, "https://app.example.com", payload.SystemURL)
	require.Equal(t, "admin", payload.Role)
	require.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), payload.OccurredAt)
}

func TestParsePersonAccessCredentialsTask_UnexpectedType(t *testing.T) {
	task := asynq.NewTask(TypeNotificationDummy, []byte(`{}`))

	_, err := ParsePersonAccessCredentialsTask(task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected task type")
}

func TestParsePersonAccessCredentialsTask_InvalidJSON(t *testing.T) {
	task := asynq.NewTask(TypePersonAccessCredentials, []byte(`{`))

	_, err := ParsePersonAccessCredentialsTask(task)
	require.Error(t, err)
}

func TestParsePersonAccessCredentialsTask_MissingVersion(t *testing.T) {
	task := asynq.NewTask(TypePersonAccessCredentials, []byte(`{"company_id":"company-1","person_id":"person-1","user_id":"user-1","recipient_email":"maria@example.com","temporary_password":"secret","occurred_at":"2026-01-01T00:00:00Z"}`))

	_, err := ParsePersonAccessCredentialsTask(task)
	require.Error(t, err)
	require.EqualError(t, err, "missing payload version")
}

func TestParsePersonAccessCredentialsTask_InvalidPayload(t *testing.T) {
	payload := `{"version":1,"company_id":"company-1","person_id":"person-1","user_id":"user-1","recipient_email":"","temporary_password":"secret","occurred_at":"2026-01-01T00:00:00Z"}`
	task := asynq.NewTask(TypePersonAccessCredentials, []byte(payload))

	_, err := ParsePersonAccessCredentialsTask(task)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid person access credentials payload")
}
