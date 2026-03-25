<template>
  <n-spin :show="displayLoading">
    <div v-if="displayItems.length === 0 && !displayLoading" class="fm-empty">
      <n-empty :description="isTrashView ? 'Trash is empty' : 'Folder is empty'">
        <template #icon>
          <n-icon size="48" :depth="3">
            <Icon :icon="isTrashView ? 'mdi:trash-can' : 'mdi:folder-open-outline'" />
          </n-icon>
        </template>
      </n-empty>
    </div>

    <div v-else class="fm-grid">
      <div
        v-for="item in displayItems"
        :key="item.id"
        :data-item-id="item.id"
        class="fm-item"
        :class="{
          'is-folder': item.is_folder,
          'is-trash': isTrashView,
          'highlighted': item.id === highlightedId
        }"
        @dblclick="!isTrashView && $emit('action', { type: 'open', item })"
        @contextmenu.prevent="isTrashView
          ? $emit('action', { type: 'trash-context', item, eventX: $event.clientX, eventY: $event.clientY })
          : $emit('action', { type: 'context', item, eventX: $event.clientX, eventY: $event.clientY })"
      >
        <div class="fm-item-icon">
          <n-icon v-if="isTrashView" size="36">
            <Icon icon="mdi:delete-outline" />
          </n-icon>
          <FileIcon
            v-else
            :filename="item.name"
            :isFolder="item.is_folder"
            :size="36"
            :style="item.color ? { color: item.color } : {}"
          />
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
        <div v-if="!isTrashView" class="fm-item-quick-actions">
          <button
            class="fm-quick-action-btn"
            :class="{ active: isStarred(item.id) }"
            :aria-label="isStarred(item.id) ? 'Unstar item' : 'Star item'"
            @click.stop="$emit('action', { type: 'toggle-star', item })"
          >
            <n-icon size="16">
              <Icon :icon="isStarred(item.id) ? 'mdi:star' : 'mdi:star-outline'" />
            </n-icon>
          </button>
        </div>
        <button
          class="fm-item-menu-btn"
          :aria-label="isTrashView ? 'Item actions' : 'Open menu'"
          @click.stop="$emit('action', { type: 'menu', item, element: $event.currentTarget as HTMLElement })"
        >
          <n-icon size="16">
            <Icon icon="mdi:dots-horizontal" />
          </n-icon>
        </button>
        <div v-if="isTrashView" class="fm-trash-actions">
          <n-button size="tiny" @click.stop="$emit('action', { type: 'restore', item })">
            <template #icon>
              <n-icon><Icon icon="mdi:restore" /></n-icon>
            </template>
            Restore
          </n-button>
          <n-button size="tiny" type="error" @click.stop="$emit('action', { type: 'permanent-delete', item })">
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
import FileIcon from './FileIcon.vue'
import type { Item } from '~/types/folder'
import { useFolderStore } from '~/stores/folder'
import { useQuickAccessStore } from '~/stores/quick-access'

defineProps<{
  displayItems: Item[]
  displayLoading: boolean
  isTrashView: boolean
  showCreateFolderAction?: boolean
}>()

defineEmits<{
  (e: 'action', event: GridActionEvent): void
}>()

const folderStore = useFolderStore()
const quickAccessStore = useQuickAccessStore()
const highlightedId = computed(() => folderStore.highlightedId)
const isStarred = (itemId: string) => quickAccessStore.isStarred(itemId)

// Auto-scroll to highlighted item when it changes and the item is in the DOM
watch(highlightedId, async (id) => {
  if (!id) return

  // Retry until item appears in DOM (max ~3s)
  const start = Date.now()
  const tryScroll = async () => {
    await nextTick()
    const el = document.querySelector(`[data-item-id="${id}"]`)
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
      return
    }
    if (Date.now() - start < 3000) {
      setTimeout(tryScroll, 100)
    }
  }
  tryScroll()
})

interface GridActionEvent {
  type: 'open' | 'menu' | 'context' | 'trash-context' | 'restore' | 'permanent-delete' | 'toggle-star'
  item: Item
  element?: HTMLElement
  eventX?: number
  eventY?: number
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
  border-radius: var(--radius-xl);
  background: rgba(255, 255, 255, 0.025);
  border: 1px dashed rgba(255, 255, 255, 0.1);
}

.fm-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 1rem;
  padding: 0.25rem;
}

.fm-item {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.25rem 0.75rem;
  cursor: pointer;
  text-align: center;
  border-radius: var(--radius-md);
  border: 1px solid rgba(255, 255, 255, 0.03);
  background: rgba(255, 255, 255, 0.018);
  transition: all var(--transition-base);
}

.fm-item:hover {
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.08), transparent 40%),
    var(--color-surface-hover);
  border-color: rgba(255, 255, 255, 0.08);
  transform: translateY(-1px);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.22);
}

.fm-item:active {
  transform: translateY(0);
}

.fm-item.is-folder .fm-item-icon {
  filter: drop-shadow(0 0 6px rgba(59, 130, 246, 0.5));
}

.fm-item.highlighted {
  border-color: rgba(59, 130, 246, 0.35);
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.16), 0 10px 22px rgba(8, 47, 73, 0.14);
}

.fm-item-icon {
  margin-bottom: 0.5rem;
  color: var(--color-text-secondary);
  transition: filter var(--transition-base), transform var(--transition-base);
}

.fm-item:hover .fm-item-icon {
  transform: scale(1.05);
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

.fm-item-quick-actions {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity var(--transition-base);
}

.fm-item:hover .fm-item-quick-actions {
  opacity: 1;
}

.fm-item-quick-actions:has(.fm-quick-action-btn.active) {
  opacity: 1;
}

.fm-quick-action-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.78);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.fm-quick-action-btn.active {
  color: #fbbf24;
}

.fm-quick-action-btn:hover {
  color: var(--color-text-primary);
  background: rgba(30, 41, 59, 0.92);
}

.fm-item-menu-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0;
  transition: all var(--transition-base);
}

.fm-item:hover .fm-item-menu-btn {
  opacity: 1;
}

.fm-item-menu-btn:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
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

/* Mobile / touch: always show 3-dot button */
@media (hover: none), (pointer: coarse) {
  .fm-item-menu-btn {
    opacity: 1;
  }
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .fm-grid {
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 0.75rem;
    padding: 0.25rem;
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
