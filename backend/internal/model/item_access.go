package model

import (
	"time"

	"github.com/google/uuid"
)

type ItemStar struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	ItemID    uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Item Item `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE" json:"-"`
}

type ItemActivity struct {
	UserID         uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	ItemID         uuid.UUID `gorm:"type:uuid;not null;primaryKey"`
	LastEventType  string    `gorm:"size:32;not null"`
	LastAccessedAt time.Time `gorm:"not null;index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Item Item `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE" json:"-"`
}
