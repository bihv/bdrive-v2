# File Properties Modal — Design Spec

## 1. Overview

Add a **right-click → "Thông tin" (Properties)** context menu option for files and folders in the file manager. Displays a modal with item metadata and no actions (actions remain in the context menu).

## 2. UX Specification

### Trigger
- Right-click any file or folder → context menu → select **"Thông tin"**
- Menu option appears at top of context menu, before the divider

### Modal Layout

```
┌─────────────────────────────────────┐
│  Thông tin: <item name>          [×] │
├─────────────────────────────────────┤
│                                     │
│  [Icon 48px]  item.name             │
│               <type label>          │
│                                     │
│  ─── Thông tin ────                │
│                                     │
│  Loại         <value>              │
│  Kích thước   <value>              │
│  Vị trí       <value>              │
│  Ngày tạo     <value>              │
│  Ngày sửa     <value>              │
│  Số mục con   <value> (folders)    │
│                                     │
└─────────────────────────────────────┘
```

### Field Mapping

| Field | File | Folder |
|-------|------|--------|
| Loại | Resolved from `mime_type` (e.g., "PDF document", "JPEG image") | "Thư mục" |
| Kích thước | Formatted bytes (e.g., "2.4 MB") | "—" (not shown) |
| Vị trí | `path` field | `path` field |
| Ngày tạo | `created_at` formatted as `dd/MM/yyyy HH:mm` | same |
| Ngày sửa | `updated_at` formatted as `dd/MM/yyyy HH:mm` | same |
| Số mục con | Not shown | `child_count` + " mục" |

### MIME Type Resolution (Files)

| MIME Pattern | Label |
|---|---|
| `application/pdf` | PDF document |
| `image/*` | `<ext> image` (e.g., "JPEG image") |
| `video/*` | `<ext> video` |
| `audio/*` | Audio file |
| `text/*` | Text file |
| `application/zip`, `application/x-rar` | Archive file |
| `application/*` (others) | File |
| fallback | "File" |

### Interaction

- Modal triggered by `v-model:show` (two-way binding)
- No footer buttons — close via × button or click outside
- Item data comes from existing `Item` type (no API call needed — data already available in context menu handler)

## 3. Component Specification

### File: `frontend/app/components/folders/FileProperties.vue`

**Props:**
```ts
defineProps<{
  show: boolean
  item: Item | null
}>()
```

**Emits:**
```ts
const emit = defineEmits<{
  'update:show': [value: boolean]
}>()
```

**Internal logic:**
- `fileTypeLabel` — computed resolving mime_type to human label
- `formattedSize` — computed formatting bytes
- `formattedDate` — helper formatting ISO date string to `dd/MM/yyyy HH:mm`
- `getItemIcon` — reuse from index.vue (same icon rules)

## 4. Page Integration

### File: `frontend/app/pages/index.vue`

**Changes:**

1. Import `FileProperties` component
2. Add state:
   ```ts
   const showPropertiesModal = ref(false)
   const propertiesTarget = ref<Item | null>(null)
   ```
3. Add "Thông tin" to `contextMenuOptions`:
   ```ts
   { label: 'Thông tin', key: 'properties' },
   ```
4. Handle in `onContextSelect`:
   ```ts
   if (key === 'properties') {
     const item = items.value.find(i => i.id === contextTarget.value?.id)
     propertiesTarget.value = item || null
     showPropertiesModal.value = true
   }
   ```
5. Mount in template:
   ```vue
   <FileProperties
     v-model:show="showPropertiesModal"
     :item="propertiesTarget"
   />
   ```

## 5. Design System Alignment

- Use existing CSS variables (`--color-*`, `--font-size-*`, `--radius-*`, `--transition-*`)
- Match Naive UI `n-modal` preset "card" styling
- Consistent spacing (use `.5rem` / `1rem` grid)
- Mobile responsive text sizing
- No custom animations — rely on Naive UI default modal transition

## 6. Files Summary

| File | Action |
|------|--------|
| `frontend/app/components/folders/FileProperties.vue` | Create |
| `frontend/app/pages/index.vue` | Modify |
