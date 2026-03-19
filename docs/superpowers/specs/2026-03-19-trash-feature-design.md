# Trash Feature Design

## Status
- **Date**: 2026-03-19
- **Approved**: Yes (user confirmed with "ok")

## Overview

Thêm tính năng thùng rác cho phép người dùng xem, khôi phục và xóa vĩnh viễn các item đã xóa. Item khi xóa sẽ được soft-delete (GORM DeletedAt) để có thể khôi phục. Xóa vĩnh viễn sẽ xóa cả DB record và B2 file.

## Backend

### API Endpoints

| Method | Endpoint | Mô tả |
|--------|----------|--------|
| GET | `/api/v1/trash` | Liệt kê các item trong thùng rác của user |
| POST | `/api/v1/trash/:id/restore` | Khôi phục item từ thùng rác |
| DELETE | `/api/v1/trash/:id` | Xóa vĩnh viễn item (DB + B2 file) |

### Repository Layer (`internal/repository/item_repository.go`)

Thêm methods:

1. **`FindTrash(userID uuid.UUID) ([]model.Item, error)`**
   - Query tất cả item có `deleted_at IS NOT NULL` của user
   - Không filter folder/file — lấy hết
   - Order by `deleted_at DESC`

2. **`GetTrashDescendantIDs(itemID, userID uuid.UUID) ([]uuid.UUID, error)`**
   - Tương tự `GetAllDescendantIDs` nhưng query trên các item đã soft-delete (`deleted_at IS NOT NULL`)
   - Dùng recursive CTE

3. **`PermanentDelete(id, userID uuid.UUID) error`**
   - Hard delete: `UNSCOPED` delete
   - Xóa record vĩnh viễn khỏi DB

4. **`Restore(id, userID uuid.UUID) error`**
   - Đặt `deleted_at = NULL` để khôi phục

### Service Layer (`internal/service/item_service.go`)

Thêm methods:

1. **`ListTrash(userID uuid.UUID) ([]model.Item, error)`**
   - Gọi `itemRepo.FindTrash(userID)`
   - Trả về danh sách item đã xóa

2. **`RestoreItem(id, userID uuid.UUID, targetParentID *uuid.UUID, newName *string) error`**
   - Validate item tồn tại và đã bị xóa
   - Nếu là folder: khôi phục folder + tất cả descendants (recursive)
   - Nếu là file: khôi phục file đó
   - Nếu `targetParentID` được truyền → đặt parent = targetParentID, recalculate path
   - Nếu `newName` được truyền → đổi tên trước khi restore
   - Kiểm tra tên trùng → trả `NAME_CONFLICT` error
   - Dùng transaction để đảm bảo consistency

3. **`PermanentDeleteItem(id, userID uuid.UUID) error`**
   - Validate item tồn tại và đã bị xóa
   - Query tất cả storage_keys cần xóa trên B2:
     - Nếu folder: lấy storage_keys của folder + tất cả descendants
     - Nếu file: lấy storage_key của file đó
   - Xóa file trên B2 (sequential, log error nhưng không rollback nếu 1 file thất bại)
   - Hard delete record trong DB
   - Dùng transaction: nếu B2 xóa thất bại → rollback DB

### DTOs

Request:
```go
type RestoreItemRequest struct {
    TargetParentID *string `json:"targetParentID"`
    NewName         *string `json:"newName"`
}
```

Response (success):
```go
// Dùng lại dto.ItemResponse như các endpoint khác
```

Error codes:
- `409 Conflict` + `PARENT_DELETED`: parent không tồn tại, frontend show modal chọn folder
- `409 Conflict` + `NAME_CONFLICT`: tên bị trùng, frontend show modal đổi tên

Thêm handlers:

1. **`ListTrash`** — GET `/api/v1/trash`
   - Parse userID từ token
   - Gọi service.ListTrash
   - Trả về danh sách item

2. **`RestoreItem`** — POST `/api/v1/trash/:id/restore`
   - Parse id và userID
   - Nhận request body: `{ targetParentID?: string, newName?: string }`
   - Gọi service.RestoreItem với các optional params
   - Nếu parent không tồn tại → trả `409 Conflict` với `code: "PARENT_DELETED"`
   - Nếu tên trùng → trả `409 Conflict` với `code: "NAME_CONFLICT"`
   - Trả về item đã khôi phục

3. **`PermanentDeleteItem`** — DELETE `/api/v1/trash/:id`
   - Parse id và userID
   - Gọi service.PermanentDeleteItem
   - Trả về success

### Router (`internal/router/router.go`)

Thêm routes trong nhóm `/api/v1`:
- `GET /trash` → `ListTrash`
- `POST /trash/:id/restore` → `RestoreItem`
- `DELETE /trash/:id` → `PermanentDeleteItem`

