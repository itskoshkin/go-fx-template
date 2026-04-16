package objectStorage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"

	"go-fx-template/internal/config"
	"go-fx-template/internal/services"
)

type MinioStorage struct {
	client *minio.Client
	bucket string
}

func newMinioStorage(client *minio.Client) *MinioStorage {
	return &MinioStorage{client: client, bucket: viper.GetString(config.MinioBucketName)}
}

func NewMinioStorage(client *minio.Client) services.ObjectStorage { return newMinioStorage(client) }

func (s *MinioStorage) GetBaseURL() string {
	return fmt.Sprintf("%s/%s", s.client.EndpointURL().String(), s.bucket)
}

func (s *MinioStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

func (s *MinioStorage) GetObjectURL(key string) string {
	return fmt.Sprintf("%s/%s/%s", s.client.EndpointURL().String(), s.bucket, key)
}

func (s *MinioStorage) ExtractObjectKey(storedValue string) string {
	storedValue = strings.TrimSpace(storedValue)
	baseURL := strings.TrimSpace(s.GetBaseURL())

	if storedValue == "" {
		return ""
	}

	if baseURL == "" {
		return strings.TrimPrefix(storedValue, "/")
	}

	parsedStoredURL, storedErr := url.Parse(storedValue)
	parsedBaseURL, baseErr := url.Parse(baseURL)
	if storedErr == nil && baseErr == nil {
		basePath := strings.TrimSuffix(parsedBaseURL.Path, "/")
		objectPath := strings.TrimPrefix(parsedStoredURL.Path, basePath+"/")
		return strings.TrimPrefix(objectPath, "/")
	}

	objectKey := strings.TrimPrefix(storedValue, strings.TrimSuffix(baseURL, "/")+"/")
	if queryStart := strings.Index(objectKey, "?"); queryStart >= 0 {
		objectKey = objectKey[:queryStart]
	}

	return strings.TrimPrefix(objectKey, "/")
}

func (s *MinioStorage) Delete(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}
