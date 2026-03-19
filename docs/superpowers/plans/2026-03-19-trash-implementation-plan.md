# Trash Feature Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Thêm tính năng thùng rác cho phép xem, khôi phục và xóa vĩnh viễn các item đã xóa.

**Architecture:** Backend thêm 3 API endpoints cho trash (list, restore, permanent delete) sử dụng GORM Unscoped queries. Frontend thêm trash view mode qua query param, thùng rác nằm trên sidebar, user info dời xuống cuối sidebar.

**Tech Stack:** Go/Gin, GORM, Backblaze B2, Vue 3/TypeScript, Pinia, Naive UI

---

## Chunk 1: Backend - Repository Layer

### Task 1: Thêm repository methods

**Files:**
- Modify: `backend/internal/repository/item_repository.go:183`

- [ ] **Step 1: Thêm FindTrash method**

```go
// FindTrash returns all soft-deleted items for a user.
func (r *ItemRepository) FindTrash(userID uuid.UUID) ([]model.Item, error) {
	var items []model.Item
	err := r.db.Unscoped().
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Order("deleted_at DESC").
		Find(&items).Error
	return items, err
}
```

- [ ] **Step 2: Thêm GetTrashDescendantIDs method**

```go
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
```

- [ ] **Step 3: Thêm PermanentDelete method**

```go
// PermanentDelete hard-deletes an item from the database.
func (r *ItemRepository) PermanentDelete(id, userID uuid.UUID) error {
	return r.db.Unscoped().Where("id = ? AND user_id = ?", id, userID).Delete(&model.Item{}).Error
}
```

- [ ] **Step 4: Thêm Restore method**

```go
// Restore un-deletes an item by setting deleted_at to NULL.
func (r *ItemRepository) Restore(id, userID uuid.UUID) error {
	return r.db.Unscoped().Model(&model.Item{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("deleted_at", nil).Error
}
```

- [ ] **Step 5: Thêm GetStorageKeysForPermanentDelete method**

```go
// GetStorageKeysForPermanentDelete returns all storage keys that need deletion from B2.
func (r *ItemRepository) GetStorageKeysForPermanentDelete(id, userID uuid.UUID) ([]string, error) {
	var keys []string

	// Get the item first
	item, err := r.FindByID(id, userID)
	if err != nil {
		return nil, err
	}

	if item.StorageKey != nil && *item.StorageKey != "" {
		keys = append(keys, *item.StorageKey)
	}

	if item.IsFolder {
		// Get all descendants' storage keys
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
```

- [ ] **Step 6: Thêm UpdatePathAndDepth method cho restore**

```go
// UpdatePathAndDepth updates the path and depth of an item after restore to a new parent.
func (r *ItemRepository) UpdatePathAndDepth(id uuid.UUID, parentID *uuid.UUID, depth int, path string) error {
	return r.db.Unscoped().Model(&model.Item{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"parent_id": parentID,
			"depth":     depth,
			"path":      path,
		}).Error
}
```

- [ ] **Step 7: Commit**

```bash
cd /Volumes/Data2/project/home/bdrive
git add backend/internal/repository/item_repository.go
git commit -m "feat(backend): add trash repository methods"
```

---

## Chunk 2: Backend - DTO Layer

### Task 2: Thêm DTOs cho trash

**Files:**
- Modify: `backend/internal/dto/item_dto.go:125`

- [ ] **Step 1: Thêm RestoreItemRequest DTO**

```go
// RestoreItemRequest is the request body for restoring an item from trash.
type RestoreItemRequest struct {
	TargetParentID *string `json:"targetParentID"`
	NewName        *string `json:"newName"`
}
```

- [ ] **Step 2: Thêm TrashItemResponse DTO (extends ItemResponse)**

```go
// TrashItemResponse is the API response for a trash item.
type TrashItemResponse struct {
	ItemResponse
	DeletedAt string `json:"deleted_at"`
}
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/dto/item_dto.go
git commit -m "feat(backend): add trash DTOs"
```

---

## Chunk 3: Backend - Service Layer

### Task 3: Thêm service methods

**Files:**
- Modify: `backend/internal/service/item_service.go:682`

