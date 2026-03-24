<template>
  <div class="fm-header">
    <div class="fm-breadcrumb">
      <template v-if="isTrashView">
        <span class="trash-title">Trash</span>
      </template>
      <template v-else>
        <n-breadcrumb>
          <n-breadcrumb-item
            v-for="(crumb, i) in breadcrumbs"
            :key="i"
            :clickable="i < breadcrumbs.length - 1"
            @click="$emit('breadcrumb-click', i)"
          >
            {{ crumb.name }}
          </n-breadcrumb-item>
        </n-breadcrumb>
      </template>
    </div>
    <div class="fm-actions">
      <div class="fm-view-switcher">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              quaternary
              size="small"
              :type="viewMode === 'grid' ? 'primary' : 'default'"
              @click="setViewMode('grid')"
            >
              <template #icon>
                <n-icon><Icon icon="mdi:view-grid" /></n-icon>
              </template>
            </n-button>
          </template>
          Grid
        </n-tooltip>
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              quaternary
              size="small"
              :type="viewMode === 'list' ? 'primary' : 'default'"
              @click="setViewMode('list')"
            >
              <template #icon>
                <n-icon><Icon icon="mdi:view-list" /></n-icon>
              </template>
            </n-button>
          </template>
          List
        </n-tooltip>
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button
              quaternary
              size="small"
              :type="viewMode === 'column' ? 'primary' : 'default'"
              @click="setViewMode('column')"
            >
              <template #icon>
                <n-icon><Icon icon="mdi:view-column" /></n-icon>
              </template>
            </n-button>
          </template>
          Columns
        </n-tooltip>
      </div>
      <template v-if="isTrashView">
        <n-button
          v-if="hasTrashItems"
          type="warning"
          size="small"
          @click="$emit('empty-trash')"
        >
          <template #icon>
            <n-icon><Icon icon="mdi:delete-sweep" /></n-icon>
          </template>
          Empty Trash
        </n-button>
      </template>
      <template v-else>
        <n-button type="primary" size="small" @click="$emit('upload-click')">
          <template #icon>
            <n-icon><Icon icon="mdi:upload" /></n-icon>
          </template>
          Upload
        </n-button>
        <n-button type="primary" size="small" @click="$emit('create-folder-click')">
          <template #icon>
            <n-icon><Icon icon="mdi:folder-plus" /></n-icon>
          </template>
          New folder
        </n-button>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { BreadcrumbItem } from '~/types/folder'
import { useFileManagerView } from '~/composables/useFileManagerView'

const { viewMode, setViewMode } = useFileManagerView()

defineProps<{
  isTrashView: boolean
  breadcrumbs: BreadcrumbItem[]
  hasTrashItems: boolean
}>()

defineEmits<{
  (e: 'breadcrumb-click', index: number): void
  (e: 'empty-trash'): void
  (e: 'upload-click'): void
  (e: 'create-folder-click'): void
}>()
</script>

<style scoped>
.fm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
  gap: 1rem;
}

.fm-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.fm-view-switcher {
  display: flex;
  gap: 0.125rem;
  background: var(--glass-bg);
  border: var(--glass-border);
  border-radius: var(--radius-sm);
  padding: 0.125rem;
}

.trash-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
  color: var(--color-text-primary);
}

@media (max-width: 768px) {
  .fm-header {
    flex-direction: column;
    align-items: stretch;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .fm-breadcrumb {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .fm-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .fm-view-switcher {
    display: none;
  }
}

@media (max-width: 480px) {
  .fm-actions :deep(.n-button) {
    font-size: var(--font-size-xs);
    padding: 0 0.75rem;
  }
}
</style>
