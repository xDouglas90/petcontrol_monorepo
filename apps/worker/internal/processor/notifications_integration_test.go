package processor

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
)

type captureSender struct {
	ch chan queue.DummyNotificationPayload
}

func (s *captureSender) SendDummyNotification(_ context.Context, payload queue.DummyNotificationPayload) error {
	s.ch <- payload
	return nil
}

func TestNotificationsProcessor_ConsumesDummyTask(t *testing.T) {
	redisServer := miniredis.RunT(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &captureSender{ch: make(chan queue.DummyNotificationPayload, 1)}
	processor := NewNotificationsProcessor(logger, sender)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeNotificationDummy, processor.HandleDummyNotification)

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisServer.Addr()},
		asynq.Config{Concurrency: 1, Queues: map[string]int{"notifications": 1}},
	)
	go func() { _ = server.Run(mux) }()
	t.Cleanup(server.Shutdown)

	payload := queue.DummyNotificationPayload{
		CompanyID:  "11111111-1111-1111-1111-111111111111",
		UserID:     "22222222-2222-2222-2222-222222222222",
		Message:    "dummy",
		EnqueuedAt: time.Now().UTC(),
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	_, err = client.Enqueue(asynq.NewTask(queue.TypeNotificationDummy, body, asynq.Queue("notifications")))
	require.NoError(t, err)

	select {
	case received := <-sender.ch:
		require.Equal(t, payload.CompanyID, received.CompanyID)
		require.Equal(t, payload.UserID, received.UserID)
		require.Equal(t, payload.Message, received.Message)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for worker consumption")
	}
}