- [ ] **Step 1: Thêm ListTrash method**

```go
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
		resp := dto.ToItemResponse(&item, 0)
		responses = append(responses, dto.TrashItemResponse{
			ItemResponse: *resp,
			DeletedAt:    deletedAt,
		})
	}

	return responses, nil
}
```

- [ ] **Step 2: Thêm RestoreItem method**

```go
// RestoreItem restores an item from trash.
func (s *ItemService) RestoreItem(id, userID uuid.UUID, req *dto.RestoreItemRequest) (*dto.ItemResponse, error) {
	// Find item (including deleted)
	var item model.Item
	err := s.itemRepo.FindByID(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("item not found")
		}
		return nil, fmt.Errorf("internal error")
	}

	// Verify item is actually deleted
	if item.DeletedAt.Valid == false {
		return nil, fmt.Errorf("item is not in trash")
	}

	var targetParentID *uuid.UUID
	var targetPath string
	var targetDepth int

	// Determine target parent
	if req != nil && req.TargetParentID != nil {
		pid, err := uuid.Parse(*req.TargetParentID)
		if err != nil {
			return nil, fmt.Errorf("invalid parent ID")
		}
		targetParentID = &pid

		// Verify parent exists and is not deleted
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
		// Check parent is not deleted
		if parent.DeletedAt.Valid {
			return nil, &TrashError{Code: "PARENT_DELETED", Message: "target parent is in trash"}
		}
		targetPath = parent.Path + "/" + item.Name
		targetDepth = parent.Depth + 1
	} else {
		// Use original parent
		if item.ParentID != nil {
			parent, err := s.itemRepo.FindByID(*item.ParentID, userID)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, &TrashError{Code: "PARENT_DELETED", Message: "original parent not found"}
				}
				return nil, fmt.Errorf("internal error")
			}
			// Check parent is not deleted
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

	// Determine name
	newName := item.Name
	if req != nil && req.NewName != nil && *req.NewName != "" {
		newName = *req.NewName
		targetPath = getPathPrefix(targetPath) + "/" + newName
	}

	// Check for name conflict in target parent
	exists, err := s.itemRepo.NameExistsInParent(userID, targetParentID, newName, &id)
	if err != nil {
		return nil, fmt.Errorf("internal error")
	}
	if exists {
		return nil, &TrashError{Code: "NAME_CONFLICT", Message: "an item with this name already exists"}
	}

	// Restore in transaction
	err = s.itemRepo.RestoreInTransaction(id, userID, targetParentID, newName, targetDepth, targetPath)
	if err != nil {
		s.log.Error("Failed to restore item", zap.Error(err))
		return nil, fmt.Errorf("failed to restore item")
	}

	// Fetch restored item
	restored, err := s.itemRepo.FindByID(id, userID)
	if err != nil {
		return nil, fmt.Errorf("internal error")
	}

	return ToItemResponse(restored, 0), nil
}

// TrashError represents a trash-specific error.
type TrashError struct {
	Code    string
	Message string
}

func (e *TrashError) Error() string {
	return e.Message
}
```

- [ ] **Step 3: Thêm PermanentDeleteItem method**

```go
// PermanentDeleteItem permanently deletes an item from trash (DB + B2).
func (s *ItemService) PermanentDeleteItem(id, userID uuid.UUID) error {
	// Get storage keys to delete from B2
	keys, err := s.itemRepo.GetStorageKeysForPermanentDelete(id, userID)
	if err != nil {
		s.log.Error("Failed to get storage keys", zap.Error(err))
		return fmt.Errorf("internal error")
	}

	ctx := context.Background()
	var b2Errors []string

	// Delete from B2
	for _, key := range keys {
		if err := s.b2Client.DeleteFile(ctx, key); err != nil {
			s.log.Warn("Failed to delete B2 file", zap.String("key", key), zap.Error(err))
			b2Errors = append(b2Errors, key)
		}
	}

	// If B2 is completely unavailable, return error
	if len(keys) > 0 && len(b2Errors) == len(keys) {
		return fmt.Errorf("storage service unavailable")
	}

	// Log partial failures but continue
	if len(b2Errors) > 0 {
		s.log.Warn("Some B2 files failed to delete",
			zap.Int("total", len(keys)),
			zap.Int("failed", len(b2Errors)))
	}

	// Hard delete from DB
	if err := s.itemRepo.PermanentDelete(id, userID); err != nil {
		s.log.Error("Failed to permanent delete from DB", zap.Error(err))
		return fmt.Errorf("failed to delete item")
	}

	return nil
}
```

