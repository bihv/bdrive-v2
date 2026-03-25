# Search Palette — Design Spec

## 1. Concept & Vision

Một overlay palette cho phép người dùng tìm kiếm nhanh file/folder theo tên bất kỳ đâu trong hệ thống, với trải nghiệm giống **Notion Quick Search** — real-time, nhẹ nhàng, và luôn sẵn sàng qua phím tắt `Cmd+K`. Cảm giác sử dụng: gọn gàng, thông minh, không gây gián đoạn.

---

## 2. Design Language

### 2.1 Aesthetic Direction

**Notion-inspired command palette** — thiết kế tối giản, nền trắng/sáng tối tân, focus hoàn toàn vào nội dung tìm kiếm. Overlay backdrop mờ blur tạo chiều sâu mà không che khuất context.

### 2.2 Color Palette

#### Light Mode

| Token | Hex | Usage |
|---|---|---|
| `--sp-bg` | `#ffffff` | Modal background |
| `--sp-border` | `#e5e5e5` | Border input, divider |
| `--sp-border-focus` | `#3366ff` | Input focus ring |
| `--sp-text-primary` | `#1a1a1a` | Tên file |
| `--sp-text-secondary` | `#888888` | Path breadcrumb |
| `--sp-text-placeholder` | `#aaaaaa` | Placeholder text |
| `--sp-accent` | `#3366ff` | Selected border, icon accent |
| `--sp-highlight` | `#e8efff` | Selected row background |
| `--sp-hover` | `#f5f5f5` | Hover background |
| `--sp-shadow` | `rgba(0, 0, 0, 0.12)` | Modal shadow |

#### Dark Mode

| Token | Hex | Usage |
|---|---|---|
| `--sp-bg` | `#1e1e1e` | Modal background |
| `--sp-border` | `#333333` | Border input, divider |
| `--sp-border-focus` | `#6699ff` | Input focus ring |
| `--sp-text-primary` | `#e8e8e8` | Tên file |
| `--sp-text-secondary` | `#666666` | Path breadcrumb |
| `--sp-text-placeholder` | `#555555` | Placeholder text |
| `--sp-accent` | `#6699ff` | Selected border, icon accent |
| `--sp-highlight` | `#1a2844` | Selected row background |
| `--sp-hover` | `#2a2a2a` | Hover background |
| `--sp-shadow` | `rgba(0, 0, 0, 0.4)` | Modal shadow |

### 2.3 Typography

- **Font**: Inter (Google Fonts), fallback: `-apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif`
- **Modal title/input**: 16px, weight 400, line-height 1.5
- **Result name**: 14px, weight 500, line-height 1.4
- **Result path**: 12px, weight 400, line-height 1.4
- **Footer hint**: 12px, weight 400, color `--sp-text-secondary`

### 2.4 Spacing & Layout

- **Modal padding**: 16px
- **Input height**: 48px (padding 12px vertical)
- **Result item height**: 52px (padding 10px vertical, 12px horizontal)
- **Result item gap**: 2px (border-radius từng item)
- **Modal max-width**: 560px
- **Modal max-height**: 480px (results area: 360px)
- **Border radius**: 12px (modal), 8px (input), 6px (result items)

### 2.5 Motion

| Animation | Duration | Easing | Description |
|---|---|---|---|
| Modal appear | 250ms | `cubic-bezier(0.16, 1, 0.3, 1)` | translateY(8px→0) + opacity(0→1) |
| Modal disappear | 200ms | `ease-in` | translateY(8px) + opacity(1→0) |
| Backdrop appear | 200ms | `ease-out` | opacity(0→1) |
| Result hover | 150ms | `ease-out` | background-color transition |
| Selected change | 100ms | `ease-out` | background-color + border-color |
| Loading spinner | 600ms | `linear` | continuous rotation |

---

## 3. Layout & Structure

### 3.1 Overlay Architecture

```
App Overlay (z-index: 9998)
└── SearchPalette Overlay (z-index: 9999)
    ├── Backdrop (position: fixed, inset: 0)
    │   └── backdrop-filter: blur(8px)
    │   └── background: rgba(0, 0, 0, 0.4)
    └── Modal Container (centered, max 560×480)
        ├── Header (search input)
        ├── Results List (scrollable, virtualized)
        └── Footer (count + shortcuts)
```

### 3.2 Modal Layout (560×480)

```
┌─────────────────────────────────────────────────────┐
│  ┌───────────────────────────────────────────────┐  │
│  │ 🔍  Tìm kiếm file hoặc thư mục...          │  │ ← 16px padding
│  └───────────────────────────────────────────────┘  │
│  ─────────────────────────────────────────────────  │ ← divider 1px
│                                                     │
│  📄  Tài liệu dự án.txt                           │ ← 52px height
│  ─────────────────────────────────────────────────  │
│  📁  Báo cáo tháng 3                              │
│  ─────────────────────────────────────────────────  │
│  📄  Tài liệu dự án (2).md                        │
│  ─────────────────────────────────────────────────  │
│  📁  Tai lieu du an                               │
│                                                     │
│  ─────────────────────────────────────────────────  │ ← divider
│  3 kết quả  •  ↵ để mở  •  esc để đóng          │ ← 36px height
└─────────────────────────────────────────────────────┘
```

