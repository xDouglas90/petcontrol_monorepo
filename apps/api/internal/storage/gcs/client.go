package gcs

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"encoding/json"
	"github.com/xdouglas90/petcontrol_monorepo/internal/config"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"os"
)

var ErrObjectNotFound = errors.New("object not found")

type Client struct {
	storageClient        *storage.Client
	signerServiceAccount string
	signerPrivateKey     []byte
}

func NewClient(ctx context.Context, cfg config.UploadsConfig) (*Client, error) {
	if strings.TrimSpace(cfg.GCSBucketName) == "" {
		return nil, nil
	}

	var opts []option.ClientOption
	if cfg.GCSCredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.GCSCredentialsFile))
	}

	storageClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}

	client := &Client{
		storageClient:        storageClient,
		signerServiceAccount: strings.TrimSpace(cfg.GCSSignerServiceAccount),
		signerPrivateKey:     []byte(cfg.GCSSignerPrivateKey),
	}

	if (client.signerServiceAccount == "" || len(client.signerPrivateKey) == 0) && cfg.GCSCredentialsFile != "" {
		data, err := os.ReadFile(cfg.GCSCredentialsFile)
		if err == nil {
			var creds struct {
				ClientEmail string `json:"client_email"`
				PrivateKey  string `json:"private_key"`
			}
			if err := json.Unmarshal(data, &creds); err == nil {
				if client.signerServiceAccount == "" {
					client.signerServiceAccount = creds.ClientEmail
				}
				if len(client.signerPrivateKey) == 0 {
					client.signerPrivateKey = []byte(creds.PrivateKey)
				}
			}
		}
	}

	return client, nil
}

func (c *Client) Close() error {
	if c == nil || c.storageClient == nil {
		return nil
	}
	return c.storageClient.Close()
}

func (c *Client) SignedUploadURL(_ context.Context, bucketName string, objectKey string, contentType string, expiresAt time.Time) (string, http.Header, error) {
	options := &storage.SignedURLOptions{
		Scheme:         storage.SigningSchemeV4,
		Method:         http.MethodPut,
		GoogleAccessID: c.signerServiceAccount,
		PrivateKey:     c.signerPrivateKey,
		Expires:        expiresAt,
		ContentType:    contentType,
	}

	signedURL, err := storage.SignedURL(bucketName, objectKey, options)
	if err != nil {
		return "", nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", contentType)
	return signedURL, headers, nil
}

func (c *Client) StatObject(ctx context.Context, bucketName string, objectKey string) (ObjectMetadata, error) {
	attrs, err := c.storageClient.Bucket(bucketName).Object(objectKey).Attrs(ctx)
	if err != nil {
		var googleErr *googleapi.Error
		if errors.As(err, &googleErr) && googleErr.Code == http.StatusNotFound {
			return ObjectMetadata{}, ErrObjectNotFound
		}
		if errors.Is(err, storage.ErrObjectNotExist) {
			return ObjectMetadata{}, ErrObjectNotFound
		}
		return ObjectMetadata{}, err
	}

	return ObjectMetadata{
		ContentType: attrs.ContentType,
		SizeBytes:   attrs.Size,
	}, nil
}
