package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/biho/onedrive/internal/model"
)

// ItemRepository handles database operations for items.
type ItemRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

type ItemActivityRecord struct {
	model.Item
	LastEventType  string
	LastAccessedAt time.Time
}

// NewItemRepository creates a new ItemRepository.
func NewItemRepository(db *gorm.DB, log *zap.Logger) *ItemRepository {
	return &ItemRepository{db: db, log: log}
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

// FindByIDUnscoped finds an item by its ID and user ID, including soft-deleted items.
func (r *ItemRepository) FindByIDUnscoped(id, userID uuid.UUID) (*model.Item, error) {
	var item model.Item
	err := r.db.Unscoped().Where("id = ? AND user_id = ?", id, userID).First(&item).Error
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
	query := r.db.Model(&model.Item{}).Where("user_id = ? AND name = ? AND deleted_at IS NULL", userID, name)

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

// FindByStorageKey finds an item by its storage key and user ID.
func (r *ItemRepository) FindByStorageKey(storageKey string, userID uuid.UUID) (*model.Item, error) {
	var item model.Item
	err := r.db.Where("storage_key = ? AND user_id = ?", storageKey, userID).First(&item).Error
	return &item, err
}

// FindTrash returns all soft-deleted items for a user.
func (r *ItemRepository) FindTrash(userID uuid.UUID) ([]model.Item, error) {
	var items []model.Item
	err := r.db.Unscoped().
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Order("deleted_at DESC").
		Find(&items).Error
	return items, err
}

// GetTrashDescendantIDs returns all descendant IDs for a deleted folder.
func (r *ItemRepository) GetTrashDescendantIDs(itemID, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	query := `
		WITH RECURSIVE descendants AS (
			SELECT id FROM items WHERE parent_id = ? AND user_id = ? AND deleted_at IS NOT NULL
			UNION ALL
			SELECT i.id FROM items i
			INNER JOIN descendants d ON i.parent_id = d.id
			WHERE i.deleted_at IS NOT NULL
		)
		SELECT id FROM descendants
	`
	err := r.db.Raw(query, itemID, userID).Pluck("id", &ids).Error
	return ids, err
}

// PermanentDelete hard-deletes an item from the database.
func (r *ItemRepository) PermanentDelete(id, userID uuid.UUID) error {
	return r.db.Unscoped().Where("id = ? AND user_id = ?", id, userID).Delete(&model.Item{}).Error
}

// Restore un-deletes an item by setting deleted_at to NULL.
func (r *ItemRepository) Restore(id, userID uuid.UUID) error {
	return r.db.Unscoped().Model(&model.Item{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("deleted_at", nil).Error
}

// RestoreInTransaction restores an item and optionally moves it to a new parent.
// If the item is a folder, it also restores all descendants and updates their paths.
func (r *ItemRepository) RestoreInTransaction(id, userID uuid.UUID, parentID *uuid.UUID, name string, depth int, path string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// First, get the original item path to calculate path changes for descendants
		var originalItem model.Item
		if err := tx.Unscoped().Where("id = ? AND user_id = ?", id, userID).First(&originalItem).Error; err != nil {
			return fmt.Errorf("failed to find item: %w", err)
		}
		originalPath := originalItem.Path
		isFolder := originalItem.IsFolder

		// Restore the main item
		if err := tx.Unscoped().Model(&model.Item{}).
			Where("id = ? AND user_id = ?", id, userID).
			Updates(map[string]interface{}{
				"parent_id":  parentID,
				"name":       name,
				"depth":      depth,
				"path":       path,
				"deleted_at": nil,
			}).Error; err != nil {
			return fmt.Errorf("failed to update item: %w", err)
		}

		// If this is a folder, restore all descendants and update their paths
		if isFolder && originalPath != path {
			// Get all descendants using recursive CTE
			var descendants []model.Item
			query := `
				WITH RECURSIVE descendants AS (
					SELECT id, parent_id, name, depth, path, is_folder FROM items
					WHERE parent_id = ? AND user_id = ? AND deleted_at IS NOT NULL
					UNION ALL
					SELECT i.id, i.parent_id, i.name, i.depth, i.path, i.is_folder FROM items i
					INNER JOIN descendants d ON i.parent_id = d.id
					WHERE i.deleted_at IS NOT NULL
				)
				SELECT id, parent_id, name, depth, path, is_folder FROM descendants
			`
			if err := tx.Raw(query, id, userID).Scan(&descendants).Error; err != nil {
				return fmt.Errorf("failed to get descendants: %w", err)
			}

			// Restore each descendant and update its path
			for _, desc := range descendants {
				// Calculate new path for this descendant
				newDescPath := strings.Replace(desc.Path, originalPath, path, 1)
				newDescDepth := depth + countPathDepth(desc.Path, originalPath)

				if err := tx.Unscoped().Model(&model.Item{}).
					Where("id = ? AND user_id = ?", desc.ID, userID).
					Updates(map[string]interface{}{
						"parent_id":  desc.ParentID,
						"depth":      newDescDepth,
						"path":       newDescPath,
						"deleted_at": nil,
					}).Error; err != nil {
					return fmt.Errorf("failed to restore descendant %s: %w", desc.ID, err)
				}
			}
		}

		return nil
	})
}

// countPathDepth calculates the depth difference between a descendant path and the original path
func countPathDepth(descPath, originalPath string) int {
	// Count the number of slashes after the original path prefix
	relativePath := strings.TrimPrefix(descPath, originalPath)
	if !strings.HasPrefix(relativePath, "/") {
		return 0
	}
	relativePath = strings.TrimPrefix(relativePath, "/")
	if relativePath == "" {
		return 0
	}
	return strings.Count(relativePath, "/") + 1
}

// GetStorageKeysForPermanentDelete returns all storage keys that need deletion from B2.
func (r *ItemRepository) GetStorageKeysForPermanentDelete(id, userID uuid.UUID) ([]string, error) {
	var keys []string

	item, err := r.FindByIDUnscoped(id, userID)
	if err != nil {
		return nil, err
	}

	if item.StorageKey != nil && *item.StorageKey != "" {
		keys = append(keys, *item.StorageKey)
	}

	if item.IsFolder {
		descendantIDs, err := r.GetTrashDescendantIDs(id, userID)
		if err != nil {
			return nil, err
		}
		if len(descendantIDs) > 0 {
			var descendants []model.Item
			if err := r.db.Unscoped().
				Where("id IN ? AND user_id = ?", descendantIDs, userID).
				Find(&descendants).Error; err != nil {
				return nil, err
			}
			for _, d := range descendants {
				if d.StorageKey != nil && *d.StorageKey != "" {
					keys = append(keys, *d.StorageKey)
				}
			}
		}
	}

	return keys, nil
}

// Search searches for items by name (fuzzy/case-insensitive) for a user.
// Excludes soft-deleted items. Results are sorted by relevance: exact prefix match first,
// then by name length, then alphabetically.
func (r *ItemRepository) Search(userID uuid.UUID, query string, limit int) ([]model.Item, error) {
	var items []model.Item

	// Use ILIKE for case-insensitive matching
	// Sort by relevance: folders first, then by match quality
	// Exact prefix match ranks higher, then contains match
	err := r.db.Where(
		"user_id = ? AND deleted_at IS NULL AND name ILIKE ?",
		userID, "%"+query+"%",
	).
		Order("is_folder DESC, LENGTH(name) ASC, name ASC").
		Limit(limit).
		Find(&items).Error

	return items, err
}

func (r *ItemRepository) ListStarred(userID uuid.UUID) ([]model.Item, error) {
	var items []model.Item
	err := r.db.
		Table("items").
		Select("items.*").
		Joins("INNER JOIN item_stars ON item_stars.item_id = items.id AND item_stars.user_id = ?", userID).
		Where("items.user_id = ? AND items.deleted_at IS NULL", userID).
		Order("item_stars.created_at DESC, items.name ASC").
		Find(&items).Error
	return items, err
}

func (r *ItemRepository) UpsertStar(userID, itemID uuid.UUID) error {
	star := model.ItemStar{
		UserID: userID,
		ItemID: itemID,
	}
	return r.db.
		Where("user_id = ? AND item_id = ?", userID, itemID).
		FirstOrCreate(&star).Error
}

func (r *ItemRepository) DeleteStar(userID, itemID uuid.UUID) error {
	return r.db.Where("user_id = ? AND item_id = ?", userID, itemID).Delete(&model.ItemStar{}).Error
}

func (r *ItemRepository) ListRecentFiles(userID uuid.UUID, limit int) ([]ItemActivityRecord, error) {
	if limit <= 0 {
		limit = 20
	}

	var items []ItemActivityRecord
	err := r.db.
		Table("items").
		Select("items.*, item_activities.last_event_type, item_activities.last_accessed_at").
		Joins("INNER JOIN item_activities ON item_activities.item_id = items.id AND item_activities.user_id = ?", userID).
		Where("items.user_id = ? AND items.deleted_at IS NULL AND items.is_folder = false", userID).
		Order("item_activities.last_accessed_at DESC").
		Limit(limit).
		Scan(&items).Error
	return items, err
}

func (r *ItemRepository) UpsertActivity(userID, itemID uuid.UUID, eventType string, accessedAt time.Time) error {
	activity := model.ItemActivity{
		UserID:         userID,
		ItemID:         itemID,
		LastEventType:  eventType,
		LastAccessedAt: accessedAt,
	}

	return r.db.
		Where("user_id = ? AND item_id = ?", userID, itemID).
		Assign(model.ItemActivity{
			LastEventType:  eventType,
			LastAccessedAt: accessedAt,
		}).
		FirstOrCreate(&activity).Error
}

func (r *ItemRepository) DeleteAccessDataForItem(userID, itemID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND item_id = ?", userID, itemID).Delete(&model.ItemStar{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND item_id = ?", userID, itemID).Delete(&model.ItemActivity{}).Error; err != nil {
			return err
		}
		return nil
	})
}
