<template>
  <n-spin :show="displayLoading">
    <div v-if="displayItems.length === 0 && !displayLoading" class="fm-empty">
      <n-empty :description="isTrashView ? 'Trash is empty' : 'Folder is empty'">
        <template #icon>
          <n-icon size="48" :depth="3">
            <Icon :icon="isTrashView ? 'mdi:trash-can' : 'mdi:folder-open-outline'" />
          </n-icon>
        </template>
        <template v-if="!isTrashView" #extra>
          <n-button size="small" @click="$emit('create-folder-click')">
            New folder
          </n-button>
        </template>
      </n-empty>
    </div>

    <div v-else class="fm-list-wrapper">
      <table class="fm-list-table">
        <thead>
          <tr>
            <th class="col-name">Name</th>
            <th class="col-size">Size</th>
            <th class="col-modified">Modified</th>
            <th class="col-actions"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in displayItems"
            :key="item.id"
            class="fm-list-row"
            :class="{ 'is-trash': isTrashView }"
            @dblclick="!isTrashView && $emit('action', { type: 'open', item })"
            @contextmenu.prevent="isTrashView
              ? $emit('action', { type: 'trash-context', item, eventX: $event.clientX, eventY: $event.clientY })
              : $emit('action', { type: 'context', item, eventX: $event.clientX, eventY: $event.clientY })"
          >
            <td class="col-name">
              <div class="fm-list-name-cell">
                <span class="fm-list-icon">
                  <n-icon v-if="isTrashView" size="20">
                    <Icon icon="mdi:delete-outline" />
                  </n-icon>
                  <FileIcon
                    v-else
                    :filename="item.name"
                    :isFolder="item.is_folder"
                    :size="20"
                    :style="item.color ? { color: item.color } : {}"
                  />
                </span>
                <span class="fm-list-name">{{ item.name }}</span>
              </div>
            </td>
            <td class="col-size">
              <template v-if="isTrashView">—</template>
              <template v-else-if="item.is_folder">—</template>
              <template v-else>{{ formatSize(item.size) }}</template>
            </td>
            <td class="col-modified">
              <template v-if="isTrashView">
                {{ formatDeletedDate(item.deleted_at) }}
              </template>
              <template v-else>
                {{ formatDate(item.updated_at) }}
              </template>
            </td>
            <td class="col-actions">
              <div class="fm-list-actions">
                <div v-if="isTrashView" class="fm-list-trash-btns">
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
                <button
                  v-else
                  class="fm-list-menu-btn"
                  :aria-label="isTrashView ? 'Item actions' : 'Open menu'"
                  @click.stop="$emit('action', { type: 'menu', item, element: $event.currentTarget as HTMLElement })"
                >
                  <n-icon size="16">
                    <Icon icon="mdi:dots-horizontal" />
                  </n-icon>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </n-spin>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import FileIcon from './FileIcon.vue'
import type { Item } from '~/types/folder'

defineProps<{
  displayItems: Item[]
  displayLoading: boolean
  isTrashView: boolean
}>()

defineEmits<{
  (e: 'action', event: ListActionEvent): void
  (e: 'create-folder-click'): void
}>()

interface ListActionEvent {
  type: 'open' | 'menu' | 'context' | 'trash-context' | 'restore' | 'permanent-delete'
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
  return new Date(dateStr).toLocaleDateString('vi-VN', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('vi-VN', {
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

.fm-list-wrapper {
  overflow-x: auto;
}

.fm-list-table {
  width: 100%;
  border-collapse: collapse;
  font-size: var(--font-size-sm);
}

.fm-list-table thead th {
  text-align: left;
  padding: 0.625rem 1rem;
  font-weight: 600;
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-text-muted);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  background: var(--color-bg-primary);
  z-index: 1;
}

.fm-list-row {
  cursor: pointer;
  transition: background var(--transition-fast);
}

.fm-list-row:hover {
  background: var(--color-surface-hover);
}

.fm-list-row:hover .fm-list-menu-btn,
.fm-list-row:hover .fm-list-trash-btns {
  opacity: 1;
}

.fm-list-row td {
  padding: 0.625rem 1rem;
  border-bottom: 1px solid var(--color-border);
  vertical-align: middle;
  color: var(--color-text-primary);
}

.col-name { min-width: 280px; width: 45%; }
.col-size { width: 12%; color: var(--color-text-secondary); }
.col-modified { width: 20%; color: var(--color-text-secondary); }
.col-actions { width: 80px; text-align: right; }

.fm-list-name-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.fm-list-icon {
  flex-shrink: 0;
  display: flex;
  color: var(--color-text-secondary);
  transition: filter var(--transition-base);
}

.is-folder .fm-list-icon {
  filter: drop-shadow(0 0 4px rgba(59, 130, 246, 0.4));
}

.fm-list-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: calc(100% - 2rem);
  display: block;
}

.fm-list-actions {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.fm-list-menu-btn {
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
  transition: all var(--transition-fast);
}

.fm-list-menu-btn:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
}

.fm-list-trash-btns {
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity var(--transition-fast);
}

@media (hover: none), (pointer: coarse) {
  .fm-list-menu-btn { opacity: 1; }
  .fm-list-trash-btns { opacity: 1; }
}

@media (max-width: 768px) {
  .fm-list-table thead th,
  .fm-list-row td {
    padding: 0.5rem 0.625rem;
  }
  .col-modified { display: none; }
  .col-size { width: 15%; }
}
</style>
