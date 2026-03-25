package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/model"
	"github.com/biho/onedrive/internal/repository"
	"github.com/biho/onedrive/pkg/storage"
)

const (
	// DefaultURLExpiry is the default expiry time for pre-signed URLs (1 hour)
	DefaultURLExpiry = 3600
	// LargeFileThreshold is the threshold for large file upload (5MB)
	LargeFileThreshold = int64(5 * 1024 * 1024)
	// DefaultPartSize is the default part size for large file upload (5MB)
	DefaultPartSize = int64(5 * 1024 * 1024)
)

// ItemService handles business logic for items.
type ItemService struct {
	itemRepo *repository.ItemRepository
	b2Client *storage.B2Client
	log      *zap.Logger
}

// NewItemService creates a new ItemService.
func NewItemService(itemRepo *repository.ItemRepository, b2Client *storage.B2Client, log *zap.Logger) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		b2Client: b2Client,
		log:      log,
	}
}

// CreateFolder creates a new folder item.
func (s *ItemService) CreateFolder(userID uuid.UUID, req *dto.CreateFolderRequest) (*model.Item, error) {
	// Check duplicate name in same parent
	exists, err := s.itemRepo.NameExistsInParent(userID, req.ParentID, req.Name, nil)
	if err != nil {
		s.log.Error("Failed to check name existence", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}
	if exists {
		return nil, fmt.Errorf("a folder with this name already exists")
	}

	// Calculate depth and path
	depth := 0
	path := "/" + req.Name

	if req.ParentID != nil {
		parent, err := s.itemRepo.FindByID(*req.ParentID, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent folder not found")
			}
			return nil, fmt.Errorf("internal error")
		}
		if !parent.IsFolder {
			return nil, fmt.Errorf("parent is not a folder")
		}
		depth = parent.Depth + 1
		path = parent.Path + "/" + req.Name
	}

	item := &model.Item{
		UserID:   userID,
		ParentID: req.ParentID,
		Name:     req.Name,
		IsFolder: true,
		Depth:    depth,
		Path:     path,
		Color:    req.Color,
	}

	if err := s.itemRepo.Create(item); err != nil {
		s.log.Error("Failed to create folder", zap.Error(err))
		return nil, fmt.Errorf("failed to create folder")
	}

	return item, nil
}

// UploadFile handles file upload to B2 storage.
// It automatically chooses between simple upload (<=5MB) and multipart upload (>5MB).
func (s *ItemService) UploadFile(ctx context.Context, userID uuid.UUID, parentID *uuid.UUID, filename string, fileSize int64, fileContent io.Reader) (*model.Item, error) {
	// Validate parent folder if provided
	var parentPath string
	if parentID != nil {
		parent, err := s.itemRepo.FindByID(*parentID, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent folder not found")
			}
			return nil, fmt.Errorf("internal error")
		}
		if !parent.IsFolder {
			return nil, fmt.Errorf("parent is not a folder")
		}
		parentPath = parent.Path
	}

	// Check for duplicate name
	exists, err := s.itemRepo.NameExistsInParent(userID, parentID, filename, nil)
	if err != nil {
		s.log.Error("Failed to check name existence", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}
	if exists {
		return nil, fmt.Errorf("a file with this name already exists")
	}

	// Determine MIME type
	mimeType := storage.GetMimeType(filename)

	// Generate storage key
	storageKey := storage.GenerateStorageKey(userID.String(), parentPath, filename)

	// Read file content into memory for upload
	content, err := io.ReadAll(fileContent)
	if err != nil {
		s.log.Error("Failed to read file content", zap.Error(err))
		return nil, fmt.Errorf("failed to read file content")
	}

	// Upload to B2
	if s.b2Client == nil {
		s.log.Error("B2 client not configured")
		return nil, fmt.Errorf("storage not configured")
	}

	_, err = s.b2Client.UploadFile(ctx, storageKey, content, mimeType)
	if err != nil {
		s.log.Error("Failed to upload file to B2", zap.Error(err))
		return nil, fmt.Errorf("failed to upload file")
	}

	// Calculate depth
	depth := 0
	if parentID != nil {
		parent, _ := s.itemRepo.FindByID(*parentID, userID)
		if parent != nil {
			depth = parent.Depth + 1
		}
	}

	// Build full path
	fullPath := parentPath + "/" + filename
	if parentPath == "" {
		fullPath = "/" + filename
	}

	// Create item in database
	item := &model.Item{
		UserID:     userID,
		ParentID:   parentID,
		Name:       filename,
		IsFolder:   false,
		Depth:      depth,
		Path:       fullPath,
		MimeType:   &mimeType,
		Size:       fileSize,
		StorageKey: &storageKey,
	}

	if err := s.itemRepo.Create(item); err != nil {
		s.log.Error("Failed to create item", zap.Error(err))
		// Try to delete the uploaded file from B2
		_ = s.b2Client.DeleteFile(ctx, storageKey)
		return nil, fmt.Errorf("failed to save file metadata")
	}

	return item, nil
}

