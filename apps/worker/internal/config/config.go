package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	RedisAddr           string
	WebhookAddr         string
	WhatsAppVerifyToken string
	WorkerQueue         string
	WorkerQueueType     string
	Concurrency         int
}

func Load() (Config, error) {
	cfg := Config{
		RedisAddr:           resolveRedisAddr(),
		WebhookAddr:         envWithDefault("WORKER_HTTP_ADDR", ":8091"),
		WhatsAppVerifyToken: os.Getenv("WHATSAPP_VERIFY_TOKEN"),
		WorkerQueue:         envWithDefault("WORKER_QUEUE", "notifications"),
		WorkerQueueType:     envWithDefault("WORKER_QUEUE_TYPE", "notifications:dummy"),
		Concurrency:         5,
	}

	if raw := envWithDefault("WORKER_CONCURRENCY", "5"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return Config{}, fmt.Errorf("invalid WORKER_CONCURRENCY")
		}
		cfg.Concurrency = value
	}

	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}

	return cfg, nil
}

func envWithDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func resolveRedisAddr() string {
	if addr := strings.TrimSpace(os.Getenv("REDIS_ADDR")); addr != "" {
		return addr
	}

	host := envWithDefault("REDIS_HOST", "localhost")
	port := envWithDefault("REDIS_PORT", "6379")
	return host + ":" + port
}
