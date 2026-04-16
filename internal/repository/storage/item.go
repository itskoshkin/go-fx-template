package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-fx-template/internal/models"
	"go-fx-template/internal/services"
)

type ItemStorageImpl struct {
	db *gorm.DB
}

func newItemStorage(db *gorm.DB) *ItemStorageImpl { return &ItemStorageImpl{db: db} }

func NewItemStorage(db *gorm.DB) services.ItemStorage { return newItemStorage(db) }

func (s *ItemStorageImpl) CreateItem(ctx context.Context, item models.Item) error {
	if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
		return fmt.Errorf("create item: %w", err)
	}

	return nil
}

func (s *ItemStorageImpl) GetAllItemsForUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]models.Item, int64, error) {
	var items []models.Item
	var total int64

	result := s.db.WithContext(ctx).Model(&models.Item{}).Where("user_id = ?", userID)
	if err := result.Count(&total).Error; err != nil {
		return nil, -1, fmt.Errorf("count items for user %s: %w", userID, err)
	}

	if err := result.Order("created_at DESC").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, -1, fmt.Errorf("list items for user %s: %w", userID, err)
	}

	return items, total, nil
}

func (s *ItemStorageImpl) GetItemByID(ctx context.Context, itemID uuid.UUID, requestedByUserID uuid.UUID) (models.Item, error) {
	var item models.Item
	if err := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", itemID, requestedByUserID).First(&item).Error; err != nil {
		return models.Item{}, fmt.Errorf("get item by ID %s: %w", itemID, err)
	}

	return item, nil
}

func (s *ItemStorageImpl) UpdateItemByID(ctx context.Context, itemID uuid.UUID, requestedByUserID uuid.UUID, item models.Item, fields []string) error {
	result := s.db.WithContext(ctx).Model(&models.Item{}).Where("id = ? AND user_id = ?", itemID, requestedByUserID).Select(fields).Updates(&item)
	if result.Error != nil {
		return fmt.Errorf("update item with ID %s: %w", itemID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update item with ID %s: %w", itemID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *ItemStorageImpl) RemoveItemImage(ctx context.Context, itemID, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&models.Item{}).Where("id = ? AND user_id = ?", itemID, userID).Updates(map[string]any{"image": nil, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("remove item image with ID %s: %w", itemID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("remove item image with ID %s: %w", itemID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *ItemStorageImpl) DeleteItemByID(ctx context.Context, itemID uuid.UUID, requestedByUserID uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ? AND user_id = ?", itemID, requestedByUserID).Delete(&models.Item{})
	if result.Error != nil {
		return fmt.Errorf("delete item with ID %s: %w", itemID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete item with ID %s: %w", itemID, gorm.ErrRecordNotFound)
	}

	return nil
}