- [ ] **Step 4: Thêm RestoreInTransaction method vào repository**

Cập nhật `item_repository.go`:

```go
// RestoreInTransaction restores an item and optionally moves it to a new parent.
func (r *ItemRepository) RestoreInTransaction(id, userID uuid.UUID, parentID *uuid.UUID, name string, depth int, path string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update name, path, depth, and restore
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
		return nil
	})
}
```

- [ ] **Step 5: Thêm helper getPathPrefix**

```go
func getPathPrefix(fullPath string) string {
	lastSlash := strings.LastIndex(fullPath, "/")
	if lastSlash <= 0 {
		return ""
	}
	return fullPath[:lastSlash]
}
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/service/item_service.go backend/internal/repository/item_repository.go
git commit -m "feat(backend): add trash service methods"
```

---

## Chunk 4: Backend - Handler & Router Layer

### Task 4: Thêm handlers và routes

**Files:**
- Modify: `backend/internal/handler/item_handler.go:496`
- Modify: `backend/internal/router/router.go:67`

- [ ] **Step 1: Thêm ListTrash handler**

```go
// ListTrash handles GET /api/v1/trash
func (h *ItemHandler) ListTrash(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	items, err := h.itemService.ListTrash(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Success: false, Error: "Failed to list trash", Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    items,
	})
}
```

- [ ] **Step 2: Thêm RestoreItem handler**

```go
// RestoreItem handles POST /api/v1/trash/:id/restore
func (h *ItemHandler) RestoreItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	var req dto.RestoreItemRequest
	if body := c.Body(); len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
				Success: false, Error: "Invalid request body", Code: "INVALID_PARAM",
			})
		}
	}

	item, err := h.itemService.RestoreItem(id, userID, &req)
	if err != nil {
		if trashErr, ok := err.(*service.TrashError); ok {
			if trashErr.Code == "PARENT_DELETED" || trashErr.Code == "NAME_CONFLICT" {
				return c.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{
					Success: false, Error: trashErr.Message, Code: trashErr.Code,
				})
			}
		}
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    item,
	})
}
```

- [ ] **Step 3: Thêm PermanentDeleteItem handler**

```go
// PermanentDeleteItem handles DELETE /api/v1/trash/:id
func (h *ItemHandler) PermanentDeleteItem(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid user", Code: "UNAUTHORIZED",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Success: false, Error: "Invalid item ID", Code: "INVALID_PARAM",
		})
	}

	if err := h.itemService.PermanentDeleteItem(id, userID); err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "item not found" {
			status = fiber.StatusNotFound
		}
		if err.Error() == "storage service unavailable" {
			status = fiber.StatusServiceUnavailable
		}
		return c.Status(status).JSON(dto.ErrorResponse{
			Success: false, Error: err.Error(), Code: "INTERNAL_ERROR",
		})
	}

	return c.JSON(dto.SuccessResponse{
		Success: true,
		Data:    fiber.Map{"message": "Item permanently deleted"},
	})
}
```

- [ ] **Step 4: Thêm import json và sửa handler struct**

Thêm vào item_handler.go imports:

```go
"encoding/json"
```

- [ ] **Step 5: Thêm routes trong router.go**

Thêm sau các item routes:

```go
// Trash routes (protected)
trash := api.Group("/trash")
trash.Use(middleware.AuthMiddleware(accessSecret))
trash.Get("/", itemHandler.ListTrash)
trash.Post("/:id/restore", itemHandler.RestoreItem)
trash.Delete("/:id", itemHandler.PermanentDeleteItem)
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/item_handler.go backend/internal/router/router.go
git commit -m "feat(backend): add trash handlers and routes"
```

---

## Chunk 5: Backend - Build & Verify

