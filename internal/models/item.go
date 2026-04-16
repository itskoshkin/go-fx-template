package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Item struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	User        *User     `gorm:"constraint:OnDelete:CASCADE;"`
	Image       *string   `gorm:"type:text"`
	Title       string    `gorm:"size:255;not null"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}
