package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

func mustStartMiniRedis(t *testing.T) *miniredis.Miniredis {
	t.Helper()

	srv, err := miniredis.Run()
	if err != nil {
		// Some sandboxed environments disallow binding to local TCP ports.
		t.Skipf("skipping queue integration test; cannot start miniredis: %v", err)
	}
	t.Cleanup(srv.Close)
	return srv
}

func TestAsynqPublisher_EnqueueDummyNotification(t *testing.T) {
	redisServer := mustStartMiniRedis(t)

	processed := make(chan DummyNotificationPayload, 1)
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeNotificationDummy, func(_ context.Context, task *asynq.Task) error {
		var payload DummyNotificationPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return err
		}
		processed <- payload
		return nil
	})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisServer.Addr()},
		asynq.Config{Concurrency: 1, Queues: map[string]int{"notifications": 1}},
	)
	go func() {
		_ = server.Run(mux)
	}()
	t.Cleanup(server.Shutdown)

	publisher := NewAsynqPublisher(redisServer.Addr(), "notifications")
	t.Cleanup(func() {
		require.NoError(t, publisher.Close())
	})

	err := publisher.EnqueueDummyNotification(context.Background(), DummyNotificationPayload{
		CompanyID:  "11111111-1111-1111-1111-111111111111",
		UserID:     "22222222-2222-2222-2222-222222222222",
		Message:    "dummy",
		EnqueuedAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	select {
	case payload := <-processed:
		require.Equal(t, "dummy", payload.Message)
		require.Equal(t, "11111111-1111-1111-1111-111111111111", payload.CompanyID)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for task consumption")
	}
}

func TestAsynqPublisher_EnqueueScheduleConfirmation(t *testing.T) {
	redisServer := mustStartMiniRedis(t)

	processed := make(chan ScheduleConfirmationPayload, 1)
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeScheduleConfirmed, func(_ context.Context, task *asynq.Task) error {
		var payload ScheduleConfirmationPayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			return err
		}
		processed <- payload
		return nil
	})

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisServer.Addr()},
		asynq.Config{Concurrency: 1, Queues: map[string]int{"notifications": 1}},
	)
	go func() {
		_ = server.Run(mux)
	}()
	t.Cleanup(server.Shutdown)

	publisher := NewAsynqPublisher(redisServer.Addr(), "notifications")
	t.Cleanup(func() {
		require.NoError(t, publisher.Close())
	})

	err := publisher.EnqueueScheduleConfirmation(context.Background(), ScheduleConfirmationPayload{
		Version:    1,
		ScheduleID: "33333333-3333-3333-3333-333333333333",
		CompanyID:  "11111111-1111-1111-1111-111111111111",
		ChangedBy:  "22222222-2222-2222-2222-222222222222",
		Status:     "confirmed",
		OccurredAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	select {
	case payload := <-processed:
		require.Equal(t, "confirmed", payload.Status)
		require.Equal(t, "33333333-3333-3333-3333-333333333333", payload.ScheduleID)
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for task consumption")
	}
}