### Task 5: Build backend

- [ ] **Step 1: Build để kiểm tra lỗi**

```bash
cd /Volumes/Data2/project/home/bdrive/backend && go build ./...
```

**Expected:** Build thành công, không có lỗi

- [ ] **Step 2: Nếu có lỗi, fix và commit lại**

---

## Chunk 6: Frontend - Types & Store

### Task 6: Cập nhật types và store

**Files:**
- Modify: `frontend/app/types/folder.ts:91`
- Modify: `frontend/app/stores/folder.ts:96`

- [ ] **Step 1: Thêm deletedAt vào Item type**

```typescript
export interface Item {
    id: string
    parent_id: string | null
    name: string
    is_folder: boolean
    depth: number
    path: string
    mime_type: string | null
    size: number
    color: string | null
    sort_order: number
    child_count: number
    created_at: string
    updated_at: string
    deleted_at?: string  // Thêm dòng này
}
```

- [ ] **Step 2: Thêm RestoreItemRequest type**

```typescript
export interface RestoreItemRequest {
    targetParentID?: string
    newName?: string
}
```

- [ ] **Step 3: Commit types**

```bash
git add frontend/app/types/folder.ts
git commit -m "feat(frontend): add deletedAt and RestoreItemRequest types"
```

- [ ] **Step 4: Cập nhật folder store**

Thêm vào state:

```typescript
state: () => ({
    items: [] as Item[],
    folderTree: [] as FolderTreeNode[],
    currentFolderId: null as string | null,
    currentPath: '/' as string,
    loading: false,
    treeLoading: false,
    trashItems: [] as Item[],        // Thêm
    isTrashView: false as boolean,   // Thêm
    trashLoading: false as boolean,  // Thêm
}),
```

Thêm actions:

```typescript
setTrashItems(items: Item[]) {
    this.trashItems = items
},

setTrashView(v: boolean) {
    this.isTrashView = v
},

setTrashLoading(v: boolean) {
    this.trashLoading = v
},

removeTrashItem(id: string) {
    this.trashItems = this.trashItems.filter(i => i.id !== id)
},

updateTrashItem(updated: Item) {
    const index = this.trashItems.findIndex(i => i.id === updated.id)
    if (index !== -1) {
        this.trashItems[index] = updated
    }
},
```

- [ ] **Step 5: Commit store**

```bash
git add frontend/app/stores/folder.ts
git commit -m "feat(frontend): add trash state and actions to folder store"
```

---

## Chunk 7: Frontend - API Layer

### Task 7: Thêm API methods

**Files:**
- Modify: `frontend/app/composables/useApi.ts:133`

- [ ] **Step 1: Thêm API methods cho trash**

Thêm vào cuối return statement:

```typescript
async function getTrash(): Promise<Item[]> {
    return apiFetch<Item[]>('/api/v1/trash')
}

async function restoreItem(id: string, body?: RestoreItemRequest): Promise<Item> {
    return apiFetch<Item>('/api/v1/trash/' + id + '/restore', {
        method: 'POST',
        body: JSON.stringify(body || {}),
    })
}

async function permanentDeleteItem(id: string): Promise<void> {
    return apiFetch<void>('/api/v1/trash/' + id, {
        method: 'DELETE',
    })
}
```

Cập nhật return:

```typescript
return {
    apiFetch,
    post,
    get,
    uploadToURL,
    refreshToken,
    getTrash,
    restoreItem,
    permanentDeleteItem,
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/composables/useApi.ts
git commit -m "feat(frontend): add trash API methods"
```

---

## Chunk 8: Frontend - Layout (Sidebar)

### Task 8: Cập nhật layout

**Files:**
- Modify: `frontend/app/layouts/default.vue`

- [ ] **Step 1: Thêm trash sidebar item**

Tìm phần `<div class="sidebar-content">` và thêm vào trước `<!-- User info -->`:

```vue
<!-- Trash -->
<div
  class="trash-item"
  :class="{ active: isTrashView }"
  @click="navigateToTrash"
>
  <n-icon size="18"><Icon icon="mdi:delete-outline" /></n-icon>
  <span>Thùng rác</span>
  <n-badge
    v-if="trashItems.length > 0"
    :value="trashItems.length"
    :max="99"
    type="error"
    class="trash-badge"
  />
</div>
```

