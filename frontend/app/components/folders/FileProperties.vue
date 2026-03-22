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
      <!-- Loading -->
      <div v-if="loading" class="properties-loading">
        <n-spin size="small" />
      </div>

      <template v-else-if="item">
        <!-- Header: Icon + Name -->
        <div class="properties-header">
          <n-icon :size="48" :style="item.color ? { color: item.color } : {}">
            <Icon :icon="getItemIcon(item)" />
          </n-icon>
          <div class="properties-name-block">
            <div class="properties-name">{{ item.name }}</div>
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

          <div v-if="!item.is_folder" class="properties-row">
            <span class="properties-row-label">Kích thước</span>
            <span class="properties-row-value">{{ formattedSize }}</span>
          </div>

          <div class="properties-row">
            <span class="properties-row-label">Vị trí</span>
            <span class="properties-row-value path-value">{{ item.path || '—' }}</span>
          </div>

          <div class="properties-row">
            <span class="properties-row-label">Ngày tạo</span>
            <span class="properties-row-value">{{ formattedCreatedAt }}</span>
          </div>

          <div class="properties-row">
            <span class="properties-row-label">Ngày sửa</span>
            <span class="properties-row-value">{{ formattedUpdatedAt }}</span>
          </div>

          <div v-if="item.is_folder" class="properties-row">
            <span class="properties-row-label">Số mục con</span>
            <span class="properties-row-value">{{ item.child_count }} mục</span>
          </div>
        </div>
      </template>

      <div v-else class="properties-empty">
        <n-empty description="Không tìm thấy thông tin" size="small" />
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { Item } from '~/types/folder'

const props = defineProps<{
  show: boolean
  itemId: string | null
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
}>()

const api = useApi()
const item = ref<Item | null>(null)
const loading = ref(false)

const modalTitle = computed(() => {
  return props.itemId ? 'Thông tin' : 'Thông tin'
})

watch(
  () => props.show,
  async (isOpen) => {
    if (!isOpen || !props.itemId) {
      item.value = null
      return
    }
    loading.value = true
    try {
      item.value = await api.getItem(props.itemId)
    } catch {
      item.value = null
    } finally {
      loading.value = false
    }
  },
)

watch(
  () => props.itemId,
  async (id) => {
    if (!id || !props.show) return
    loading.value = true
    try {
      item.value = await api.getItem(id)
    } catch {
      item.value = null
    } finally {
      loading.value = false
    }
  },
)

function getItemIcon(i: Item): string {
  if (i.is_folder) return 'mdi:folder'
  if (i.mime_type?.startsWith('image/')) return 'mdi:file-image'
  if (i.mime_type?.startsWith('video/')) return 'mdi:file-video'
  if (i.mime_type === 'application/pdf') return 'mdi:file-pdf-box'
  return 'mdi:file-document-outline'
}

const fileTypeLabel = computed(() => {
  if (!item.value) return ''
  if (item.value.is_folder) return 'Thư mục'

  const mime = item.value.mime_type || ''

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
  if (!item.value || item.value.size === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(item.value.size) / Math.log(k))
  return parseFloat((item.value.size / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
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

const formattedCreatedAt = computed(() => formatDate(item.value?.created_at))
const formattedUpdatedAt = computed(() => formatDate(item.value?.updated_at))
</script>

<style scoped>
.properties-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.properties-loading {
  display: flex;
  justify-content: center;
  padding: 2rem;
}

.properties-empty {
  padding: 1rem 0;
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
