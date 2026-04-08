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

func TestAsynqPublisher_EnqueueDummyNotification(t *testing.T) {
	redisServer := miniredis.RunT(t)

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