### 3.3 Responsive Strategy

- **Desktop (≥768px)**: Modal 560×480px, canh giữa, margin 0
- **Mobile (<768px)**: Modal full-width với margin 16px mỗi bên, height 60vh, border-radius 16px

---

## 4. Features & Interactions

### 4.1 Opening the Palette

- **Keyboard**: `Cmd+K` (macOS) / `Ctrl+K` (Windows/Linux)
- **UI Button**: Icon search trên `FileManagerHeader.vue`
- **Behavior**: Focus immediately vào input, select all text hiện tại (nếu có)

### 4.2 Search Behavior

- **Trigger**: Real-time, mỗi ký tự gõ
- **Debounce**: 250ms sau khi ngừng gõ mới call API (prevent spam)
- **Minimum query**: 1 ký tự mới bắt đầu tìm kiếm
- **Loading state**: Spinner nhỏ bên trong input (bên phải) khi đang fetch
- **Cancel**: Nếu user tiếp tục gõ trong khi fetch, cancel request cũ

### 4.3 Results Display

- **Fuzzy search**: Backend thực hiện fuzzy matching theo tên file/folder
- **Max results**: 20 kết quả mỗi lần tìm
- **Highlight matching**: Tô đậm phần text match trong tên file
- **Path display**: Hiện breadcrumb path ngắn gọn (e.g., `Drive / Thư mục cha`)
- **Icon**: FileIcon component dựa trên MIME type hoặc folder icon

### 4.4 Keyboard Navigation

| Key | Action |
|---|---|
| `↑` Arrow Up | Di chuyển selection lên 1 item |
| `↓` Arrow Down | Di chuyển selection xuống 1 item |
| `Enter` | Navigate đến file (đóng palette + highlight) |
| `Escape` | Đóng palette |
| `Backspace` (input rỗng) | Đóng palette |

- **Boundary**: Không wrap-around (ở item đầu → Up không làm gì, ở cuối → Down không làm gì)
- **Auto-scroll**: Khi selection vượt viewport, tự động scroll để giữ selected item visible

### 4.5 Click Behavior

- **Click vào result**: Navigate đến thư mục chứa file + highlight file đó (tương tự double-click trong file manager)
- **Click backdrop**: Đóng palette

### 4.6 Result States

| State | Visual |
|---|---|
| Default | icon + name (primary) + path (secondary) |
| Hover | nền `--sp-hover`, transition 150ms |
| Selected (keyboard) | nền `--sp-highlight`, viền trái 2px `--sp-accent` |
| Loading | spinner trong input |
| Empty (no results) | "Không tìm thấy kết quả nào" + search icon mờ |
| Empty (initial) | "Nhập từ khóa để tìm kiếm" |

---

## 5. Component Inventory

### 5.1 `SearchPalette.vue` (Root)

**Props**: none (global singleton)
**Emits**: none
**State**:
- `isOpen: boolean`
- `query: string`
- `results: SearchResult[]`
- `selectedIndex: number`
- `isLoading: boolean`
- `error: string | null`

**Behavior**:
- Mounted: register global keyboard listener (`Cmd+K`)
- On open: reset query, results, selectedIndex; focus input
- On close: un-focus input, reset state
- On `selectedIndex` change: scroll into view if needed

### 5.2 `SearchInput.vue`

**Props**: `modelValue: string`, `loading: boolean`, `placeholder: string`
**Emits**: `update:modelValue`, `submit`, `clear`
**States**: default, focused (ring border `--sp-border-focus`), filled

**Visual**:
```
┌─────────────────────────────────────────────────────┐
│ 🔍  [placeholder text........................]  [⏳] │  ← ⏳ only when loading
└─────────────────────────────────────────────────────┘
```
- Left: search icon (mdi:magnify, 18px, `--sp-text-placeholder`)
- Right: clear button (X) khi có text, spinner khi loading

### 5.3 `SearchResults.vue`

**Props**: `results: SearchResult[]`, `selectedIndex: number`
**Emits**: `select(index: number)`
**Behavior**: Virtual scroll nếu results > 20

### 5.4 `SearchItem.vue`

**Props**: `item: SearchResult`, `selected: boolean`, `searchQuery: string`
**Emits**: `click`, `mouseenter`
**Visual**:
```
┌─────────────────────────────────────────────────────┐
│ [icon]  [Tài liệu dự án]     Drive / Thư mục A    │  ← name bold match
└─────────────────────────────────────────────────────┘
```
- Icon: 20×20, dùng `FileIcon` component
- Name: 14px weight 500, phần match tô đậm
- Path: 12px weight 400, `--sp-text-secondary`, right-aligned
- Selected: nền `--sp-highlight`, viền trái 2px `--sp-accent`

### 5.5 `SearchFooter.vue`

