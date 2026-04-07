package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppHost     string
	AppPort     string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppHost:     getEnv("API_HOST", "0.0.0.0"),
		AppPort:     getEnv("API_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
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
