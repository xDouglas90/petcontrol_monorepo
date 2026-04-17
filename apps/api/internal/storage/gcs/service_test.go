package gcs

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/config"
)

type fakeSigner struct {
	lastBucketName  string
	lastObjectKey   string
	lastContentType string
	lastExpiresAt   time.Time
}

func (f *fakeSigner) SignedUploadURL(_ context.Context, bucketName string, objectKey string, contentType string, expiresAt time.Time) (string, http.Header, error) {
	f.lastBucketName = bucketName
	f.lastObjectKey = objectKey
	f.lastContentType = contentType
	f.lastExpiresAt = expiresAt

	headers := http.Header{}
	headers.Set("Content-Type", contentType)
	return "https://signed.example.com/upload", headers, nil
}

type fakeInspector struct {
	metadata ObjectMetadata
	err      error
}

func (f *fakeInspector) StatObject(_ context.Context, _ string, _ string) (ObjectMetadata, error) {
	return f.metadata, f.err
}

func TestServiceCreateIntentBuildsSignedUpload(t *testing.T) {
	signer := &fakeSigner{}
	service := NewService(config.UploadsConfig{
		GCSBucketName:      "petcontrol-assets",
		GCSUploadsBasePath: "tenant-assets",
		GCSSignedURLTTL:    15 * time.Minute,
		GCSPublicBaseURL:   "https://cdn.example.com/media",
	}, signer, nil)
	service.now = func() time.Time {
		return time.Date(2026, 4, 17, 18, 0, 0, 0, time.UTC)
	}
	service.newID = func() string {
		return "test-uuid"
	}

	intent, err := service.CreateIntent(context.Background(), CreateIntentInput{
		Resource:    "pets",
		Field:       "image_url",
		FileName:    "Thor Avatar.PNG",
		ContentType: "image/png",
		SizeBytes:   2048,
	})
	require.NoError(t, err)

	require.Equal(t, "petcontrol-assets", signer.lastBucketName)
	require.Equal(t, "tenant-assets/pets/image_url/2026/04/test-uuid-thor-avatar.png", signer.lastObjectKey)
	require.Equal(t, "image/png", signer.lastContentType)
	require.Equal(t, "https://signed.example.com/upload", intent.UploadURL)
	require.Equal(t, "PUT", intent.Method)
	require.Equal(t, "tenant-assets/pets/image_url/2026/04/test-uuid-thor-avatar.png", intent.ObjectKey)
	require.Equal(t, "https://cdn.example.com/media/tenant-assets/pets/image_url/2026/04/test-uuid-thor-avatar.png", intent.PublicURL)
	require.Equal(t, "image/png", intent.Headers.Get("Content-Type"))
	require.Equal(t, time.Date(2026, 4, 17, 18, 15, 0, 0, time.UTC), intent.ExpiresAt)
}

func TestServiceCreateIntentRejectsUnsupportedContentType(t *testing.T) {
	service := NewService(config.UploadsConfig{GCSBucketName: "petcontrol-assets"}, &fakeSigner{}, nil)

	_, err := service.CreateIntent(context.Background(), CreateIntentInput{
		Resource:    "pets",
		Field:       "image_url",
		FileName:    "thor.gif",
		ContentType: "image/gif",
		SizeBytes:   1024,
	})
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
}

func TestServiceCompleteRejectsWrongTargetPrefix(t *testing.T) {
	service := NewService(config.UploadsConfig{GCSBucketName: "petcontrol-assets", GCSUploadsBasePath: "tenant-assets"}, nil, &fakeInspector{
		metadata: ObjectMetadata{ContentType: "image/png", SizeBytes: 1024},
	})

	_, err := service.Complete(context.Background(), CompleteUploadInput{
		Resource:  "pets",
		Field:     "image_url",
		ObjectKey: "tenant-assets/services/image_url/2026/04/test.png",
	})
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
}

func TestServiceResolveObjectKeyChecksObjectMetadata(t *testing.T) {
	service := NewService(config.UploadsConfig{GCSBucketName: "petcontrol-assets", GCSUploadsBasePath: "tenant-assets"}, nil, &fakeInspector{
		metadata: ObjectMetadata{ContentType: "image/png", SizeBytes: 2048},
	})

	publicURL, err := service.ResolveObjectKey(context.Background(), "pets", "image_url", "tenant-assets/pets/image_url/2026/04/test.png")
	require.NoError(t, err)
	require.Equal(t, "https://storage.googleapis.com/petcontrol-assets/tenant-assets/pets/image_url/2026/04/test.png", publicURL)
}

func TestServiceResolveObjectKeyMapsMissingObjectToNotFound(t *testing.T) {
	service := NewService(config.UploadsConfig{GCSBucketName: "petcontrol-assets"}, nil, &fakeInspector{err: ErrObjectNotFound})

	_, err := service.ResolveObjectKey(context.Background(), "pets", "image_url", "pets/image_url/2026/04/test.png")
	require.ErrorIs(t, err, apperror.ErrNotFound)
}

func TestServiceResolveObjectKeyRejectsUnexpectedMetadata(t *testing.T) {
	service := NewService(config.UploadsConfig{GCSBucketName: "petcontrol-assets"}, nil, &fakeInspector{
		metadata: ObjectMetadata{ContentType: "application/pdf", SizeBytes: 1024},
	})

	_, err := service.ResolveObjectKey(context.Background(), "pets", "image_url", "pets/image_url/2026/04/test.pdf")
	require.ErrorIs(t, err, apperror.ErrUnprocessableEntity)
}

func TestServiceResolveObjectKeyReturnsInspectorError(t *testing.T) {
	service := NewService(config.UploadsConfig{GCSBucketName: "petcontrol-assets"}, nil, &fakeInspector{
		err: errors.New("boom"),
	})

	_, err := service.ResolveObjectKey(context.Background(), "pets", "image_url", "pets/image_url/2026/04/test.png")
	require.ErrorContains(t, err, "boom")
}
