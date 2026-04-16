package services

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	"go-fx-template/internal/api/controllers"
)

type ObjectStorage interface {
	Upload(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	GetObjectURL(key string) string
	ExtractObjectKey(storedValue string) string
	Delete(ctx context.Context, key string) error
}

type ImageServiceImpl struct{ objectStorage ObjectStorage }

func newImageService(os ObjectStorage) *ImageServiceImpl { return &ImageServiceImpl{objectStorage: os} }

func NewImageService(os ObjectStorage) controllers.ImageService { return newImageService(os) }

func (s *ImageServiceImpl) Upload(ctx context.Context, entity string, entityID uuid.UUID, reader io.Reader, size int64, contentType string) (string, error) {
	key := fmt.Sprintf("%s/%s/%s", entity, entityID, uuid.Must(uuid.NewV7()))

	if err := s.objectStorage.Upload(ctx, key, reader, size, contentType); err != nil {
		return "", err
	}

	return s.objectStorage.GetObjectURL(key), nil
}

func (s *ImageServiceImpl) Delete(ctx context.Context, storedValue string) error {
	key := s.objectStorage.ExtractObjectKey(storedValue)
	if key == "" {
		return nil
	}

	return s.objectStorage.Delete(ctx, key)
}