- [ ] **Step 2: Dời user info xuống cuối sidebar**

Di chuyển `<div class="sidebar-footer">` xuống dưới trash item:

```vue
<!-- Trash -->
<div class="trash-item" ... />

<!-- User info -->
<div class="sidebar-footer">
  ...
</div>
```

- [ ] **Step 3: Thêm computed và methods**

Trong `<script setup>`:

```typescript
import { useFolderStore } from '~/stores/folder'

const folderStore = useFolderStore()
const { isTrashView, trashItems } = storeToRefs(folderStore)

const router = useRouter()
const route = useRoute()

function navigateToTrash() {
    router.push({ path: '/', query: { view: 'trash' } })
}

// Watch route for trash view
watch(() => route.query.view, (view) => {
    folderStore.setTrashView(view === 'trash')
}, { immediate: true })
```

- [ ] **Step 4: Thêm CSS cho trash item**

Thêm vào `<style scoped>`:

```css
.trash-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    color: var(--color-text-secondary);
    transition: all var(--transition-base);
    border-radius: 6px;
    margin: 0.25rem 0.5rem;
    position: relative;
}

.trash-item:hover {
    background: var(--color-surface-hover);
    color: var(--color-text-primary);
}

.trash-item.active {
    background: rgba(234, 84, 85, 0.15);
    color: rgb(234, 84, 85);
}

.trash-badge {
    margin-left: auto;
}
```

- [ ] **Step 5: Thêm CSS cho mobile responsive**

Thêm vào phần `@media (max-width: 768px)`:

```css
.trash-item {
    padding: 0.5rem 0.75rem;
}
```

- [ ] **Step 6: Commit**

```bash
git add frontend/app/layouts/default.vue
git commit -m "feat(frontend): add trash item to sidebar"
```

---

## Chunk 9: Frontend - Trash View Page

### Task 9: Cập nhật index.vue cho trash view

**Files:**
- Modify: `frontend/app/pages/index.vue`

- [ ] **Step 1: Cập nhật script - thêm trash logic**

Thêm imports:

```typescript
import type { Item, RestoreItemRequest, FolderTreeNode } from '~/types/folder'
```

Thêm trash state:

```typescript
const folderStore = useFolderStore()
const { isTrashView, trashItems, trashLoading } = storeToRefs(folderStore)

const api = useApi()
const message = useMessage()
const dialog = useDialog()

// Trash-specific state
const showRestoreDialog = ref(false)
const showRenameDialog = ref(false)
const restoreTarget = ref<{ id: string; name: string } | null>(null)
const newItemName = ref('')
const selectedRestoreFolder = ref<string | null>(null)
```

Thêm computed cho items:

```typescript
const displayItems = computed(() => {
    return isTrashView.value ? trashItems.value : items.value
})

const displayLoading = computed(() => {
    return isTrashView.value ? trashLoading.value : loading.value
})
```

Thêm methods cho trash:

