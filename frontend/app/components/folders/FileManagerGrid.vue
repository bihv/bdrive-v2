<template>
  <n-spin :show="displayLoading">
    <div v-if="displayItems.length === 0 && !displayLoading" class="fm-empty">
      <n-empty :description="isTrashView ? 'Trash is empty' : 'Folder is empty'">
        <template #icon>
          <n-icon size="48" :depth="3">
            <Icon :icon="isTrashView ? 'mdi:delete-outline' : 'mdi:folder-open-outline'" />
          </n-icon>
        </template>
        <template v-if="!isTrashView" #extra>
          <n-button size="small" @click="$emit('create-folder-click')">
            New folder
          </n-button>
        </template>
      </n-empty>
    </div>

    <div v-else class="fm-grid">
      <div
        v-for="item in displayItems"
        :key="item.id"
        class="fm-item glass-card"
        :class="{ 'is-folder': item.is_folder, 'is-trash': isTrashView }"
        @dblclick="!isTrashView && $emit('item-dblclick', item)"
        @contextmenu.prevent="isTrashView ? $emit('trash-context', $event, item) : $emit('item-context', $event, item)"
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
            <span v-if="item.is_folder">{{ item.child_count }} items</span>
            <span v-else>{{ formatSize(item.size) }}</span>
          </template>
        </div>
        <div v-if="isTrashView" class="fm-trash-actions">
          <n-button size="tiny" @click.stop="$emit('restore-item', item)">
            <template #icon>
              <n-icon><Icon icon="mdi:restore" /></n-icon>
            </template>
            Restore
          </n-button>
          <n-button size="tiny" type="error" @click.stop="$emit('permanent-delete-item', item)">
            <template #icon>
              <n-icon><Icon icon="mdi:delete-forever" /></n-icon>
            </template>
          </n-button>
        </div>
      </div>
    </div>
  </n-spin>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { Item } from '~/types/folder'

defineProps<{
  displayItems: Item[]
  displayLoading: boolean
  isTrashView: boolean
}>()

defineEmits<{
  (e: 'create-folder-click'): void
  (e: 'item-dblclick', item: Item): void
  (e: 'trash-context', event: MouseEvent, item: Item): void
  (e: 'item-context', event: MouseEvent, item: Item): void
  (e: 'restore-item', item: Item): void
  (e: 'permanent-delete-item', item: Item): void
}>()

function getItemIcon(item: Item): string {
  if (item.is_folder) return 'mdi:folder'
  if (item.mime_type?.startsWith('image/')) return 'mdi:file-image'
  if (item.mime_type?.startsWith('video/')) return 'mdi:file-video'
  if (item.mime_type === 'application/pdf') return 'mdi:file-pdf-box'
  return 'mdi:file-document-outline'
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatDeletedDate(dateStr?: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('vi-VN', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}
</script>

<style scoped>
.fm-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 400px;
}

.fm-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 1rem;
}

.fm-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.25rem 0.75rem;
  cursor: pointer;
  transition: all var(--transition-base);
  text-align: center;
}

.fm-item:hover {
  background: var(--color-surface-hover);
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.fm-item-icon {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
}

.fm-item.is-folder .fm-item-icon {
  color: var(--color-primary);
}

.fm-item-name {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.fm-item-meta {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  margin-top: 0.25rem;
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

/* Mobile Responsive */
@media (max-width: 768px) {
  .fm-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 0.75rem;
  }

  .fm-item {
    padding: 1rem 0.5rem;
  }

  .fm-item-icon {
    font-size: 0.8em;
  }

  .fm-item-icon :deep(.n-icon) {
    font-size: 32px !important;
  }

  .fm-item-name {
    font-size: var(--font-size-xs);
  }

  .fm-empty {
    min-height: 300px;
  }
}

@media (max-width: 480px) {
  .fm-grid {
    grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
    gap: 0.5rem;
  }

  .fm-item {
    padding: 0.75rem 0.25rem;
  }

  .fm-item-icon :deep(.n-icon) {
    font-size: 28px !important;
  }

  .fm-item-name {
    font-size: 0.7rem;
  }

  .fm-item-meta {
    font-size: 0.65rem;
  }
}
</style>
