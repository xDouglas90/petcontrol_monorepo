package config

import (
	"fmt"
	"os"
)

type Config struct {
	AppHost     string
	AppPort     string
	DatabaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		AppHost:     getEnv("API_HOST", "0.0.0.0"),
		AppPort:     getEnv("API_PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
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