// GetItem returns an item by ID.
func (s *ItemService) GetItem(id, userID uuid.UUID) (*model.Item, int64, error) {
	item, err := s.itemRepo.FindByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, fmt.Errorf("item not found")
		}
		return nil, 0, fmt.Errorf("internal error")
	}

	childCount, _ := s.itemRepo.CountChildren(id)
	return item, childCount, nil
}

// ListItems returns direct children of a parent.
func (s *ItemService) ListItems(userID uuid.UUID, parentID *uuid.UUID) ([]model.Item, error) {
	items, err := s.itemRepo.FindChildren(userID, parentID)
	if err != nil {
		s.log.Error("Failed to list items", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}
	return items, nil
}

// SearchItems searches for items by name for a user.
// Returns up to `limit` results, sorted by relevance (folders first, then exact match, then alphabetical).
func (s *ItemService) SearchItems(userID uuid.UUID, query string, limit int) ([]model.Item, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	items, err := s.itemRepo.Search(userID, query, limit)
	if err != nil {
		s.log.Error("Failed to search items", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	return items, nil
}

func (s *ItemService) ListStarredItems(userID uuid.UUID) ([]dto.ItemResponse, error) {
	items, err := s.itemRepo.ListStarred(userID)
	if err != nil {
		s.log.Error("Failed to list starred items", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}
	return ToItemResponseList(items), nil
}

func (s *ItemService) ListRecentItems(userID uuid.UUID, limit int) ([]dto.RecentItemResponse, error) {
	records, err := s.itemRepo.ListRecentFiles(userID, limit)
	if err != nil {
		s.log.Error("Failed to list recent items", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	responses := make([]dto.RecentItemResponse, 0, len(records))
	for i := range records {
		itemResp := ToItemResponse(&records[i].Item, 0)
		responses = append(responses, dto.RecentItemResponse{
			ItemResponse:    *itemResp,
			LastAccessedAt: records[i].LastAccessedAt.UTC().Format("2006-01-02T15:04:05Z"),
			LastEventType:  records[i].LastEventType,
		})
	}
	return responses, nil
}

func (s *ItemService) AddStar(userID, itemID uuid.UUID) error {
	item, err := s.itemRepo.FindByID(itemID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("item not found")
		}
		return fmt.Errorf("internal error")
	}

	if err := s.itemRepo.UpsertStar(userID, item.ID); err != nil {
		s.log.Error("Failed to star item", zap.Error(err))
		return fmt.Errorf("internal error")
	}
	return nil
}

func (s *ItemService) RemoveStar(userID, itemID uuid.UUID) error {
	if _, err := s.itemRepo.FindByID(itemID, userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("item not found")
		}
		return fmt.Errorf("internal error")
	}

	if err := s.itemRepo.DeleteStar(userID, itemID); err != nil {
		s.log.Error("Failed to unstar item", zap.Error(err))
		return fmt.Errorf("internal error")
	}
	return nil
}

func (s *ItemService) TrackItemActivity(userID, itemID uuid.UUID, eventType string) error {
	item, err := s.itemRepo.FindByID(itemID, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("item not found")
		}
		return fmt.Errorf("internal error")
	}

	if item.IsFolder {
		return fmt.Errorf("recent is only supported for files")
	}

	if err := s.itemRepo.UpsertActivity(userID, itemID, eventType, time.Now().UTC()); err != nil {
		s.log.Error("Failed to track item activity", zap.Error(err))
		return fmt.Errorf("internal error")
	}
	return nil
}

// UpdateItem updates an item's mutable fields.
func (s *ItemService) UpdateItem(id, userID uuid.UUID, req *dto.UpdateItemRequest) (*model.Item, error) {
	item, err := s.itemRepo.FindByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("item not found")
		}
		return nil, fmt.Errorf("internal error")
	}

	if req.Name != nil && *req.Name != item.Name {
		// Check duplicate name
		exists, err := s.itemRepo.NameExistsInParent(userID, item.ParentID, *req.Name, &id)
		if err != nil {
			return nil, fmt.Errorf("internal error")
		}
		if exists {
			return nil, fmt.Errorf("an item with this name already exists")
		}

		oldPath := item.Path
		// Rebuild path with new name
		if item.ParentID != nil {
			parent, err := s.itemRepo.FindByID(*item.ParentID, userID)
			if err == nil {
				item.Path = parent.Path + "/" + *req.Name
			}
		} else {
			item.Path = "/" + *req.Name
		}
		item.Name = *req.Name

		// Update all descendant paths if this is a folder
		if item.IsFolder && oldPath != item.Path {
			if err := s.itemRepo.UpdateChildPaths(userID, oldPath, item.Path, 0); err != nil {
				s.log.Error("Failed to update child paths", zap.Error(err))
			}
		}
	}

	if req.Color != nil {
		item.Color = req.Color
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}

	if err := s.itemRepo.Update(item); err != nil {
		s.log.Error("Failed to update item", zap.Error(err))
		return nil, fmt.Errorf("failed to update item")
	}

	return item, nil
}

// DeleteItem deletes an item and all its descendants.
func (s *ItemService) DeleteItem(id, userID uuid.UUID) error {
	item, err := s.itemRepo.FindByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("item not found")
		}
		return fmt.Errorf("internal error")
	}

	if item.IsFolder {
		_ = s.itemRepo.DeleteAccessDataForItem(userID, id)
		return s.itemRepo.CascadeDelete(id, userID)
	}

	_ = s.itemRepo.DeleteAccessDataForItem(userID, id)
	return s.itemRepo.Delete(id, userID)
}