```typescript
async function loadTrash() {
    folderStore.setTrashLoading(true)
    try {
        const data = await api.getTrash()
        folderStore.setTrashItems(data)
    } catch (e: any) {
        message.error(e?.data?.error || 'Không thể tải thùng rác')
    } finally {
        folderStore.setTrashLoading(false)
    }
}

async function handleRestore(item: Item) {
    try {
        await api.restoreItem(item.id)
        message.success('Đã khôi phục')
        folderStore.removeTrashItem(item.id)
    } catch (e: any) {
        const code = e?.data?.code
        if (code === 'PARENT_DELETED') {
            // Show folder selection dialog
            restoreTarget.value = { id: item.id, name: item.name }
            showRestoreDialog.value = true
        } else if (code === 'NAME_CONFLICT') {
            // Show rename dialog
            restoreTarget.value = { id: item.id, name: item.name }
            newItemName.value = item.name
            showRenameDialog.value = true
        } else {
            message.error(e?.data?.error || 'Không thể khôi phục')
        }
    }
}

async function handleRestoreWithFolder(folderId: string | null) {
    if (!restoreTarget.value) return
    try {
        const body: RestoreItemRequest = { targetParentID: folderId || undefined }
        await api.restoreItem(restoreTarget.value.id, body)
        message.success('Đã khôi phục')
        folderStore.removeTrashItem(restoreTarget.value.id)
        showRestoreDialog.value = false
        restoreTarget.value = null
    } catch (e: any) {
        if (e?.data?.code === 'NAME_CONFLICT') {
            newItemName.value = restoreTarget.value.name
            showRenameDialog.value = true
        } else {
            message.error(e?.data?.error || 'Không thể khôi phục')
        }
    }
}

async function handleRestoreWithRename() {
    if (!restoreTarget.value || !newItemName.value.trim()) return
    try {
        const body: RestoreItemRequest = { newName: newItemName.value.trim() }
        await api.restoreItem(restoreTarget.value.id, body)
        message.success('Đã khôi phục')
        folderStore.removeTrashItem(restoreTarget.value.id)
        showRenameDialog.value = false
        restoreTarget.value = null
    } catch (e: any) {
        message.error(e?.data?.error || 'Không thể khôi phục')
    }
}

async function handlePermanentDelete(item: Item) {
    dialog.warning({
        title: 'Xóa vĩnh viễn',
        content: `Bạn có chắc muốn xóa vĩnh viễn "${item.name}"? Hành động này không thể hoàn tác.`,
        positiveText: 'Xóa vĩnh viễn',
        negativeText: 'Hủy',
        onPositiveClick: async () => {
            try {
                await api.permanentDeleteItem(item.id)
                message.success('Đã xóa vĩnh viễn')
                folderStore.removeTrashItem(item.id)
            } catch (e: any) {
                message.error(e?.data?.error || 'Không thể xóa vĩnh viễn')
            }
        },
    })
}

function formatDeletedDate(dateStr: string): string {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return date.toLocaleDateString('vi-VN', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
    })
}
```

- [ ] **Step 2: Cập nhật template - trash header**

Thay thế header:

```vue
<div class="fm-header">
    <div class="fm-breadcrumb">
        <template v-if="isTrashView">
            <span class="trash-title">Thùng rác</span>
        </template>
        <template v-else>
            <n-breadcrumb>
                <n-breadcrumb-item
                    v-for="(crumb, i) in breadcrumbs"
                    :key="i"
                    :clickable="i < breadcrumbs.length - 1"
                    @click="onBreadcrumbClick(i)"
                >
                    {{ crumb.name }}
                </n-breadcrumb-item>
            </n-breadcrumb>
        </template>
    </div>
    <div class="fm-actions">
        <template v-if="isTrashView">
            <n-button
                v-if="trashItems.length > 0"
                type="warning"
                size="small"
                @click="handleEmptyTrash"
            >
                <template #icon>
                    <n-icon><Icon icon="mdi:delete-sweep" /></n-icon>
                </template>
                Dọn thùng rác
            </n-button>
        </template>
        <template v-else>
            <n-button type="primary" size="small" @click="showUploadFile = true">
                <template #icon>
                    <n-icon><Icon icon="mdi:upload" /></n-icon>
                </template>
                Tải lên
            </n-button>
            <n-button type="primary" size="small" @click="showCreateFolder = true">
                <template #icon>
                    <n-icon><Icon icon="mdi:folder-plus" /></n-icon>
                </template>
                Tạo thư mục
            </n-button>
        </template>
    </div>
</div>
```

- [ ] **Step 3: Cập nhật items grid - hiển thị trash items**

Thay thế `<n-spin :show="loading">`:

