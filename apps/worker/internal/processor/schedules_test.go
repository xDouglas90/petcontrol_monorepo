package processor

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type scheduleConfirmationSenderStub struct {
	payload queue.ScheduleConfirmationPayload
	err     error
	called  bool
}

func (s *scheduleConfirmationSenderStub) SendScheduleConfirmation(_ context.Context, payload queue.ScheduleConfirmationPayload) error {
	s.called = true
	s.payload = payload
	return s.err
}

func TestScheduleConfirmationProcessor_HandleScheduleConfirmation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &scheduleConfirmationSenderStub{}
	p := NewScheduleConfirmationProcessor(logger, sender)

	task := asynq.NewTask(queue.TypeScheduleConfirmed, []byte(`{"version":1,"schedule_id":"schedule-1","company_id":"company-1","changed_by":"user-1","status":"confirmed","status_notes":"ok","occurred_at":"2026-01-01T00:00:00Z"}`))

	err := p.HandleScheduleConfirmation(context.Background(), task)
	require.NoError(t, err)
	require.True(t, sender.called)
	require.Equal(t, "schedule-1", sender.payload.ScheduleID)
	require.Equal(t, "confirmed", sender.payload.Status)
}

func TestScheduleConfirmationProcessor_HandleScheduleConfirmation_InvalidPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &scheduleConfirmationSenderStub{}
	p := NewScheduleConfirmationProcessor(logger, sender)

	task := asynq.NewTask(queue.TypeScheduleConfirmed, []byte(`{`))

	err := p.HandleScheduleConfirmation(context.Background(), task)
	require.Error(t, err)
	require.False(t, sender.called)
}

func TestScheduleConfirmationProcessor_HandleScheduleConfirmation_SenderError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &scheduleConfirmationSenderStub{err: errors.New("failed send")}
	p := NewScheduleConfirmationProcessor(logger, sender)

	task := asynq.NewTask(queue.TypeScheduleConfirmed, []byte(`{"version":1,"schedule_id":"schedule-1","company_id":"company-1","changed_by":"user-1","status":"confirmed","status_notes":"ok","occurred_at":"2026-01-01T00:00:00Z"}`))

	err := p.HandleScheduleConfirmation(context.Background(), task)
	require.Error(t, err)
	require.True(t, sender.called)
}
