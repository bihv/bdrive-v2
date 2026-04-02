package model

import (
	"time"

	"github.com/google/uuid"
)

type PublicLink struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	OwnerUserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"owner_user_id"`
	ItemID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"item_id"`
	Token          string     `gorm:"size:128;not null;uniqueIndex" json:"token"`
	PasswordHash   *string    `gorm:"size:500" json:"-"`
	ExpiresAt      *time.Time `gorm:"index" json:"expires_at"`
	RevokedAt      *time.Time `gorm:"index" json:"revoked_at"`
	AccessCount    int64      `gorm:"default:0;not null" json:"access_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	SessionVersion int        `gorm:"default:1;not null" json:"session_version"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	Owner User `gorm:"foreignKey:OwnerUserID;constraint:OnDelete:CASCADE" json:"-"`
	Item  Item `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE" json:"-"`
}

type PublicLinkAuditLog struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	PublicLinkID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"public_link_id"`
	ItemID          uuid.UUID  `gorm:"type:uuid;not null;index" json:"item_id"`
	RequestedItemID *uuid.UUID `gorm:"type:uuid;index" json:"requested_item_id"`
	ActorType       string     `gorm:"size:64;not null" json:"actor_type"`
	Action          string     `gorm:"size:64;not null" json:"action"`
	Result          string     `gorm:"size:32;not null" json:"result"`
	DenyReason      *string    `gorm:"size:64" json:"deny_reason"`
	IPAddress       string     `gorm:"size:64" json:"ip_address"`
	UserAgent       string     `gorm:"size:1024" json:"user_agent"`
	CreatedAt       time.Time  `gorm:"index" json:"created_at"`

	PublicLink PublicLink `gorm:"foreignKey:PublicLinkID;constraint:OnDelete:CASCADE" json:"-"`
	Item       Item       `gorm:"foreignKey:ItemID;constraint:OnDelete:CASCADE" json:"-"`
}
