package dto

import "github.com/google/uuid"

// CreateFolderRequest is the request body for creating a new folder.
type CreateFolderRequest struct {
	Name     string     `json:"name" validate:"required,min=1,max=255"`
	ParentID *uuid.UUID `json:"parent_id" validate:"omitempty"`
	Color    *string    `json:"color" validate:"omitempty,max=7"`
}

// UpdateItemRequest is the request body for updating an item.
type UpdateItemRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=255"`
	Color     *string `json:"color" validate:"omitempty,max=7"`
	SortOrder *int    `json:"sort_order" validate:"omitempty,min=0"`
}

// ItemResponse is the API response for a single item.
type ItemResponse struct {
	ID         string  `json:"id"`
	ParentID   *string `json:"parent_id"`
	Name       string  `json:"name"`
	IsFolder   bool    `json:"is_folder"`
	Depth      int     `json:"depth"`
	Path       string  `json:"path"`
	MimeType   *string `json:"mime_type"`
	Size       int64   `json:"size"`
	Color      *string `json:"color"`
	SortOrder  int     `json:"sort_order"`
	ChildCount int     `json:"child_count"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// ItemTreeNode is a node in the folder tree response.
type ItemTreeNode struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Path     string         `json:"path"`
	Depth    int            `json:"depth"`
	Color    *string        `json:"color"`
	Children []ItemTreeNode `json:"children"`
}

// UploadRequest is the request body for uploading a file.
// The file content should be sent as multipart form data.
type UploadRequest struct {
	ParentID *uuid.UUID `json:"parent_id" form:"parent_id" validate:"omitempty"`
	Name     string     `json:"name" form:"name" validate:"required,min=1,max=255"`
}

// UploadResponse is the API response for a successful file upload.
type UploadResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ParentID    *string `json:"parent_id"`
	Size        int64   `json:"size"`
	MimeType    string  `json:"mime_type"`
	Path        string  `json:"path"`
	StorageKey  string  `json:"storage_key"`
	DownloadURL *string `json:"download_url"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// GetUploadURLRequest is the request body for getting a pre-signed upload URL.
type GetUploadURLRequest struct {
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"`
	Name        string     `json:"name" validate:"required,min=1,max=255"`
	Size        int64      `json:"size" validate:"required,min=1"`
	ContentType string     `json:"content_type" validate:"required"`
	PartSize    int64      `json:"part_size" validate:"omitempty,min=5242880"`
}

// SimpleUploadResponse is the response for a simple upload pre-signed URL.
type SimpleUploadResponse struct {
	UploadURL  string `json:"upload_url"`
	StorageKey string `json:"storage_key"`
	ItemID     string `json:"item_id"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
}

// InitiateLargeUploadRequest is the request body for initiating a large file upload.
type InitiateLargeUploadRequest struct {
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"`
	Name        string     `json:"name" validate:"required,min=1,max=255"`
	Size        int64      `json:"size" validate:"required,min=1"`
	ContentType string     `json:"content_type" validate:"required"`
	PartSize    int64      `json:"part_size" validate:"required,min=5242880"` // minimum 5MB
}

// InitiateLargeUploadResponse is the response for initiating a large file upload.
type InitiateLargeUploadResponse struct {
	UploadID   string `json:"upload_id"`
	StorageKey string `json:"storage_key"`
	ItemID     string `json:"item_id"`
	PartSize   int64  `json:"part_size"`
	TotalParts int    `json:"total_parts"`
}

// CompleteLargeUploadRequest is the request body for completing a large file upload.
type CompleteLargeUploadRequest struct {
	UploadID   string                 `json:"upload_id" validate:"required"`
	StorageKey string                 `json:"storage_key" validate:"required"`
	Parts      []CompletedPartRequest `json:"parts" validate:"required,min=1"`
}

// CompletedPartRequest represents a completed part.
type CompletedPartRequest struct {
	PartNumber int    `json:"part_number" validate:"required,min=1"`
	ETag       string `json:"etag" validate:"required"`
}

// CompleteLargeUploadResponse is the response for completing a large file upload.
type CompleteLargeUploadResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	StorageKey  string  `json:"storage_key"`
	Size        int64   `json:"size"`
	MimeType    string  `json:"mime_type"`
	DownloadURL *string `json:"download_url"`
}

// RestoreItemRequest is the request body for restoring an item from trash.
type RestoreItemRequest struct {
	TargetParentID *string `json:"targetParentID"`
	NewName        *string `json:"newName"`
}

// TrashItemResponse is the API response for a trash item.
type TrashItemResponse struct {
	ItemResponse
	DeletedAt string `json:"deleted_at"`
}
