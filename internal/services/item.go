package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"go-fx-template/internal/api/controllers"
	"go-fx-template/internal/api/models"
	"go-fx-template/internal/models"
	"go-fx-template/internal/services/errors"
)

type ItemStorage interface {
	CreateItem(ctx context.Context, item models.Item) error
	GetAllItemsForUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]models.Item, int64, error)
	GetItemByID(ctx context.Context, id uuid.UUID, r uuid.UUID) (models.Item, error)
	UpdateItemByID(ctx context.Context, id uuid.UUID, r uuid.UUID, item models.Item, fields []string) error
	RemoveItemImage(ctx context.Context, itemID, userID uuid.UUID) error
	DeleteItemByID(ctx context.Context, id uuid.UUID, r uuid.UUID) error
}
type ItemServiceImpl struct{ itemStorage ItemStorage }

func newItemService(is ItemStorage) *ItemServiceImpl { return &ItemServiceImpl{itemStorage: is} }

func NewItemService(is ItemStorage) controllers.ItemService { return newItemService(is) }

func (svc *ItemServiceImpl) CreateItem(ctx context.Context, userID uuid.UUID, req apiModels.CreateItemRequest) (models.Item, error) {
	if userID == uuid.Nil {
		return models.Item{}, svcErr.ValidationError{Message: "user ID is required"}
	}

	if err := req.Validate(); err != nil {
		return models.Item{}, svcErr.ValidationError{Message: "validation failed: " + err.Error()}
	}

	item := models.Item{
		ID:          uuid.Must(uuid.NewV7()), // Will panic only if `rand` is broken
		UserID:      userID,
		Image:       nil,
		Title:       req.Title,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := svc.itemStorage.CreateItem(ctx, item); err != nil {
		return models.Item{}, svcErr.ReturnMappedStorageError(err, "create", "item", "ID", item.ID.String())
	}

	return item, nil
}

func (svc *ItemServiceImpl) GetAllItemsForUser(ctx context.Context, userID uuid.UUID, paginationParams apiModels.PaginationParams) ([]models.Item, int64, error) {
	if userID == uuid.Nil {
		return nil, -1, svcErr.ValidationError{Message: "user ID is required"}
	}

	paginationParams.Normalize()

	items, total, err := svc.itemStorage.GetAllItemsForUser(ctx, userID, paginationParams.Offset(), paginationParams.Limit)
	if err != nil {
		return nil, -1, svcErr.ReturnMappedStorageError(err, "get", "items for user", "ID", userID.String())
	}

	return items, total, nil
}

func (svc *ItemServiceImpl) GetItemByID(ctx context.Context, itemID, requestedByUserID uuid.UUID) (models.Item, error) {
	if itemID == uuid.Nil {
		return models.Item{}, svcErr.ValidationError{Message: "item ID is required"}
	}

	if requestedByUserID == uuid.Nil {
		return models.Item{}, svcErr.ValidationError{Message: "user ID is required"}
	}

	item, err := svc.itemStorage.GetItemByID(ctx, itemID, requestedByUserID)
	if err != nil {
		return models.Item{}, svcErr.ReturnMappedStorageError(err, "get", "item", "ID", itemID.String())
	}

	return item, nil
}

func (svc *ItemServiceImpl) UpdateItemByID(ctx context.Context, itemID uuid.UUID, requestedByUserID uuid.UUID, req apiModels.UpdateItemRequest) (models.Item, error) {
	if itemID == uuid.Nil {
		return models.Item{}, svcErr.ValidationError{Message: "item ID is required"}
	}

	if requestedByUserID == uuid.Nil {
		return models.Item{}, svcErr.ValidationError{Message: "user ID is required"}
	}

	if err := req.Validate(); err != nil {
		return models.Item{}, svcErr.ValidationError{Message: "validation failed: " + err.Error()}
	}

	item := models.Item{UpdatedAt: time.Now()}
	fields := []string{"UpdatedAt"}
	if req.Image != nil {
		item.Image = req.Image
		fields = append(fields, "Image")
	}
	if req.Title != nil {
		item.Title = *req.Title
		fields = append(fields, "Title")
	}
	if req.Description != nil {
		item.Description = req.Description
		fields = append(fields, "Description")
	}

	if err := svc.itemStorage.UpdateItemByID(ctx, itemID, requestedByUserID, item, fields); err != nil {
		return models.Item{}, svcErr.ReturnMappedStorageError(err, "update", "item", "ID", itemID.String())
	}

	item, err := svc.GetItemByID(ctx, itemID, requestedByUserID)
	if err != nil {
		return models.Item{}, svcErr.ReturnMappedStorageError(err, "update", "item", "ID", itemID.String())
	}

	return item, nil
}

func (svc *ItemServiceImpl) RemoveItemImage(ctx context.Context, itemID, requestedByUserID uuid.UUID) error {
	if itemID == uuid.Nil {
		return svcErr.ValidationError{Message: "item ID is required"}
	}
	if requestedByUserID == uuid.Nil {
		return svcErr.ValidationError{Message: "user ID is required"}
	}

	if err := svc.itemStorage.RemoveItemImage(ctx, itemID, requestedByUserID); err != nil {
		return svcErr.ReturnMappedStorageError(err, "remove image", "item", "ID", itemID.String())
	}

	return nil
}

func (svc *ItemServiceImpl) DeleteItemByID(ctx context.Context, itemID uuid.UUID, requestedByUserID uuid.UUID) error {
	if itemID == uuid.Nil {
		return svcErr.ValidationError{Message: "item ID is required"}
	}

	if requestedByUserID == uuid.Nil {
		return svcErr.ValidationError{Message: "user ID is required"}
	}

	if err := svc.itemStorage.DeleteItemByID(ctx, itemID, requestedByUserID); err != nil {
		return svcErr.ReturnMappedStorageError(err, "delete", "item", "ID", itemID.String())
	}

	return nil
}