**Props**: `resultCount: number`, `query: string`
**Visual**: "X kết quả  •  ↵ để mở  •  esc để đóng"
- Ẩn khi không có query hoặc không có kết quả

---

## 6. Technical Approach

### 6.1 File Structure

```
frontend/app/
├── components/
│   └── search/
│       ├── SearchPalette.vue    ← Root component (overlay)
│       ├── SearchInput.vue      ← Input field
│       ├── SearchResults.vue    ← Virtual scroll list
│       ├── SearchItem.vue       ← Single result row
│       └── SearchFooter.vue     ← Hint footer
├── composables/
│   └── useSearchPalette.ts      ← State + keyboard + API logic
├── styles/
│   └── search-palette.css        ← CSS variables + animations
```

### 6.2 Backend API

**Endpoint**: `GET /api/v1/items/search`

**Query params**:
| Param | Type | Description |
|---|---|---|
| `q` | `string` | Search query, min 1 char |
| `limit` | `number` | Max results (default: 20) |

**Response**:
```json
{
  "success": true,
  "data": [
    {
      "id": "item_123",
      "name": "Tài liệu dự án.txt",
      "type": "file",
      "mimeType": "text/plain",
      "parentId": "folder_456",
      "path": "Drive / Thư mục A"
    },
    {
      "id": "folder_789",
      "name": "Báo cáo tháng 3",
      "type": "folder",
      "parentId": "folder_456",
      "path": "Drive / Thư mục A"
    }
  ]
}
```

### 6.3 Composable Logic (`useSearchPalette.ts`)

```typescript
// State
const isOpen = ref(false)
const query = ref('')
const results = ref<SearchResult[]>([])
const selectedIndex = ref(0)
const isLoading = ref(false)

// Keyboard registration
onMounted(() => {
  document.addEventListener('keydown', handleGlobalKeydown)
})
onUnmounted(() => {
  document.removeEventListener('keydown', handleGlobalKeydown)
})

// Global keydown handler
const handleGlobalKeydown = (e: KeyboardEvent) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault()
    open()
  }
}

// Search with debounce
const search = useDebounceFn(async (q: string) => {
  if (!q.trim()) { results.value = []; return }
  isLoading.value = true
  try {
    const data = await $fetch('/api/v1/items/search', { params: { q, limit: 20 } })
    results.value = data
    selectedIndex.value = 0
  } finally {
    isLoading.value = false
  }
}, 250)

// Keyboard navigation in modal
const handleModalKeydown = (e: KeyboardEvent) => {
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      selectedIndex.value = Math.min(selectedIndex.value + 1, results.value.length - 1)
      break
    case 'ArrowUp':
      e.preventDefault()
      selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
      break
    case 'Enter':
      e.preventDefault()
      navigateToResult(results.value[selectedIndex.value])
      break
    case 'Escape':
      e.preventDefault()
      close()
      break
  }
}

// Navigate action
const navigateToResult = async (item: SearchResult) => {
  close()
  // 1. Navigate to parent folder via folder store
  // 2. Wait for view to update
  // 3. Highlight the file in the view
  await folderStore.navigateTo(item.parentId)
  folderStore.setHighlightedId(item.id)
}
```

### 6.4 Folder Store Extension

Update `frontend/app/stores/folder.ts`:
- Add `highlightedId: string | null` state
- Add `setHighlightedId(id: string | null)` action
- `FileManagerGrid.vue` / `FileManagerList.vue`: highlight item nếu `id === highlightedId`

### 6.5 Global Registration

In `frontend/app/app.vue` hoặc `default.vue` layout:

```vue
<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
  <SearchPalette />
</template>
```

---

## 7. Edge Cases

| Case | Behavior |
|---|---|
| Query quá ngắn (<1 char) | Không call API, hiện "Nhập từ khóa..." |
| API error | Hiện "Đã xảy ra lỗi khi tìm kiếm" trong results |
| No results | Hiện empty state "Không tìm thấy kết quả nào" |
| Results > viewport | Virtual scroll, chỉ render items visible |
| Network offline | Hiện "Không có kết quả" kèm icon mờ |
| Click backdrop | Đóng palette, không navigate |
| Rapid typing | Debounce 250ms, cancel previous request |
| Folder item clicked | Navigate vào folder (mở folder) |
| File item clicked | Navigate đến folder + highlight file |

---

## 8. Accessibility

- **Focus trap**: Khi palette mở, Tab chỉ di chuyển trong palette (input → results → footer)
- **ARIA**: `role="dialog"`, `aria-label="Tìm kiếm file và thư mục"`, `aria-modal="true"`
- **Results**: Mỗi item có `role="option"`, container có `role="listbox"`, `aria-activedescendant` trỏ item đang chọn
- **Screen reader**: Đọc số kết quả, tên item khi navigate
- **Reduced motion**: Nếu `prefers-reduced-motion`, bỏ animation (instant show/hide)

---

## 9. Non-Goals (Out of Scope)

- Search theo nội dung file (chỉ tên)
- Search trong trash
- Recent files list (có thể thêm sau)
- Filters (type, date, size)
- Thumbnail preview trong results
