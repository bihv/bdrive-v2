package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/dto"
	"github.com/biho/onedrive/internal/model"
	"github.com/biho/onedrive/internal/repository"
)

// ItemService handles business logic for items.
type ItemService struct {
	itemRepo *repository.ItemRepository
	log      *zap.Logger
}

// NewItemService creates a new ItemService.
func NewItemService(itemRepo *repository.ItemRepository, log *zap.Logger) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
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
		return s.itemRepo.CascadeDelete(id, userID)
	}

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

// sanitizeName removes path separators from names.
func sanitizeName(name string) string {
	return strings.ReplaceAll(name, "/", "_")
}
