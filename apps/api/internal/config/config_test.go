package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadParsesUploadsConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/petcontrol")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("GCP_PROJECT_ID", "petcontrol-dev")
	t.Setenv("GCS_BUCKET_NAME", "petcontrol-assets")
	t.Setenv("GCS_UPLOADS_BASE_PATH", "/tenant-assets/")
	t.Setenv("GCS_SIGNED_URL_TTL_SECONDS", "900")
	t.Setenv("GCS_PUBLIC_BASE_URL", "https://cdn.example.com/media/")
	t.Setenv("GCS_SIGNER_SERVICE_ACCOUNT_EMAIL", "uploads@example.iam.gserviceaccount.com")
	t.Setenv("GCS_SIGNER_PRIVATE_KEY", "line-1\\nline-2")

	cfg, err := Load()
	require.NoError(t, err)

	require.Equal(t, "petcontrol-dev", cfg.Uploads.GCPProjectID)
	require.Equal(t, "petcontrol-assets", cfg.Uploads.GCSBucketName)
	require.Equal(t, "tenant-assets", cfg.Uploads.GCSUploadsBasePath)
	require.Equal(t, 15*time.Minute, cfg.Uploads.GCSSignedURLTTL)
	require.Equal(t, "https://cdn.example.com/media", cfg.Uploads.GCSPublicBaseURL)
	require.Equal(t, "uploads@example.iam.gserviceaccount.com", cfg.Uploads.GCSSignerServiceAccount)
	require.Equal(t, "line-1\nline-2", cfg.Uploads.GCSSignerPrivateKey)
}

func TestLoadRejectsInvalidSignedURLTTL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/petcontrol")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("GCS_SIGNED_URL_TTL_SECONDS", "invalid")

	_, err := Load()
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid GCS_SIGNED_URL_TTL_SECONDS")
}
