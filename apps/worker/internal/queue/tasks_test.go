package queue

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

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
