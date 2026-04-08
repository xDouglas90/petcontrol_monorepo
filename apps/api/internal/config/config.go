package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppHost     string
	AppPort     string
	DatabaseURL string
	RedisAddr   string
	WorkerQueue string
	JWTSecret   string
	JWTTTL      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppHost:     getEnv("API_HOST", "0.0.0.0"),
		AppPort:     getEnv("API_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisAddr:   resolveRedisAddr(),
		WorkerQueue: getEnv("WORKER_QUEUE", "notifications"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:      30 * time.Minute,
	}

	rawTTL := firstNonEmptyEnv("JWT_ACCESS_TOKEN_TTL", "JWT_TTL")
	if rawTTL == "" {
		rawTTL = "30m"
	}

	ttl, err := time.ParseDuration(rawTTL)
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_ACCESS_TOKEN_TTL/JWT_TTL: %w", err)
	}
	cfg.JWTTTL = ttl

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func (c Config) Address() string {
	return fmt.Sprintf("%s:%s", c.AppHost, c.AppPort)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
	}
	return ""
}

func resolveRedisAddr() string {
	if addr := strings.TrimSpace(os.Getenv("REDIS_ADDR")); addr != "" {
		return addr
	}

	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	return host + ":" + port
}
