package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_HOST", "localhost")
	t.Setenv("REDIS_PORT", "6379")
	t.Setenv("WORKER_QUEUE", "")
	t.Setenv("WORKER_CONCURRENCY", "")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "localhost:6379", cfg.RedisAddr)
	require.Equal(t, "notifications", cfg.WorkerQueue)
	require.Equal(t, 5, cfg.Concurrency)
}

func TestLoadConfig_InvalidConcurrency(t *testing.T) {
	t.Setenv("WORKER_CONCURRENCY", "invalid")

	_, err := Load()
	require.Error(t, err)
}
