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

type senderStub struct {
	payload queue.DummyNotificationPayload
	err     error
	called  bool
}

func (s *senderStub) SendDummyNotification(_ context.Context, payload queue.DummyNotificationPayload) error {
	s.called = true
	s.payload = payload
	return s.err
}

func TestNotificationsProcessor_HandleDummyNotification(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &senderStub{}
	p := NewNotificationsProcessor(logger, sender)

	task := asynq.NewTask(queue.TypeNotificationDummy, []byte(`{"company_id":"c-1","user_id":"u-1","message":"ok","enqueued_at":"2026-01-01T00:00:00Z"}`))

	err := p.HandleDummyNotification(context.Background(), task)
	require.NoError(t, err)
	require.True(t, sender.called)
	require.Equal(t, "c-1", sender.payload.CompanyID)
}

func TestNotificationsProcessor_HandleDummyNotification_InvalidPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &senderStub{}
	p := NewNotificationsProcessor(logger, sender)

	task := asynq.NewTask(queue.TypeNotificationDummy, []byte(`{`))

	err := p.HandleDummyNotification(context.Background(), task)
	require.Error(t, err)
	require.False(t, sender.called)
}

func TestNotificationsProcessor_HandleDummyNotification_SenderError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &senderStub{err: errors.New("failed send")}
	p := NewNotificationsProcessor(logger, sender)

	task := asynq.NewTask(queue.TypeNotificationDummy, []byte(`{"company_id":"c-1","user_id":"u-1","message":"ok","enqueued_at":"2026-01-01T00:00:00Z"}`))

	err := p.HandleDummyNotification(context.Background(), task)
	require.Error(t, err)
	require.True(t, sender.called)
}
