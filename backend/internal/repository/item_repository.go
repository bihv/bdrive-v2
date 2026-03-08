package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/model"
)

// ItemRepository handles database operations for items.
type ItemRepository struct {
	db *gorm.DB
}

// NewItemRepository creates a new ItemRepository.
func NewItemRepository(db *gorm.DB) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create inserts a new item into the database.
func (r *ItemRepository) Create(item *model.Item) error {
	return r.db.Create(item).Error
}

// FindByID finds an item by its ID and user ID.
func (r *ItemRepository) FindByID(id, userID uuid.UUID) (*model.Item, error) {
	var item model.Item
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	return &item, err
}

// FindChildren returns all direct children of a parent item for a user.
func (r *ItemRepository) FindChildren(userID uuid.UUID, parentID *uuid.UUID) ([]model.Item, error) {
	var items []model.Item
	query := r.db.Where("user_id = ?", userID)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	err := query.Order("is_folder DESC, sort_order ASC, name ASC").Find(&items).Error
	return items, err
}

// FindRootItems returns all root-level items for a user.
func (r *ItemRepository) FindRootItems(userID uuid.UUID) ([]model.Item, error) {
	return r.FindChildren(userID, nil)
}

// Update updates an item in the database.
func (r *ItemRepository) Update(item *model.Item) error {
	return r.db.Save(item).Error
}

// Delete soft-deletes an item.
func (r *ItemRepository) Delete(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Item{}).Error
}

// DeleteByParentID soft-deletes all children of a parent item.
func (r *ItemRepository) DeleteByParentID(parentID, userID uuid.UUID) error {
	return r.db.Where("parent_id = ? AND user_id = ?", parentID, userID).Delete(&model.Item{}).Error
}

// NameExistsInParent checks if an item name already exists in the same parent.
func (r *ItemRepository) NameExistsInParent(userID uuid.UUID, parentID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	query := r.db.Model(&model.Item{}).Where("user_id = ? AND name = ?", userID, name)

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

// CountChildren returns the number of direct children of an item.
func (r *ItemRepository) CountChildren(itemID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&model.Item{}).Where("parent_id = ?", itemID).Count(&count).Error
	return count, err
}

// GetAllDescendantIDs returns all descendant IDs using recursive CTE.
func (r *ItemRepository) GetAllDescendantIDs(itemID, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id FROM items WHERE parent_id = ? AND user_id = ? AND deleted_at IS NULL
			UNION ALL
			SELECT i.id FROM items i
			INNER JOIN descendants d ON i.parent_id = d.id
			WHERE i.deleted_at IS NULL
		)
		SELECT id FROM descendants
	`
	err := r.db.Raw(query, itemID, userID).Pluck("id", &ids).Error
	return ids, err
}

// UpdateChildPaths batch updates path and depth for descendants when a parent is renamed.
func (r *ItemRepository) UpdateChildPaths(userID uuid.UUID, oldPath, newPath string, depthDiff int) error {
	query := `
		UPDATE items
		SET path = ? || SUBSTRING(path FROM ?),
		    depth = depth + ?,
		    updated_at = NOW()
		WHERE user_id = ?
		  AND path LIKE ?
		  AND deleted_at IS NULL
	`
	return r.db.Exec(query, newPath, len(oldPath)+1, depthDiff, userID, oldPath+"/%").Error
}

// GetFolderTree returns all folders for a user, ordered for tree building.
func (r *ItemRepository) GetFolderTree(userID uuid.UUID) ([]model.Item, error) {
	var items []model.Item
	err := r.db.Where("user_id = ? AND is_folder = true", userID).
		Order("depth ASC, sort_order ASC, name ASC").
		Find(&items).Error
	return items, err
}

// GetByPath finds an item by its path and user ID.
func (r *ItemRepository) GetByPath(userID uuid.UUID, path string) (*model.Item, error) {
	var item model.Item
	err := r.db.Where("user_id = ? AND path = ?", userID, path).First(&item).Error
	return &item, err
}

// CascadeDelete soft-deletes an item and all its descendants.
func (r *ItemRepository) CascadeDelete(itemID, userID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Get all descendant IDs first
		var ids []uuid.UUID
		query := `
			WITH RECURSIVE descendants AS (
				SELECT id FROM items WHERE parent_id = ? AND user_id = ? AND deleted_at IS NULL
				UNION ALL
				SELECT i.id FROM items i
				INNER JOIN descendants d ON i.parent_id = d.id
				WHERE i.deleted_at IS NULL
			)
			SELECT id FROM descendants
		`
		if err := tx.Raw(query, itemID, userID).Pluck("id", &ids).Error; err != nil {
			return fmt.Errorf("failed to get descendants: %w", err)
		}

		// Delete all descendants
		if len(ids) > 0 {
			if err := tx.Where("id IN ? AND user_id = ?", ids, userID).Delete(&model.Item{}).Error; err != nil {
				return fmt.Errorf("failed to delete descendants: %w", err)
			}
		}

		// Delete the item itself
		if err := tx.Where("id = ? AND user_id = ?", itemID, userID).Delete(&model.Item{}).Error; err != nil {
			return fmt.Errorf("failed to delete item: %w", err)
		}

		return nil
	})
}
