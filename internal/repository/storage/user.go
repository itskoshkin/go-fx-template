package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go-fx-template/internal/models"
	"go-fx-template/internal/repository/storage/errors"
	"go-fx-template/internal/services"
)

type UserStorageImpl struct {
	db *gorm.DB
}

func newUserStorage(db *gorm.DB) *UserStorageImpl { return &UserStorageImpl{db: db} }

func NewUserStorage(db *gorm.DB) services.UserStorage { return newUserStorage(db) }

func (s *UserStorageImpl) CreateUser(ctx context.Context, user models.User) error {
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		if dupErr := dbErr.DuplicateFieldError(err); dupErr != nil {
			return fmt.Errorf("create user: %w", dupErr)
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *UserStorageImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("get user by ID %s: %w", userID, err)
	}

	return user, nil
}

func (s *UserStorageImpl) GetUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("get user by username %q: %w", username, err)
	}

	return user, nil
}

func (s *UserStorageImpl) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("get user by email %q: %w", email, err)
	}

	return user, nil
}

func (s *UserStorageImpl) UpdateUserByID(ctx context.Context, userID uuid.UUID, user models.User, fields []string) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Select(fields).Updates(&user)
	if result.Error != nil {
		if dupErr := dbErr.DuplicateFieldError(result.Error); dupErr != nil {
			return fmt.Errorf("update user with ID %s: %w", userID, dupErr)
		}
		return fmt.Errorf("update user with ID %s: %w", userID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update user with ID %s: %w", userID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *UserStorageImpl) UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{"password": password, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("update password for user with ID %s: %w", userID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("update password for user with ID %s: %w", userID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *UserStorageImpl) RemoveAvatar(ctx context.Context, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{"avatar": nil, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("remove avatar for user with ID %s: %w", userID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("remove avatar for user with ID %s: %w", userID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *UserStorageImpl) SetUserEmailAsVerified(ctx context.Context, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{"email_verified": true, "updated_at": time.Now()})
	if result.Error != nil {
		return fmt.Errorf("verify email for user with ID %s: %w", userID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("verify email for user with ID %s: %w", userID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *UserStorageImpl) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ?", userID).Delete(&models.User{})
	if result.Error != nil {
		return fmt.Errorf("delete user with ID %s: %w", userID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("delete user with ID %s: %w", userID, gorm.ErrRecordNotFound)
	}

	return nil
}
