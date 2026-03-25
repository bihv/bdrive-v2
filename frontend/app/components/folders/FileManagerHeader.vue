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
      <template v-else-if="activeView === 'all' && viewTitle !== 'All files'">
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
      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button
            quaternary
            size="small"
            class="fm-search-btn"
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
          class="fm-action-btn fm-action-btn--danger"
          @click="$emit('empty-trash')"
        >
          <template #icon>
            <n-icon><Icon icon="mdi:delete-sweep" /></n-icon>
          </template>
          <span class="fm-action-label">Empty Trash</span>
        </n-button>
      </template>
      <template v-else>
        <n-dropdown
          trigger="click"
          :options="createActionOptions"
          @select="handleCreateActionSelect"
        >
          <n-button
            type="primary"
            size="small"
            class="fm-action-btn fm-action-btn--create"
          >
            <template #icon>
              <n-icon><Icon icon="mdi:plus" /></n-icon>
            </template>
            <span class="fm-action-label">New</span>
          </n-button>
        </n-dropdown>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { h } from 'vue'
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

const createActionOptions = [
  {
    label: 'Upload file',
    key: 'upload',
    icon: () => h(Icon, { icon: 'mdi:upload' }),
  },
  {
    label: 'New folder',
    key: 'folder',
    icon: () => h(Icon, { icon: 'mdi:folder-plus' }),
  },
]

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

function handleCreateActionSelect(key: string) {
  if (key === 'upload') {
    emit('upload-click')
    return
  }

  if (key === 'folder') {
    emit('create-folder-click')
  }
}
</script>

<style scoped>
.fm-header {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  grid-template-areas: "title tabs actions";
  align-items: center;
  margin-bottom: 1rem;
  gap: 1rem;
  padding: 1rem 1.1rem;
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-xl);
  background:
    radial-gradient(circle at left top, rgba(59, 130, 246, 0.12), transparent 28%),
    rgba(255, 255, 255, 0.03);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.05);
}

.fm-tabs {
  grid-area: tabs;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.25rem;
  padding: 0.25rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: var(--color-bg-secondary);
  width: min(100%, 540px);
  backdrop-filter: blur(18px);
  justify-self: center;
}

.fm-breadcrumb {
  grid-area: title;
  min-width: 0;
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
  grid-area: actions;
  display: flex;
  gap: 0.5rem;
  align-items: center;
  justify-content: flex-end;
  flex-wrap: nowrap;
  justify-self: end;
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
  font-size: clamp(1.15rem, 2vw, 1.6rem);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.03em;
}

@media (max-width: 768px) {
  .fm-header {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    grid-template-areas:
      "title actions"
      "tabs tabs";
    align-items: center;
    gap: 0.75rem;
    padding: 0.85rem;
    margin-bottom: 0;
  }

  .fm-breadcrumb {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    min-width: 0;
  }

  .fm-actions {
    justify-content: space-between;
    gap: 0.35rem;
  }

  .fm-search-btn {
    display: none;
  }

  .fm-tabs {
    grid-area: tabs;
    width: 100%;
    overflow-x: auto;
    grid-template-columns: repeat(3, minmax(120px, 1fr));
  }

  .fm-tab {
    min-height: 40px;
    padding: 0.5rem 0.65rem;
  }

  .fm-actions :deep(.n-button) {
    flex-shrink: 0;
  }

  .fm-view-switcher {
    flex-shrink: 0;
  }

  .fm-view-switcher :deep(.n-button) {
    min-width: 34px;
  }

  .crumb-label {
    max-width: 140px;
  }

  .fm-breadcrumb :deep(.n-breadcrumb) {
    width: 100%;
  }
}

@media (max-width: 480px) {
  .fm-tabs {
    display: flex;
    width: 100%;
    padding: 0.2rem;
    gap: 0.2rem;
  }

  .fm-tab {
    flex: 1 0 104px;
    font-size: var(--font-size-sm);
  }

  .fm-actions :deep(.n-button) {
    font-size: var(--font-size-xs);
    padding: 0 0.6rem;
  }

  .fm-actions {
    gap: 0.25rem;
  }

  .fm-actions > :nth-child(1) :deep(.n-button),
  .fm-actions > :nth-child(2) :deep(.n-button) {
    width: 36px;
    padding: 0;
  }

  .fm-action-btn--create :deep(.n-button__content),
  .fm-action-btn--danger :deep(.n-button__content) {
    gap: 0;
  }

  .fm-action-btn--create :deep(.n-button__content .fm-action-label),
  .fm-action-btn--danger :deep(.n-button__content .fm-action-label) {
    display: none;
  }

  .fm-action-btn--create,
  .fm-action-btn--danger {
    width: 36px;
  }

  .fm-action-btn--create :deep(.n-button),
  .fm-action-btn--danger :deep(.n-button) {
    width: 36px;
    min-width: 36px;
    padding: 0;
  }
}
</style>
