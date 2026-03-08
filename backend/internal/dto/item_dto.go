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
