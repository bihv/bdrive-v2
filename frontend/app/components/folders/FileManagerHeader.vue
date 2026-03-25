<template>
  <div class="fm-header">
    <div class="fm-breadcrumb">
      <template v-if="isTrashView">
        <span class="trash-title">{{ viewTitle }}</span>
      </template>
      <template v-else-if="activeView === 'all' && displayBreadcrumbs.length > 0">
        <n-breadcrumb>
          <n-breadcrumb-item
            v-for="(crumb, i) in displayBreadcrumbs"
            :key="i"
            :clickable="crumb.clickable"
            @click="crumb.clickable && $emit('breadcrumb-click', crumb.originalIndex)"
          >
            <n-dropdown
              v-if="crumb.isEllipsis"
              trigger="click"
              :options="collapsedBreadcrumbOptions"
              @select="handleCollapsedBreadcrumbSelect"
            >
              <button
                class="crumb-ellipsis-btn"
                type="button"
                aria-label="Show collapsed folders"
              >
                ...
              </button>
            </n-dropdown>
            <span
              v-else
              class="crumb-label"
              :title="crumb.name"
            >
              {{ crumb.name }}
            </span>
          </n-breadcrumb-item>
        </n-breadcrumb>
      </template>
      <template v-else-if="activeView === 'all'">
        <span class="trash-title">{{ viewTitle }}</span>
      </template>
    </div>
    <div v-if="!isTrashView" class="fm-tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="fm-tab"
        :class="{ active: tab.key === activeView }"
        @click="$emit('view-change', tab.key)"
      >
        <n-icon size="16"><Icon :icon="tab.icon" /></n-icon>
        <span>{{ tab.label }}</span>
      </button>
    </div>
    <div class="fm-actions">
      <!-- Search button — opens SearchPalette -->
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button
            quaternary
            size="small"
            @click="openSearchPalette"
          >
            <template #icon>
              <n-icon><Icon icon="mdi:magnify" /></n-icon>
            </template>
          </n-button>
        </template>
        Search <kbd style="font-size:10px; margin-left: 4px;">⌘K</kbd>
      </n-tooltip>

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
import { useSearchPalette } from '~/composables/useSearchPalette'

const { viewMode, setViewMode } = useFileManagerView()
const { open: openSearchPalette } = useSearchPalette()

const emit = defineEmits<{
  (e: 'breadcrumb-click', index: number): void
  (e: 'empty-trash'): void
  (e: 'upload-click'): void
  (e: 'create-folder-click'): void
  (e: 'view-change', view: 'all' | 'recent' | 'starred'): void
}>()

interface DisplayBreadcrumbItem extends BreadcrumbItem {
  originalIndex: number
  clickable: boolean
  isEllipsis?: boolean
}

const tabs = [
  { key: 'all', label: 'All', icon: 'mdi:folder-multiple-outline' },
  { key: 'recent', label: 'Recent', icon: 'mdi:history' },
  { key: 'starred', label: 'Starred', icon: 'mdi:star-outline' },
] as const

const props = defineProps<{
  isTrashView: boolean
  breadcrumbs: BreadcrumbItem[]
  hasTrashItems: boolean
  activeView: 'all' | 'recent' | 'starred' | 'trash'
  viewTitle: string
}>()

const displayBreadcrumbs = computed<DisplayBreadcrumbItem[]>(() => {
  if (props.activeView !== 'all') return []
  if (props.breadcrumbs.length <= 1) return []

  const mapped = props.breadcrumbs.map((crumb, index) => ({
    ...crumb,
    originalIndex: index,
    clickable: index < props.breadcrumbs.length - 1,
  }))

  if (mapped.length <= 4) {
    return mapped
  }

  return [
    mapped[0],
    {
      id: '__ellipsis__',
      name: '...',
      path: '',
      originalIndex: -1,
      clickable: false,
      isEllipsis: true,
    },
    mapped[mapped.length - 2],
    mapped[mapped.length - 1],
  ]
})

const collapsedBreadcrumbOptions = computed(() => {
  if (props.breadcrumbs.length <= 4) return []

  return props.breadcrumbs
    .slice(1, -2)
    .map((crumb, index) => ({
      label: crumb.name,
      key: String(index + 1),
    }))
})

function handleCollapsedBreadcrumbSelect(key: string) {
  const index = Number(key)
  if (Number.isNaN(index)) return
  emit('breadcrumb-click', index)
}
</script>

<style scoped>
.fm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
  gap: 1rem;
  flex-wrap: wrap;
}

.fm-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.25rem;
  padding: 0.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-secondary);
  width: min(100%, 540px);
}

.crumb-label {
  display: inline-block;
  max-width: min(28vw, 220px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
}

.crumb-ellipsis-btn {
  border: none;
  background: transparent;
  color: inherit;
  cursor: pointer;
  padding: 0;
  font: inherit;
  line-height: 1;
}

.fm-tab {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.5rem 0.8rem;
  border: none;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  outline: none;
  white-space: nowrap;
}

.fm-tab.active {
  background: var(--color-primary);
  color: white;
}

.fm-tab:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
}

.fm-tab.active:hover,
.fm-tab.active:focus-visible,
.fm-tab.active:active {
  background: var(--color-primary);
  color: white;
}

.fm-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.fm-view-switcher {
  display: flex;
  gap: 0.125rem;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 0.125rem;
}

.fm-view-switcher .n-button {
  border-radius: calc(var(--radius-md) - 2px);
  transition: all var(--transition-fast);
}

.fm-view-switcher .n-button:hover {
  background: var(--color-surface-hover);
}

.fm-view-switcher .n-button[type="primary"] {
  background: var(--color-primary) !important;
  box-shadow: 0 0 12px rgba(59, 130, 246, 0.4);
}

.fm-view-switcher .n-button[type="primary"]:hover {
  background: var(--color-primary-hover) !important;
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
    padding: 0.75rem 1rem;
    margin-bottom: 0;
  }

  .fm-breadcrumb {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .fm-actions {
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .fm-tabs {
    width: 100%;
  }

  .crumb-label {
    max-width: 120px;
  }
}

@media (max-width: 480px) {
  .fm-actions :deep(.n-button) {
    font-size: var(--font-size-xs);
    padding: 0 0.75rem;
  }
}
</style>
