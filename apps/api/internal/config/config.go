package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppHost            string
	AppPort            string
	CORSAllowedOrigins []string
	DatabaseURL        string
	RedisAddr          string
	WorkerQueue        string
	JWTSecret          string
	JWTTTL             time.Duration
	Uploads            UploadsConfig
}

type UploadsConfig struct {
	GCPProjectID            string
	GCSBucketName           string
	GCSUploadsBasePath      string
	GCSSignedURLTTL         time.Duration
	GCSPublicBaseURL        string
	GCSSignerServiceAccount string
	GCSSignerPrivateKey     string
	GCSCredentialsFile      string
}

func Load() (Config, error) {
	cfg := Config{
		AppHost:            getEnv("API_HOST", "0.0.0.0"),
		AppPort:            getEnv("API_PORT", "8080"),
		CORSAllowedOrigins: resolveAllowedOrigins(),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		RedisAddr:          resolveRedisAddr(),
		WorkerQueue:        getEnv("WORKER_QUEUE", "notifications"),
		JWTSecret:          getEnv("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:             30 * time.Minute,
		Uploads: UploadsConfig{
			GCPProjectID:            strings.TrimSpace(firstNonEmptyEnv("GCS_PROJECT_ID", "GCP_PROJECT_ID")),
			GCSBucketName:           strings.TrimSpace(firstNonEmptyEnv("GCS_BUCKET", "GCS_BUCKET_NAME")),
			GCSUploadsBasePath:      strings.Trim(strings.TrimSpace(getEnv("GCS_UPLOADS_BASE_PATH", "uploads")), "/"),
			GCSSignedURLTTL:         15 * time.Minute,
			GCSPublicBaseURL:        strings.TrimRight(strings.TrimSpace(os.Getenv("GCS_PUBLIC_BASE_URL")), "/"),
			GCSSignerServiceAccount: strings.TrimSpace(os.Getenv("GCS_SIGNER_SERVICE_ACCOUNT_EMAIL")),
			GCSSignerPrivateKey:     decodePrivateKeyEnv(os.Getenv("GCS_SIGNER_PRIVATE_KEY")),
			GCSCredentialsFile:      resolveFilePath(os.Getenv("GCS_CREDENTIALS_FILE")),
		},
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

	rawSignedURLTTL := strings.TrimSpace(os.Getenv("GCS_SIGNED_URL_TTL_SECONDS"))
	if rawSignedURLTTL != "" {
		seconds, err := time.ParseDuration(rawSignedURLTTL + "s")
		if err != nil {
			return Config{}, fmt.Errorf("invalid GCS_SIGNED_URL_TTL_SECONDS: %w", err)
		}
		cfg.Uploads.GCSSignedURLTTL = seconds
	}

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

func resolveAllowedOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if raw == "" {
		return []string{"http://localhost:5173"}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{"http://localhost:5173"}
	}

	return origins
}

func decodePrivateKeyEnv(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.ReplaceAll(trimmed, `\n`, "\n")
}

func resolveFilePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}

	// Se o caminho for absoluto, retorna como está
	if strings.HasPrefix(trimmed, "/") {
		return trimmed
	}

	// Tenta o caminho relativo ao CWD
	if _, err := os.Stat(trimmed); err == nil {
		return trimmed
	}

	// Tenta subir níveis para encontrar (caso estejamos em apps/api)
	// Como estamos em um monorepo, é comum estarmos em subpastas
	levelsUp := []string{"../", "../../", "../../../"}
	for _, up := range levelsUp {
		testPath := up + trimmed
		if _, err := os.Stat(testPath); err == nil {
			return testPath
		}
	}

	return trimmed
}