```vue
<n-spin :show="displayLoading">
    <div v-if="displayItems.length === 0 && !displayLoading" class="fm-empty">
        <n-empty :description="isTrashView ? 'Thùng rác trống' : 'Thư mục trống'">
            <template #icon>
                <n-icon size="48" :depth="3">
                    <Icon :icon="isTrashView ? 'mdi:delete-outline' : 'mdi:folder-open-outline'" />
                </n-icon>
            </template>
        </n-empty>
    </div>

    <div v-else class="fm-grid">
        <div
            v-for="item in displayItems"
            :key="item.id"
            class="fm-item glass-card"
            :class="{ 'is-folder': item.is_folder, 'is-trash': isTrashView }"
            @contextmenu.prevent="isTrashView ? onTrashContext($event, item) : onItemContext($event, item)"
        >
            <div class="fm-item-icon">
                <n-icon size="36" :style="item.color ? { color: item.color } : {}">
                    <Icon :icon="isTrashView ? 'mdi:delete-outline' : getItemIcon(item)" />
                </n-icon>
            </div>
            <div class="fm-item-name">{{ item.name }}</div>
            <div class="fm-item-meta">
                <template v-if="isTrashView">
                    {{ formatDeletedDate(item.deleted_at) }}
                </template>
                <template v-else>
                    <span v-if="item.is_folder">{{ item.child_count }} mục</span>
                    <span v-else>{{ formatSize(item.size) }}</span>
                </template>
            </div>
            <div v-if="isTrashView" class="fm-trash-actions">
                <n-button size="tiny" @click.stop="handleRestore(item)">
                    <n-icon><Icon icon="mdi:restore" /></n-icon>
                    Khôi phục
                </n-button>
                <n-button size="tiny" type="error" @click.stop="handlePermanentDelete(item)">
                    <n-icon><Icon icon="mdi:delete-forever" /></n-icon>
                </n-button>
            </div>
        </div>
    </div>
</n-spin>
```

- [ ] **Step 4: Thêm context menu cho trash**

Thêm sau context menu hiện tại:

```vue
<!-- Trash context menu -->
<n-dropdown
    :show="showTrashContextMenu"
    :x="contextX"
    :y="contextY"
    trigger="manual"
    placement="bottom-start"
    :options="trashContextOptions"
    @select="onTrashContextSelect"
    @clickoutside="showTrashContextMenu = false"
/>
```

Thêm state:

```typescript
const showTrashContextMenu = ref(false)
const trashContextTarget = ref<Item | null>(null)

const trashContextOptions = computed(() => [
    { label: 'Khôi phục', key: 'restore' },
    { type: 'divider', key: 'd1' },
    { label: 'Xóa vĩnh viễn', key: 'permanent-delete' },
])

function onTrashContext(e: MouseEvent, item: Item) {
    contextX.value = e.clientX
    contextY.value = e.clientY
    trashContextTarget.value = item
    showTrashContextMenu.value = true
}

function onTrashContextSelect(key: string) {
    showTrashContextMenu.value = false
    if (!trashContextTarget.value) return
    if (key === 'restore') handleRestore(trashContextTarget.value)
    if (key === 'permanent-delete') handlePermanentDelete(trashContextTarget.value)
}

async function handleEmptyTrash() {
    dialog.warning({
        title: 'Dọn thùng rác',
        content: `Bạn có chắc muốn xóa vĩnh viễn tất cả ${trashItems.value.length} item trong thùng rác?`,
        positiveText: 'Xóa tất cả',
        negativeText: 'Hủy',
        onPositiveClick: async () => {
            for (const item of trashItems.value) {
                try {
                    await api.permanentDeleteItem(item.id)
                } catch (e) {
                    console.error('Failed to delete', item.id, e)
                }
            }
            message.success('Đã dọn thùng rác')
            folderStore.setTrashItems([])
        },
    })
}
```

- [ ] **Step 5: Thêm Restore Folder Dialog**

Thêm sau FolderActions component:

