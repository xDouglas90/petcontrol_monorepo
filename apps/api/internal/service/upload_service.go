package service

import (
	"context"
	"net/http"
	"time"

	"github.com/xdouglas90/petcontrol_monorepo/internal/apperror"
	"github.com/xdouglas90/petcontrol_monorepo/internal/storage/gcs"
)

type UploadService struct {
	storage *gcs.Service
}

type CreateUploadIntentInput struct {
	Resource    string
	Field       string
	FileName    string
	ContentType string
	SizeBytes   int64
}

type CreateUploadIntentOutput struct {
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

func NewUploadService(storage *gcs.Service) *UploadService {
	return &UploadService{storage: storage}
}

func (s *UploadService) CreateIntent(ctx context.Context, input CreateUploadIntentInput) (CreateUploadIntentOutput, error) {
	if s == nil || s.storage == nil {
		return CreateUploadIntentOutput{}, apperror.ErrServiceUnavailable
	}
	result, err := s.storage.CreateIntent(ctx, gcs.CreateIntentInput(input))
	if err != nil {
		return CreateUploadIntentOutput{}, err
	}

	return CreateUploadIntentOutput{
		UploadURL: result.UploadURL,
		Method:    result.Method,
		Headers:   result.Headers,
		ObjectKey: result.ObjectKey,
		PublicURL: result.PublicURL,
		ExpiresAt: result.ExpiresAt,
	}, nil
}

func (s *UploadService) Complete(ctx context.Context, input CompleteUploadInput) (CompleteUploadOutput, error) {
	if s == nil || s.storage == nil {
		return CompleteUploadOutput{}, apperror.ErrServiceUnavailable
	}
	result, err := s.storage.Complete(ctx, gcs.CompleteUploadInput(input))
	if err != nil {
		return CompleteUploadOutput{}, err
	}

	return CompleteUploadOutput{
		ObjectKey: result.ObjectKey,
		PublicURL: result.PublicURL,
	}, nil
}

func (s *UploadService) ResolveObjectKey(ctx context.Context, resource string, field string, objectKey string) (string, error) {
	if s == nil || s.storage == nil {
		return "", apperror.ErrServiceUnavailable
	}
	return s.storage.ResolveObjectKey(ctx, resource, field, objectKey)
}
