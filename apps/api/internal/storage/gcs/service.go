package gcs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/config"
)

type SignedUploadURLProvider interface {
	SignedUploadURL(ctx context.Context, bucketName string, objectKey string, contentType string, expiresAt time.Time) (string, http.Header, error)
}

type ObjectInspector interface {
	StatObject(ctx context.Context, bucketName string, objectKey string) (ObjectMetadata, error)
}

type ObjectMetadata struct {
	ContentType string
	SizeBytes   int64
}

type Service struct {
	cfg       config.UploadsConfig
	signer    SignedUploadURLProvider
	inspector ObjectInspector
	now       func() time.Time
	newID     func() string
}

type CreateIntentInput struct {
	Resource    string
	Field       string
	FileName    string
	ContentType string
	SizeBytes   int64
}

type CreateIntentOutput struct {
	UploadURL string
	Method    string
	Headers   http.Header
	ObjectKey string
	PublicURL string
	ExpiresAt time.Time
}

type CompleteUploadInput struct {
	Resource  string
	Field     string
	ObjectKey string
}

type CompleteUploadOutput struct {
	ObjectKey string
	PublicURL string
}

type targetSpec struct {
	MaxSizeBytes int64
	ContentTypes map[string]struct{}
}

var allowedTargets = map[string]targetSpec{
	targetKey("people_identifications", "image_url"): {
		MaxSizeBytes: 10 << 20,
		ContentTypes: contentTypeSet("image/png", "image/jpeg", "image/webp"),
	},
	targetKey("companies", "logo_url"): {
		MaxSizeBytes: 10 << 20,
		ContentTypes: contentTypeSet("image/png", "image/jpeg", "image/webp"),
	},
	targetKey("pets", "image_url"): {
		MaxSizeBytes: 10 << 20,
		ContentTypes: contentTypeSet("image/png", "image/jpeg", "image/webp"),
	},
	targetKey("plans", "image_url"): {
		MaxSizeBytes: 10 << 20,
		ContentTypes: contentTypeSet("image/png", "image/jpeg", "image/webp"),
	},
	targetKey("services", "image_url"): {
		MaxSizeBytes: 10 << 20,
		ContentTypes: contentTypeSet("image/png", "image/jpeg", "image/webp"),
	},
	targetKey("products", "image_url"): {
		MaxSizeBytes: 10 << 20,
		ContentTypes: contentTypeSet("image/png", "image/jpeg", "image/webp"),
	},
	targetKey("company_business_costs", "invoice_url"): {
		MaxSizeBytes: 20 << 20,
		ContentTypes: contentTypeSet("application/pdf", "image/png", "image/jpeg", "image/webp"),
	},
}

var sanitizeFileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func NewService(cfg config.UploadsConfig, signer SignedUploadURLProvider, inspector ObjectInspector) *Service {
	return &Service{
		cfg:       cfg,
		signer:    signer,
		inspector: inspector,
		now:       time.Now,
		newID:     func() string { return uuid.NewString() },
	}
}

func (s *Service) CreateIntent(ctx context.Context, input CreateIntentInput) (CreateIntentOutput, error) {
	if s.signer == nil || strings.TrimSpace(s.cfg.GCSBucketName) == "" {
		return CreateIntentOutput{}, apperror.ErrServiceUnavailable
	}

	spec, resource, field, err := validateTarget(input.Resource, input.Field)
	if err != nil {
		return CreateIntentOutput{}, err
	}
	if err := validateFileForTarget(spec, input.FileName, input.ContentType, input.SizeBytes); err != nil {
		return CreateIntentOutput{}, err
	}

	objectKey := buildObjectKey(s.cfg.GCSUploadsBasePath, resource, field, s.now().UTC(), s.newID(), input.FileName)
	expiresAt := s.now().UTC().Add(s.cfg.GCSSignedURLTTL)
	uploadURL, headers, err := s.signer.SignedUploadURL(ctx, s.cfg.GCSBucketName, objectKey, strings.TrimSpace(input.ContentType), expiresAt)
	if err != nil {
		return CreateIntentOutput{}, err
	}

	return CreateIntentOutput{
		UploadURL: uploadURL,
		Method:    http.MethodPut,
		Headers:   headers,
		ObjectKey: objectKey,
		PublicURL: buildPublicURL(s.cfg, objectKey),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) Complete(ctx context.Context, input CompleteUploadInput) (CompleteUploadOutput, error) {
	objectKey, _, err := s.inspectObject(ctx, input.Resource, input.Field, input.ObjectKey)
	if err != nil {
		return CompleteUploadOutput{}, err
	}

	return CompleteUploadOutput{
		ObjectKey: objectKey,
		PublicURL: buildPublicURL(s.cfg, objectKey),
	}, nil
}

func (s *Service) ResolveObjectKey(ctx context.Context, resource string, field string, objectKey string) (string, error) {
	normalizedKey, _, err := s.inspectObject(ctx, resource, field, objectKey)
	if err != nil {
		return "", err
	}
	return buildPublicURL(s.cfg, normalizedKey), nil
}

func (s *Service) inspectObject(ctx context.Context, resource string, field string, objectKey string) (string, ObjectMetadata, error) {
	if s.inspector == nil || strings.TrimSpace(s.cfg.GCSBucketName) == "" {
		return "", ObjectMetadata{}, apperror.ErrServiceUnavailable
	}

	spec, normalizedResource, normalizedField, err := validateTarget(resource, field)
	if err != nil {
		return "", ObjectMetadata{}, err
	}

	normalizedObjectKey, err := validateObjectKey(s.cfg.GCSUploadsBasePath, normalizedResource, normalizedField, objectKey)
	if err != nil {
		return "", ObjectMetadata{}, err
	}

	metadata, err := s.inspector.StatObject(ctx, s.cfg.GCSBucketName, normalizedObjectKey)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			return "", ObjectMetadata{}, apperror.ErrNotFound
		}
		return "", ObjectMetadata{}, err
	}

	if _, ok := spec.ContentTypes[strings.TrimSpace(strings.ToLower(metadata.ContentType))]; !ok {
		return "", ObjectMetadata{}, apperror.ErrUnprocessableEntity
	}
	if metadata.SizeBytes <= 0 || metadata.SizeBytes > spec.MaxSizeBytes {
		return "", ObjectMetadata{}, apperror.ErrUnprocessableEntity
	}

	return normalizedObjectKey, metadata, nil
}

