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

type captureScheduleConfirmationSender struct {
	ch chan queue.ScheduleConfirmationPayload
}

func (s *captureScheduleConfirmationSender) SendScheduleConfirmation(_ context.Context, payload queue.ScheduleConfirmationPayload) error {
	s.ch <- payload
	return nil
}

func TestScheduleConfirmationProcessor_ConsumesTask(t *testing.T) {
	redisServer := miniredis.RunT(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	sender := &captureScheduleConfirmationSender{ch: make(chan queue.ScheduleConfirmationPayload, 1)}
	processor := NewScheduleConfirmationProcessor(logger, sender)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeScheduleConfirmed, processor.HandleScheduleConfirmation)

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisServer.Addr()},
		asynq.Config{Concurrency: 1, Queues: map[string]int{"notifications": 1}},
	)
	go func() { _ = server.Run(mux) }()
	t.Cleanup(server.Shutdown)

	payload := queue.ScheduleConfirmationPayload{
		Version:    1,
		ScheduleID: "schedule-1",
		CompanyID:  "company-1",
		ChangedBy:  "user-1",
		Status:     "confirmed",
		OccurredAt: time.Now().UTC(),
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	_, err = client.Enqueue(asynq.NewTask(queue.TypeScheduleConfirmed, body, asynq.Queue("notifications")))
	require.NoError(t, err)

	select {
	case received := <-sender.ch:
		require.Equal(t, payload.ScheduleID, received.ScheduleID)
		require.Equal(t, payload.CompanyID, received.CompanyID)
		require.Equal(t, payload.Status, received.Status)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for worker consumption")
	}
}
