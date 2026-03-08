package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a registered user.
type User struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Email             string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password          string         `gorm:"size:500;not null" json:"-"`
	FullName          string         `gorm:"size:255;not null" json:"full_name"`
	AvatarURL         *string        `gorm:"size:500" json:"avatar_url"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	IsVerified        bool           `gorm:"default:false" json:"is_verified"`
	Role              string         `gorm:"size:50;default:user" json:"role"`
	StorageQuotaBytes int64          `gorm:"default:5368709120" json:"storage_quota_bytes"`
	StorageUsedBytes  int64          `gorm:"default:0" json:"storage_used_bytes"`
	LastLoginAt       *time.Time     `json:"last_login_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// RefreshToken stores refresh tokens for token rotation.
type RefreshToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash  string     `gorm:"size:500;not null;index" json:"-"`
	DeviceInfo *string    `gorm:"size:500" json:"device_info"`
	IPAddress  string     `gorm:"size:50" json:"ip_address"`
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
