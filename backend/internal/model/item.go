package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Item represents a file or folder in the virtual file system.
type Item struct {
	ID       uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_items_user_parent,priority:1" json:"user_id"`
	ParentID *uuid.UUID `gorm:"type:uuid;index:idx_items_user_parent,priority:2" json:"parent_id"`
	Name     string     `gorm:"size:255;not null" json:"name"`
	IsFolder bool       `gorm:"default:false;not null" json:"is_folder"`
	Depth    int        `gorm:"default:0;not null" json:"depth"`
	Path     string     `gorm:"size:4096;not null;index:idx_items_path" json:"path"`

	// File-only fields
	MimeType   *string `gorm:"size:255" json:"mime_type"`
	Size       int64   `gorm:"default:0" json:"size"`
	StorageKey *string `gorm:"size:1024" json:"storage_key"`

	// Optional metadata
	Color     *string `gorm:"size:7" json:"color"`
	SortOrder int     `gorm:"default:0" json:"sort_order"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User     User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Parent   *Item  `gorm:"foreignKey:ParentID" json:"-"`
	Children []Item `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}