## Frontend

### API Layer (`frontend/app/composables/useApi.ts`)

Thêm methods:

```typescript
// Lấy danh sách thùng rác
async function getTrash(): Promise<Item[]>

// Khôi phục item
async function restoreItem(id: string): Promise<void>

// Xóa vĩnh viễn
async function permanentDeleteItem(id: string): Promise<void>
```

### Store (`frontend/app/stores/folder.ts`)

Thêm state và actions:

```typescript
state: {
    trashItems: [] as Item[],
    isTrashView: false as boolean,
    trashLoading: false as boolean,
}

actions: {
    setTrashItems(items: Item[]): void
    setTrashView(v: boolean): void
    removeTrashItem(id: string): void
    updateTrashItem(updated: Item): void
}
```

### Layout (`frontend/app/layouts/default.vue`)

Thêm sidebar item "Thùng rác" giữa folder tree và user info:

```
☁️ 1Drive
──────────────
Thư mục
  🏠 Tất cả file
  📁 Folder A
────────────────
🗑️ Thùng rác          ← thêm vào đây
────────────────
👤 User Name           ← dời xuống footer
   user@email.com
   [Logout]
```

- Click "Thùng rác" → navigate với query param `/?view=trash`
- User info dời xuống footer của sidebar (dưới trash item)

### Page (`frontend/app/pages/index.vue`)

Thêm trash view mode:

```typescript
// Nếu route có query view=trash
// → hiển thị trashItems thay vì items
// → thêm header actions: "Khôi phục tất cả", "Dọn thùng rác"
```

### Types (`frontend/app/types/folder.ts`)

Thêm `deletedAt` vào Item type (để hiển thị ngày xóa):

```typescript
interface Item {
    // ... existing fields
    deleted_at?: string
}
```

## Data Flow

### Xóa item (soft-delete — hiện tại, không thay đổi)
- User bấm xóa → `DELETE /api/v1/items/:id`
- GORM tự động set `deleted_at = NOW()` (soft delete)
- Item biến mất khỏi view thường, xuất hiện trong trash

### Khôi phục
- User bấm "Khôi phục" → `POST /api/v1/trash/:id/restore`
- Backend kiểm tra parent có tồn tại và chưa bị xóa không:
  - **Parent còn**: restore trực tiếp (`deleted_at = NULL`), không cần API call thêm
  - **Parent đã bị xóa hoặc không tồn tại**: trả về error code `PARENT_DELETED`, frontend hiển thị modal chọn thư mục đích
- Khi tên bị trùng ở thư mục đích: trả về error code `NAME_CONFLICT`, frontend hiển thị modal đổi tên
- Backend `RestoreItem` nhận optional `targetParentID` và `newName`:
  - Nếu `targetParentID` được truyền → khôi phục vào thư mục đó
  - Nếu `newName` được truyền → đổi tên trước khi restore
  - Update `path` và `depth` tương ứng sau khi khôi phục

### Xóa vĩnh viễn
- User bấm "Xóa vĩnh viễn" → `DELETE /api/v1/trash/:id`
- Backend:
  1. Query tất cả storage_keys
  2. Gọi B2 `DeleteFile` cho từng storage_key
  3. Hard delete DB records
- Item biến mất hoàn toàn

## Restore Logic Details

### Case 1: Parent còn tồn tại và không bị xóa
- Gọi `POST /api/v1/trash/:id/restore`
- Backend: `deleted_at = NULL`
- Path giữ nguyên

### Case 2: Parent đã bị xóa (soft-deleted) hoặc không tồn tại
- Backend trả về `409 Conflict` với `code: "PARENT_DELETED"`
- Frontend hiển thị modal "Chọn thư mục khôi phục" với folder tree
- User chọn thư mục đích → gọi lại `POST /api/v1/trash/:id/restore` với `targetParentID`
- Backend: `parent_id = targetParentID`, recalculate `path` và `depth`

### Case 3: Tên bị trùng ở thư mục đích
- Backend trả về `409 Conflict` với `code: "NAME_CONFLICT"`
- Frontend hiển thị modal đổi tên, input mặc định là tên hiện tại
- User nhập tên mới → gọi lại `POST /api/v1/trash/:id/restore` với `newName`
- Backend: kiểm tra tên mới không trùng, đổi tên và restore

## Error Handling

- B2 DeleteFile fail → log error nhưng vẫn tiếp tục xóa các file khác. Sau cùng, hard delete DB nếu không có lỗi nghiêm trọng.
- Nếu B2 hoàn toàn unavailable → trả error 503, không xóa DB
- Restore khi parent đã bị xóa → trả `PARENT_DELETED`, frontend hiển thị modal chọn thư mục
- Restore khi tên trùng → trả `NAME_CONFLICT`, frontend hiển thị modal đổi tên