```vue
<!-- Restore Folder Dialog -->
<n-modal
    v-model:show="showRestoreDialog"
    preset="card"
    title="Chọn thư mục khôi phục"
    style="width: 400px"
>
    <div class="restore-folder-list">
        <div
            class="restore-folder-item"
            :class="{ active: selectedRestoreFolder === null }"
            @click="selectedRestoreFolder = null"
        >
            <n-icon><Icon icon="mdi:home-outline" /></n-icon>
            <span>Thư mục gốc</span>
        </div>
        <div
            v-for="folder in folderTree"
            :key="folder.id"
            class="restore-folder-item"
            :class="{ active: selectedRestoreFolder === folder.id }"
            :style="{ paddingLeft: (folder.depth * 20 + 24) + 'px' }"
            @click="selectedRestoreFolder = folder.id"
        >
            <n-icon><Icon icon="mdi:folder" /></n-icon>
            <span>{{ folder.name }}</span>
        </div>
    </div>
    <template #footer>
        <n-space justify="end">
            <n-button @click="showRestoreDialog = false">Hủy</n-button>
            <n-button type="primary" @click="handleRestoreWithFolder(selectedRestoreFolder)">
                Khôi phục
            </n-button>
        </n-space>
    </template>
</n-modal>

<!-- Rename Dialog -->
<n-modal
    v-model:show="showRenameDialog"
    preset="card"
    title="Đổi tên"
    style="width: 400px"
>
    <n-input v-model:value="newItemName" placeholder="Tên mới" />
    <template #footer>
        <n-space justify="end">
            <n-button @click="showRenameDialog = false">Hủy</n-button>
            <n-button type="primary" @click="handleRestoreWithRename">
                Khôi phục
            </n-button>
        </n-space>
    </template>
</n-modal>
```

- [ ] **Step 6: Thêm CSS cho restore dialog**

```css
.restore-folder-list {
    max-height: 300px;
    overflow-y: auto;
}

.restore-folder-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    cursor: pointer;
    border-radius: 6px;
    transition: all var(--transition-base);
}

.restore-folder-item:hover {
    background: var(--color-surface-hover);
}

.restore-folder-item.active {
    background: rgba(64, 158, 255, 0.15);
    color: var(--color-primary);
}

.trash-title {
    font-size: var(--font-size-lg);
    font-weight: 600;
    color: var(--color-text-primary);
}

.fm-item.is-trash {
    padding-bottom: 0.5rem;
}

.fm-trash-actions {
    display: flex;
    gap: 0.25rem;
    margin-top: 0.5rem;
    opacity: 0;
    transition: opacity var(--transition-base);
}

.fm-item.is-trash:hover .fm-trash-actions {
    opacity: 1;
}
```

- [ ] **Step 7: Watch trash view và load data**

Thêm vào cuối script:

```typescript
// Watch for trash view changes
watch(isTrashView, async (isTrash) => {
    if (isTrash) {
        await loadTrash()
    }
}, { immediate: true })
```

- [ ] **Step 8: Commit**

```bash
git add frontend/app/pages/index.vue
git commit -m "feat(frontend): add trash view to index page"
```

---

## Chunk 10: Integration & Test

### Task 10: Test toàn bộ feature

- [ ] **Step 1: Start backend**

```bash
cd /Volumes/Data2/project/home/bdrive/backend
go run ./cmd/server
```

- [ ] **Step 2: Start frontend**

```bash
cd /Volumes/Data2/project/home/bdrive/frontend
npm run dev
```

- [ ] **Step 3: Test các flows:**

1. Xóa một file → file biến mất khỏi view, xuất hiện trong thùng rác
2. Mở thùng rác → thấy file đã xóa với ngày xóa
3. Bấm Khôi phục → file xuất hiện lại ở vị trí cũ
4. Xóa vĩnh viễn → file biến mất hoàn toàn
5. Kiểm tra sidebar "Thùng rác" có badge số lượng
6. Kiểm tra responsive trên mobile

---

## Summary

**Backend files changed:**
- `backend/internal/repository/item_repository.go` - 4 new methods
- `backend/internal/dto/item_dto.go` - 2 new DTOs
- `backend/internal/service/item_service.go` - 3 new methods
- `backend/internal/handler/item_handler.go` - 3 new handlers
- `backend/internal/router/router.go` - 3 new routes

**Frontend files changed:**
- `frontend/app/types/folder.ts` - added `deleted_at`, `RestoreItemRequest`
- `frontend/app/stores/folder.ts` - added trash state and actions
- `frontend/app/composables/useApi.ts` - added trash API methods
- `frontend/app/layouts/default.vue` - added trash sidebar item, moved user info
- `frontend/app/pages/index.vue` - added trash view mode

**Total commits expected:** 10 (1 per task group)
