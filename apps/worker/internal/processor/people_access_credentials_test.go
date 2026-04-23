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

type personAccessSenderStub struct {
	payload queue.PersonAccessCredentialsPayload
	err     error
	called  bool
}

func (s *personAccessSenderStub) SendPersonAccessCredentials(_ context.Context, payload queue.PersonAccessCredentialsPayload) error {
	s.called = true
	s.payload = payload
	return s.err
}

func TestPersonAccessCredentialsProcessor_HandlePersonAccessCredentials(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &personAccessSenderStub{}
	p := NewPersonAccessCredentialsProcessor(logger, sender)

	task := asynq.NewTask(queue.TypePersonAccessCredentials, []byte(`{"version":1,"company_id":"c-1","person_id":"p-1","user_id":"u-1","recipient_name":"Maria","recipient_email":"maria@petcontrol.local","temporary_password":"abc123","role":"system","occurred_at":"2026-01-01T00:00:00Z"}`))

	err := p.HandlePersonAccessCredentials(context.Background(), task)
	require.NoError(t, err)
	require.True(t, sender.called)
	require.Equal(t, "u-1", sender.payload.UserID)
	require.Equal(t, "maria@petcontrol.local", sender.payload.RecipientEmail)
}

func TestPersonAccessCredentialsProcessor_HandlePersonAccessCredentials_InvalidPayload(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &personAccessSenderStub{}
	p := NewPersonAccessCredentialsProcessor(logger, sender)

	task := asynq.NewTask(queue.TypePersonAccessCredentials, []byte(`{`))

	err := p.HandlePersonAccessCredentials(context.Background(), task)
	require.Error(t, err)
	require.False(t, sender.called)
}

func TestPersonAccessCredentialsProcessor_HandlePersonAccessCredentials_SenderError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &personAccessSenderStub{err: errors.New("send failed")}
	p := NewPersonAccessCredentialsProcessor(logger, sender)

	task := asynq.NewTask(queue.TypePersonAccessCredentials, []byte(`{"version":1,"company_id":"c-1","person_id":"p-1","user_id":"u-1","recipient_name":"Maria","recipient_email":"maria@petcontrol.local","temporary_password":"abc123","role":"system","occurred_at":"2026-01-01T00:00:00Z"}`))

	err := p.HandlePersonAccessCredentials(context.Background(), task)
	require.Error(t, err)
	require.True(t, sender.called)
}