// GetFolderTree returns the complete folder tree for a user.
func (s *ItemService) GetFolderTree(userID uuid.UUID) ([]dto.ItemTreeNode, error) {
	folders, err := s.itemRepo.GetFolderTree(userID)
	if err != nil {
		s.log.Error("Failed to get folder tree", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	return s.buildTree(folders), nil
}

// buildTree converts a flat list of folders into a tree structure.
func (s *ItemService) buildTree(folders []model.Item) []dto.ItemTreeNode {
	nodeMap := make(map[string]*dto.ItemTreeNode)

	// Pass 1: Create all nodes
	for _, f := range folders {
		id := f.ID.String()
		nodeMap[id] = &dto.ItemTreeNode{
			ID:       id,
			Name:     f.Name,
			Path:     f.Path,
			Depth:    f.Depth,
			Color:    f.Color,
			Children: []dto.ItemTreeNode{},
		}
	}

	// Pass 2: Attach children to parents
	var rootIDs []string
	for _, f := range folders {
		id := f.ID.String()
		if f.ParentID != nil {
			parentID := f.ParentID.String()
			if parent, ok := nodeMap[parentID]; ok {
				parent.Children = append(parent.Children, *nodeMap[id])
			}
		} else {
			rootIDs = append(rootIDs, id)
		}
	}

	// Since folders are ordered by depth ASC, parents are processed before children.
	// We need to rebuild roots from nodeMap to include nested children.
	roots := make([]dto.ItemTreeNode, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, s.buildNodeRecursive(nodeMap[id], nodeMap, folders))
	}

	if roots == nil {
		roots = []dto.ItemTreeNode{}
	}
	return roots
}

// buildNodeRecursive builds a tree node with correctly nested children.
func (s *ItemService) buildNodeRecursive(node *dto.ItemTreeNode, nodeMap map[string]*dto.ItemTreeNode, folders []model.Item) dto.ItemTreeNode {
	result := dto.ItemTreeNode{
		ID:       node.ID,
		Name:     node.Name,
		Path:     node.Path,
		Depth:    node.Depth,
		Color:    node.Color,
		Children: []dto.ItemTreeNode{},
	}

	// Find direct children from the flat list
	for _, f := range folders {
		if f.ParentID != nil && f.ParentID.String() == node.ID {
			childNode := nodeMap[f.ID.String()]
			result.Children = append(result.Children, s.buildNodeRecursive(childNode, nodeMap, folders))
		}
	}

	return result
}

// ToItemResponse converts a model.Item to dto.ItemResponse.
func ToItemResponse(item *model.Item, childCount int64) *dto.ItemResponse {
	var parentID *string
	if item.ParentID != nil {
		pid := item.ParentID.String()
		parentID = &pid
	}

	return &dto.ItemResponse{
		ID:         item.ID.String(),
		ParentID:   parentID,
		Name:       item.Name,
		IsFolder:   item.IsFolder,
		Depth:      item.Depth,
		Path:       item.Path,
		MimeType:   item.MimeType,
		Size:       item.Size,
		Color:      item.Color,
		SortOrder:  item.SortOrder,
		ChildCount: int(childCount),
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToItemResponseList converts a list of items to response DTOs.
func ToItemResponseList(items []model.Item) []dto.ItemResponse {
	responses := make([]dto.ItemResponse, 0, len(items))
	for i := range items {
		responses = append(responses, *ToItemResponse(&items[i], 0))
	}
	return responses
}

// ToSearchResultResponse converts a model.Item to dto.SearchResultResponse.
func ToSearchResultResponse(item *model.Item) *dto.SearchResultResponse {
	var parentID *string
	if item.ParentID != nil {
		pid := item.ParentID.String()
		parentID = &pid
	}

	itemType := "file"
	if item.IsFolder {
		itemType = "folder"
	}

	return &dto.SearchResultResponse{
		ID:        item.ID.String(),
		Name:      item.Name,
		Type:      itemType,
		MimeType:  item.MimeType,
		ParentID:  parentID,
		Path:      item.Path,
		Size:      item.Size,
		UpdatedAt: item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// sanitizeName removes path separators from names.
func sanitizeName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}

// ToUploadResponse converts a model.Item to dto.UploadResponse with download URL.
func ToUploadResponse(item *model.Item, downloadURL *string) *dto.UploadResponse {
	var parentID *string
	if item.ParentID != nil {
		pid := item.ParentID.String()
		parentID = &pid
	}

	mimeType := ""
	if item.MimeType != nil {
		mimeType = *item.MimeType
	}

	storageKey := ""
	if item.StorageKey != nil {
		storageKey = *item.StorageKey
	}

	return &dto.UploadResponse{
		ID:          item.ID.String(),
		Name:        item.Name,
		ParentID:    parentID,
		Size:        item.Size,
		MimeType:    mimeType,
		Path:        item.Path,
		StorageKey:  storageKey,
		DownloadURL: downloadURL,
		CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GetPreSignedUploadURL generates a pre-signed URL for direct upload to B2.
// For files <=5MB, returns a simple upload URL.
// For files >5MB, returns multipart upload initiation info.
func (s *ItemService) GetPreSignedUploadURL(ctx context.Context, userID uuid.UUID, req *dto.GetUploadURLRequest) (interface{}, error) {
	if s.b2Client == nil {
		return nil, fmt.Errorf("storage not configured")
	}

	// Validate parent folder if provided
	var parentPath string
	if req.ParentID != nil {
		parent, err := s.itemRepo.FindByID(*req.ParentID, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent folder not found")
			}
			return nil, fmt.Errorf("internal error")
		}
		if !parent.IsFolder {
			return nil, fmt.Errorf("parent is not a folder")
		}
		parentPath = parent.Path
	}

	// Check for duplicate name
	exists, err := s.itemRepo.NameExistsInParent(userID, req.ParentID, req.Name, nil)
	if err != nil {
		s.log.Error("Failed to check name existence", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}
	if exists {
		return nil, fmt.Errorf("a file with this name already exists")
	}

	// Generate storage key
	storageKey := storage.GenerateStorageKey(userID.String(), parentPath, req.Name)

	// Determine if this is a large file upload
	if req.Size > LargeFileThreshold {
		return s.initiateLargeUpload(ctx, userID, req, storageKey)
	}

	return s.initiateSimpleUpload(ctx, userID, req, storageKey)
}

// initiateSimpleUpload creates a pre-signed URL for simple upload (<=5MB)
func (s *ItemService) initiateSimpleUpload(ctx context.Context, userID uuid.UUID, req *dto.GetUploadURLRequest, storageKey string) (*dto.SimpleUploadResponse, error) {
	// Generate pre-signed upload URL
	uploadURL, err := s.b2Client.GetUploadURL(ctx, storageKey, req.ContentType, DefaultURLExpiry)
	if err != nil {
		s.log.Error("Failed to generate pre-signed upload URL", zap.Error(err))
		return nil, fmt.Errorf("failed to generate upload URL")
	}

	// Create placeholder item in database (status: pending upload)
	depth := 0
	if req.ParentID != nil {
		parent, _ := s.itemRepo.FindByID(*req.ParentID, userID)
		if parent != nil {
			depth = parent.Depth + 1
		}
	}

	parentPath := ""
	if req.ParentID != nil {
		parent, _ := s.itemRepo.FindByID(*req.ParentID, userID)
		if parent != nil {
			parentPath = parent.Path
		}
	}

	fullPath := parentPath + "/" + req.Name
	if parentPath == "" {
		fullPath = "/" + req.Name
	}

	item := &model.Item{
		UserID:     userID,
		ParentID:   req.ParentID,
		Name:       req.Name,
		IsFolder:   false,
		Depth:      depth,
		Path:       fullPath,
		MimeType:   &req.ContentType,
		Size:       req.Size,
		StorageKey: &storageKey,
	}

	if err := s.itemRepo.Create(item); err != nil {
		s.log.Error("Failed to create item", zap.Error(err))
		return nil, fmt.Errorf("failed to create file record")
	}

	return &dto.SimpleUploadResponse{
		UploadURL:  uploadURL,
		StorageKey: storageKey,
		ItemID:     item.ID.String(),
		MimeType:   req.ContentType,
		Size:       req.Size,
	}, nil
}

// initiateLargeUpload initiates a multipart upload for large files (>5MB)
func (s *ItemService) initiateLargeUpload(ctx context.Context, userID uuid.UUID, req *dto.GetUploadURLRequest, storageKey string) (*dto.InitiateLargeUploadResponse, error) {
	// Determine part size (use provided or default)
	partSize := req.PartSize
	if partSize <= 0 {
		partSize = DefaultPartSize
	}

	// Calculate total parts
	totalParts := int((req.Size + partSize - 1) / partSize)

	// Create multipart upload on B2
	result, err := s.b2Client.CreateMultipartUpload(ctx, storageKey, req.ContentType)
	if err != nil {
		s.log.Error("Failed to create multipart upload", zap.Error(err))
		return nil, fmt.Errorf("failed to initiate large upload")
	}

	// Create placeholder item in database
	depth := 0
	if req.ParentID != nil {
		parent, _ := s.itemRepo.FindByID(*req.ParentID, userID)
		if parent != nil {
			depth = parent.Depth + 1
		}
	}

	parentPath := ""
	if req.ParentID != nil {
		parent, _ := s.itemRepo.FindByID(*req.ParentID, userID)
		if parent != nil {
			parentPath = parent.Path
		}
	}

	fullPath := parentPath + "/" + req.Name
	if parentPath == "" {
		fullPath = "/" + req.Name
	}

	item := &model.Item{
		UserID:     userID,
		ParentID:   req.ParentID,
		Name:       req.Name,
		IsFolder:   false,
		Depth:      depth,
		Path:       fullPath,
		MimeType:   &req.ContentType,
		Size:       req.Size,
		StorageKey: &storageKey,
	}

	if err := s.itemRepo.Create(item); err != nil {
		s.log.Error("Failed to create item", zap.Error(err))
		// Abort the multipart upload
		_ = s.b2Client.AbortMultipartUpload(ctx, storageKey, result.UploadID)
		return nil, fmt.Errorf("failed to create file record")
	}

	return &dto.InitiateLargeUploadResponse{
		UploadID:   result.UploadID,
		StorageKey: storageKey,
		ItemID:     item.ID.String(),
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

// UpdateFileContent overwrites an existing file's content in B2 and updates its size in DB.
// This is used e.g. for OnlyOffice editing callbacks.
func (s *ItemService) UpdateFileContent(ctx context.Context, userID uuid.UUID, itemIDStr string, content []byte) error {
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return fmt.Errorf("invalid item ID")
	}

	item, err := s.itemRepo.FindByID(itemID, userID)
	if err != nil {
		return fmt.Errorf("item not found")
	}

	if item.IsFolder || item.StorageKey == nil {
		return fmt.Errorf("item is not a valid file")
	}

	mimeType := "application/octet-stream"
	if item.MimeType != nil {
		mimeType = *item.MimeType
	}

	// 1. Upload to B2 overwriting the same storage key
	_, err = s.b2Client.UploadFile(ctx, *item.StorageKey, content, mimeType)
	if err != nil {
		s.log.Error("Failed to overwrite file in B2", zap.Error(err), zap.String("key", *item.StorageKey))
		return fmt.Errorf("failed to upload modified file")
	}

	// 2. Update DB size
	newSize := int64(len(content))
	item.Size = newSize

	if err := s.itemRepo.Update(item); err != nil {
		s.log.Error("Failed to update item size in DB", zap.Error(err))
		return fmt.Errorf("failed to update item size in database")
	}

	return nil
}

// CompleteLargeUpload completes a multipart upload
func (s *ItemService) CompleteLargeUpload(ctx context.Context, userID uuid.UUID, req *dto.CompleteLargeUploadRequest) (*dto.CompleteLargeUploadResponse, error) {
	if s.b2Client == nil {
		return nil, fmt.Errorf("storage not configured")
	}

	var parts []types.CompletedPart
	hasValidETags := false
	for _, part := range req.Parts {
		if part.ETag != "" && !strings.HasPrefix(part.ETag, "part-") {
			hasValidETags = true
			break
		}
	}

	if hasValidETags {
		// Use provided ETags
		parts = make([]types.CompletedPart, len(req.Parts))
		for i, part := range req.Parts {
			parts[i] = types.CompletedPart{
				ETag:       &part.ETag,
				PartNumber: aws.Int32(int32(part.PartNumber)),
			}
		}
	} else {
		// Fetch parts from B2 to get actual ETags
		var err error
		parts, err = s.b2Client.ListParts(ctx, req.StorageKey, req.UploadID)
		if err != nil {
			s.log.Error("Failed to list parts", zap.Error(err))
			return nil, fmt.Errorf("failed to complete upload: could not retrieve parts")
		}
	}

	// Complete the multipart upload
	err := s.b2Client.CompleteMultipartUpload(ctx, req.StorageKey, req.UploadID, parts)
	if err != nil {
		s.log.Error("Failed to complete multipart upload", zap.Error(err))
		// Try to abort
		_ = s.b2Client.AbortMultipartUpload(ctx, req.StorageKey, req.UploadID)
		return nil, fmt.Errorf("failed to complete upload")
	}

	// Find and update the item
	item, err := s.itemRepo.FindByStorageKey(req.StorageKey, userID)
	if err != nil {
		return nil, fmt.Errorf("file record not found")
	}

	// Generate download URL
	var downloadURL *string
	url, err := s.b2Client.GetFileURL(ctx, req.StorageKey, DefaultURLExpiry)
	if err == nil {
		downloadURL = &url
	}

	mimeType := ""
	if item.MimeType != nil {
		mimeType = *item.MimeType
	}

	return &dto.CompleteLargeUploadResponse{
		ID:          item.ID.String(),
		Name:        item.Name,
		StorageKey:  req.StorageKey,
		Size:        item.Size,
		MimeType:    mimeType,
		DownloadURL: downloadURL,
	}, nil
}

// GetUploadPartURL gets a pre-signed URL for uploading a part in multipart upload
func (s *ItemService) GetUploadPartURL(ctx context.Context, userID uuid.UUID, storageKey string, uploadID string, partNumber int) (string, error) {
	if s.b2Client == nil {
		return "", fmt.Errorf("storage not configured")
	}

	return s.b2Client.GetUploadPartURL(ctx, storageKey, uploadID, partNumber, DefaultURLExpiry)
}

// TrashError represents a trash-specific error.
type TrashError struct {
	Code    string
	Message string
}

func (e *TrashError) Error() string {
	return e.Message
}

// ListTrash returns all soft-deleted items for a user.
func (s *ItemService) ListTrash(userID uuid.UUID) ([]dto.TrashItemResponse, error) {
	items, err := s.itemRepo.FindTrash(userID)
	if err != nil {
		s.log.Error("Failed to list trash", zap.Error(err))
		return nil, fmt.Errorf("internal error")
	}

	responses := make([]dto.TrashItemResponse, 0, len(items))
	for _, item := range items {
		var deletedAt string
		if item.DeletedAt.Valid {
			deletedAt = item.DeletedAt.Time.Format("2006-01-02T15:04:05Z")
		}
		resp := ToItemResponse(&item, 0)
		responses = append(responses, dto.TrashItemResponse{
			ItemResponse: *resp,
			DeletedAt:    deletedAt,
		})
	}

	return responses, nil
}

// RestoreItem restores an item from trash.
func (s *ItemService) RestoreItem(id, userID uuid.UUID, req *dto.RestoreItemRequest) (*dto.ItemResponse, error) {
	s.log.Info("RestoreItem called",
		zap.String("itemID", id.String()),
		zap.String("userID", userID.String()),
		zap.Any("req", req),
	)

	item, err := s.itemRepo.FindByIDUnscoped(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("item not found")
		}
		return nil, fmt.Errorf("internal error")
	}

	if item.DeletedAt.Valid == false {
		return nil, fmt.Errorf("item is not in trash")
	}

	var targetParentID *uuid.UUID
	var targetPath string
	var targetDepth int

	if req != nil && req.TargetParentID != nil {
		pid, err := uuid.Parse(*req.TargetParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent ID")
		}
		targetParentID = &pid

		parent, err := s.itemRepo.FindByID(pid, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, &TrashError{Code: "PARENT_DELETED", Message: "target parent not found"}
			}
			return nil, fmt.Errorf("internal error")
		}
		if !parent.IsFolder {
			return nil, fmt.Errorf("target parent is not a folder")
		}
		if parent.DeletedAt.Valid {
			return nil, &TrashError{Code: "PARENT_DELETED", Message: "target parent is in trash"}
		}
		targetPath = parent.Path + "/" + item.Name
		targetDepth = parent.Depth + 1
	} else {
		if item.ParentID != nil {
			targetParentID = item.ParentID
			parent, err := s.itemRepo.FindByID(*item.ParentID, userID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, &TrashError{Code: "PARENT_DELETED", Message: "original parent not found"}
				}
				return nil, fmt.Errorf("internal error")
			}
			if parent.DeletedAt.Valid {
				return nil, &TrashError{Code: "PARENT_DELETED", Message: "original parent is in trash"}
			}
			targetPath = parent.Path + "/" + item.Name
			targetDepth = parent.Depth + 1
		} else {
			targetPath = "/" + item.Name
			targetDepth = 0
		}
	}

	newName := item.Name
	if req != nil && req.NewName != nil && *req.NewName != "" {
		newName = *req.NewName
		targetPath = getPathPrefix(targetPath) + "/" + newName
	}

	exists, err := s.itemRepo.NameExistsInParent(userID, targetParentID, newName, &id)
	if err != nil {
		return nil, fmt.Errorf("internal error")
	}
	s.log.Info("NameExistsInParent check",
		zap.String("itemID", id.String()),
		zap.Any("parentID", targetParentID),
		zap.String("name", newName),
		zap.Bool("exists", exists),
	)
	if exists {
		return nil, &TrashError{Code: "NAME_CONFLICT", Message: "an item with this name already exists"}
	}

	err = s.itemRepo.RestoreInTransaction(id, userID, targetParentID, newName, targetDepth, targetPath)
	if err != nil {
		s.log.Error("Failed to restore item", zap.Error(err))
		return nil, fmt.Errorf("failed to restore item")
	}

	restored, err := s.itemRepo.FindByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("internal error")
	}

	return ToItemResponse(restored, 0), nil
}

// PermanentDeleteItem permanently deletes an item from trash (DB + B2).
func (s *ItemService) PermanentDeleteItem(id, userID uuid.UUID) error {
	item, err := s.itemRepo.FindByIDUnscoped(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("item not found")
		}
		s.log.Error("Failed to find item for permanent delete", zap.Error(err))
		return fmt.Errorf("internal error")
	}
	if item.DeletedAt.Valid == false {
		return fmt.Errorf("item is not in trash")
	}

	keys, err := s.itemRepo.GetStorageKeysForPermanentDelete(id, userID)
	if err != nil {
		s.log.Error("Failed to get storage keys", zap.Error(err))
		return fmt.Errorf("internal error")
	}

	ctx := context.Background()
	var b2Errors []string

	for _, key := range keys {
		if err := s.b2Client.DeleteFile(ctx, key); err != nil {
			s.log.Warn("Failed to delete B2 file", zap.String("key", key), zap.Error(err))
			b2Errors = append(b2Errors, key)
		}
	}

	if len(keys) > 0 && len(b2Errors) == len(keys) {
		return fmt.Errorf("storage service unavailable")
	}

	if len(b2Errors) > 0 {
		s.log.Warn("Some B2 files failed to delete",
			zap.Int("total", len(keys)),
			zap.Int("failed", len(b2Errors)))
	}

	if err := s.itemRepo.PermanentDelete(id, userID); err != nil {
		s.log.Error("Failed to permanent delete from DB", zap.Error(err))
		return fmt.Errorf("failed to delete item")
	}

	return nil
}

func getPathPrefix(fullPath string) string {
	lastSlash := strings.LastIndex(fullPath, "/")
	if lastSlash <= 0 {
		return ""
	}
	return fullPath[:lastSlash]
}