func validateTarget(resource string, field string) (targetSpec, string, string, error) {
	normalizedResource := strings.TrimSpace(resource)
	normalizedField := strings.TrimSpace(field)
	spec, ok := allowedTargets[targetKey(normalizedResource, normalizedField)]
	if !ok {
		return targetSpec{}, "", "", apperror.ErrUnprocessableEntity
	}
	return spec, normalizedResource, normalizedField, nil
}

func validateFileForTarget(spec targetSpec, fileName string, contentType string, sizeBytes int64) error {
	if strings.TrimSpace(fileName) == "" || strings.TrimSpace(contentType) == "" || sizeBytes <= 0 {
		return apperror.ErrUnprocessableEntity
	}
	if sizeBytes > spec.MaxSizeBytes {
		return apperror.ErrUnprocessableEntity
	}
	if _, ok := spec.ContentTypes[strings.ToLower(strings.TrimSpace(contentType))]; !ok {
		return apperror.ErrUnprocessableEntity
	}
	return nil
}

func buildObjectKey(basePath string, resource string, field string, now time.Time, uniqueID string, fileName string) string {
	parts := []string{strings.Trim(basePath, "/"), resource, field, now.Format("2006"), now.Format("01"), buildObjectFileName(uniqueID, fileName)}
	filtered := parts[:0]
	for _, part := range parts {
		if part == "" {
			continue
		}
		filtered = append(filtered, part)
	}
	return path.Join(filtered...)
}

func buildObjectFileName(uniqueID string, fileName string) string {
	safeName := sanitizeFileName(fileName)
	if safeName == "" {
		safeName = "file"
	}
	return uniqueID + "-" + safeName
}

func sanitizeFileName(fileName string) string {
	trimmed := strings.TrimSpace(fileName)
	if trimmed == "" {
		return ""
	}

	baseName := path.Base(strings.ReplaceAll(trimmed, "\\", "/"))
	sanitized := sanitizeFileNamePattern.ReplaceAllString(baseName, "-")
	sanitized = strings.Trim(sanitized, ".-")
	sanitized = strings.TrimSpace(sanitized)
	if sanitized == "" {
		return ""
	}

	return strings.ToLower(sanitized)
}

func validateObjectKey(basePath string, resource string, field string, objectKey string) (string, error) {
	trimmedKey := strings.Trim(strings.TrimSpace(objectKey), "/")
	if trimmedKey == "" || strings.Contains(trimmedKey, "..") {
		return "", apperror.ErrUnprocessableEntity
	}

	expectedPrefix := path.Join(strings.Trim(basePath, "/"), resource, field)
	if !strings.HasPrefix(trimmedKey, expectedPrefix+"/") {
		return "", apperror.ErrUnprocessableEntity
	}

	return trimmedKey, nil
}

func buildPublicURL(cfg config.UploadsConfig, objectKey string) string {
	trimmedKey := strings.Trim(strings.TrimSpace(objectKey), "/")
	if cfg.GCSPublicBaseURL != "" {
		return strings.TrimRight(cfg.GCSPublicBaseURL, "/") + "/" + trimmedKey
	}
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", cfg.GCSBucketName, trimmedKey)
}

func contentTypeSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
	}
	return result
}

func targetKey(resource string, field string) string {
	return resource + ":" + field
}
