# File Properties Modal — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan.

**Goal:** Add right-click → "Thông tin" context menu option that opens a modal displaying file/folder metadata.

**Architecture:** New Vue component `FileProperties.vue` mounted in `index.vue`. Item data already available in context menu handler — no API changes needed.

**Tech Stack:** Nuxt 4, Vue 3, Naive UI, TypeScript

---

## File Map

| File | Action |
|------|--------|
| `frontend/app/components/folders/FileProperties.vue` | Create |
| `frontend/app/pages/index.vue` | Modify |

---

## Chunk 1: Create FileProperties.vue

**Files:**
- Create: `frontend/app/components/folders/FileProperties.vue`

- [ ] **Step 1: Create the component file**

```vue
<template>
  <n-modal
    :show="show"
    preset="card"
    :title="modalTitle"
    style="width: 420px; max-width: 90vw;"
    :mask-closable="true"
    @update:show="emit('update:show', $event)"
  >
    <div class="properties-content">
      <!-- Header: Icon + Name -->
      <div class="properties-header">
        <n-icon :size="48" :style="item?.color ? { color: item.color } : {}">
          <Icon :icon="getItemIcon(item)" />
        </n-icon>
        <div class="properties-name-block">
          <div class="properties-name">{{ item?.name }}</div>
          <div class="properties-type">{{ fileTypeLabel }}</div>
        </div>
      </div>

      <!-- Divider -->
      <div class="properties-section-label">Thông tin</div>

      <!-- Info rows -->
      <div class="properties-rows">
        <div class="properties-row">
          <span class="properties-row-label">Loại</span>
          <span class="properties-row-value">{{ fileTypeLabel }}</span>
        </div>

        <div v-if="!item?.is_folder" class="properties-row">
          <span class="properties-row-label">Kích thước</span>
          <span class="properties-row-value">{{ formattedSize }}</span>
        </div>

        <div class="properties-row">
          <span class="properties-row-label">Vị trí</span>
          <span class="properties-row-value path-value">{{ item?.path || '—' }}</span>
        </div>

        <div class="properties-row">
          <span class="properties-row-label">Ngày tạo</span>
          <span class="properties-row-value">{{ formattedCreatedAt }}</span>
        </div>

        <div class="properties-row">
          <span class="properties-row-label">Ngày sửa</span>
          <span class="properties-row-value">{{ formattedUpdatedAt }}</span>
        </div>

        <div v-if="item?.is_folder" class="properties-row">
          <span class="properties-row-label">Số mục con</span>
          <span class="properties-row-value">{{ item.child_count }} mục</span>
        </div>
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { Item } from '~/types/folder'

const props = defineProps<{
  show: boolean
  item: Item | null
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

const modalTitle = computed(() => {
  return props.item ? `Thông tin: ${props.item.name}` : 'Thông tin'
})

function getItemIcon(item: Item | null): string {
  if (!item) return 'mdi:file-outline'
  if (item.is_folder) return 'mdi:folder'
  if (item.mime_type?.startsWith('image/')) return 'mdi:file-image'
  if (item.mime_type?.startsWith('video/')) return 'mdi:file-video'
  if (item.mime_type === 'application/pdf') return 'mdi:file-pdf-box'
  return 'mdi:file-document-outline'
}

const fileTypeLabel = computed(() => {
  if (!props.item) return ''
  if (props.item.is_folder) return 'Thư mục'

  const mime = props.item.mime_type || ''

  if (mime === 'application/pdf') return 'PDF document'
  if (mime.startsWith('image/')) {
    const ext = mime.split('/')[1]?.toUpperCase() || 'IMAGE'
    return `${ext} image`
  }
  if (mime.startsWith('video/')) {
    const ext = mime.split('/')[1]?.toUpperCase() || 'VIDEO'
    return `${ext} video`
  }
  if (mime.startsWith('audio/')) return 'Audio file'
  if (mime.startsWith('text/')) return 'Text file'
  if (mime === 'application/zip' || mime === 'application/x-zip-compressed') return 'Archive file'
  if (mime.startsWith('application/')) {
    const ext = mime.split('/')[1]?.toUpperCase()
    return ext ? `${ext} file` : 'File'
  }
  return 'File'
})

const formattedSize = computed(() => {
  if (!props.item || props.item.size === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(props.item.size) / Math.log(k))
  return parseFloat((props.item.size / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
})

function formatDate(dateStr?: string): string {
  if (!dateStr) return '—'
  const date = new Date(dateStr)
  return date.toLocaleDateString('vi-VN', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formattedCreatedAt = computed(() => formatDate(props.item?.created_at))
const formattedUpdatedAt = computed(() => formatDate(props.item?.updated_at))
</script>

<style scoped>
.properties-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.properties-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding-bottom: 0.5rem;
}

.properties-name-block {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.properties-name {
  font-size: var(--font-size-base);
  font-weight: 600;
  color: var(--color-text-primary);
  word-break: break-all;
}

.properties-type {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
}

.properties-section-label {
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding-bottom: 0.25rem;
  border-bottom: 1px solid var(--color-border);
}

.properties-rows {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.properties-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 1rem;
  font-size: var(--font-size-sm);
}

.properties-row-label {
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.properties-row-value {
  color: var(--color-text-primary);
  text-align: right;
  word-break: break-all;
}

.properties-row-value.path-value {
  font-size: var(--font-size-xs);
  font-family: monospace;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/app/components/folders/FileProperties.vue
git commit -m "feat: add FileProperties modal component"
```

---

## Chunk 2: Integrate into index.vue

**Files:**
- Modify: `frontend/app/pages/index.vue`

### Add import

Find the `<script setup lang="ts">` block. Add import:

```ts
import FileProperties from '~/components/folders/FileProperties.vue'
```

### Add state

After `const contextTarget = ref<{ id: string; name: string } | null>(null)`:

```ts
const showPropertiesModal = ref(false)
const propertiesTarget = ref<Item | null>(null)
```

### Add "Thông tin" to context menu

Update `contextMenuOptions`:

```ts
const contextMenuOptions = computed(() => [
  { label: 'Thông tin', key: 'properties' },
  { type: 'divider', key: 'd0' },
  { label: 'Đổi tên', key: 'rename' },
  { type: 'divider', key: 'd1' },
  { label: 'Xóa', key: 'delete' },
])
```

### Handle selection

Update `onContextSelect`:

```ts
function onContextSelect(key: string) {
  showContextMenu.value = false
  if (key === 'properties') {
    const item = displayItems.value.find(i => i.id === contextTarget.value?.id)
    propertiesTarget.value = item || null
    showPropertiesModal.value = true
  }
  if (key === 'rename') showRenameDialog.value = true
  if (key === 'delete') showDeleteDialog.value = true
}
```

### Mount component

Find `</div>` at line ~148 (before `<!-- Restore Folder Dialog -->`) and add:

```vue
    <!-- File Properties Modal -->
    <FileProperties
      v-model:show="showPropertiesModal"
      :item="propertiesTarget"
    />
```

- [ ] **Step 1: Apply all 4 edits above**

- [ ] **Step 2: Commit**

```bash
git add frontend/app/pages/index.vue
git commit -m "feat: add Thông tin context menu option"
```

---

## Chunk 3: Verify

- [ ] **Step 1: Check for lint errors**

```bash
cd frontend && npm run lint
```

- [ ] **Step 2: Verify component exists and is syntactically valid**

```bash
node -e "require('fs').readFileSync('app/components/folders/FileProperties.vue', 'utf8')"
```

- [ ] **Final commit if needed**

```bash
git add -A && git commit -m "chore: finalize file properties feature"
```
