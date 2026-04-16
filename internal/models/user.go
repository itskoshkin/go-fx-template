package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name          string    `gorm:"size:255;not null"`
	Username      string    `gorm:"size:64;not null;uniqueIndex:idx_users_username_unique,where:deleted_at IS NULL"`
	Email         *string   `gorm:"size:320;uniqueIndex:idx_users_email_unique,where:deleted_at IS NULL"`
	EmailVerified bool      `gorm:"default:false"`
	Password      string    `gorm:"size:255;not null"`
	Avatar        *string   `gorm:"type:text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (u User) RemovePassword() User {
	u.Password = ""
	return u
}
