<template>
  <div class="file-manager">
    <!-- Header -->
    <div class="fm-header">
      <div class="fm-breadcrumb">
        <n-breadcrumb>
          <n-breadcrumb-item
            v-for="(crumb, i) in breadcrumbs"
            :key="i"
            @click="onBreadcrumbClick(i)"
          >
            {{ crumb }}
          </n-breadcrumb-item>
        </n-breadcrumb>
      </div>
      <div class="fm-actions">
        <n-button
          type="primary"
          size="small"
          @click="showCreateFolder = true"
        >
          <template #icon>
            <n-icon><Icon icon="mdi:folder-plus" /></n-icon>
          </template>
          Tạo thư mục
        </n-button>
      </div>
    </div>

    <!-- Items Grid/List -->
    <n-spin :show="loading">
      <div v-if="items.length === 0 && !loading" class="fm-empty">
        <n-empty description="Thư mục trống">
          <template #icon>
            <n-icon size="48" :depth="3">
              <Icon icon="mdi:folder-open-outline" />
            </n-icon>
          </template>
          <template #extra>
            <n-button size="small" @click="showCreateFolder = true">
              Tạo thư mục mới
            </n-button>
          </template>
        </n-empty>
      </div>

      <div v-else class="fm-grid">
        <div
          v-for="item in items"
          :key="item.id"
          class="fm-item glass-card"
          :class="{ 'is-folder': item.is_folder }"
          @dblclick="onItemDblClick(item)"
          @contextmenu.prevent="onItemContext($event, item)"
        >
          <div class="fm-item-icon">
            <n-icon
              size="36"
              :style="item.color ? { color: item.color } : {}"
            >
              <Icon :icon="getItemIcon(item)" />
            </n-icon>
          </div>
          <div class="fm-item-name">{{ item.name }}</div>
          <div class="fm-item-meta">
            <span v-if="item.is_folder">{{ item.child_count }} mục</span>
            <span v-else>{{ formatSize(item.size) }}</span>
          </div>
        </div>
      </div>
    </n-spin>

    <!-- Context menu for items -->
    <n-dropdown
      :show="showContextMenu"
      :x="contextX"
      :y="contextY"
      trigger="manual"
      placement="bottom-start"
      :options="contextMenuOptions"
      @select="onContextSelect"
      @clickoutside="showContextMenu = false"
    />

    <!-- Modals -->
    <FolderActions
      v-model:show-create="showCreateFolder"
      v-model:show-rename="showRenameDialog"
      v-model:show-delete="showDeleteDialog"
      :target-item="contextTarget"
      :parent-id="currentFolderId || undefined"
      @create="handleCreate"
      @rename="handleRename"
      @delete="handleDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { Item } from '~/types/folder'

definePageMeta({
  layout: 'default',
  middleware: 'auth',
})

useHead({
  title: '1Drive - File Manager',
})

const {
  loadItems,
  createFolder,
  updateItem,
  deleteItem,
  navigateToFolder,
  items,
  currentFolderId,
  breadcrumbs,
  loading,
} = useFolder()

const message = useMessage()

const showCreateFolder = ref(false)
const showRenameDialog = ref(false)
const showDeleteDialog = ref(false)
const showContextMenu = ref(false)
const contextX = ref(0)
const contextY = ref(0)
const contextTarget = ref<{ id: string; name: string } | null>(null)

const contextMenuOptions = computed(() => [
  { label: 'Đổi tên', key: 'rename' },
  { type: 'divider', key: 'd1' },
  { label: 'Xóa', key: 'delete' },
])

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

function onItemDblClick(item: Item) {
  if (item.is_folder) {
    navigateToFolder(item.id, item.path)
  }
}

function onItemContext(e: MouseEvent, item: Item) {
  contextX.value = e.clientX
  contextY.value = e.clientY
  contextTarget.value = { id: item.id, name: item.name }
  showContextMenu.value = true
}

function onContextSelect(key: string) {
  showContextMenu.value = false
  if (key === 'rename') showRenameDialog.value = true
  if (key === 'delete') showDeleteDialog.value = true
}

function onBreadcrumbClick(index: number) {
  if (index === 0) {
    navigateToFolder(null, '/')
  }
}

async function handleCreate(name: string, parentId?: string) {
  try {
    await createFolder({ name, parent_id: parentId })
    message.success('Đã tạo thư mục')
    showCreateFolder.value = false
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể tạo thư mục')
  }
}

async function handleRename(name: string) {
  if (!contextTarget.value) return
  try {
    await updateItem(contextTarget.value.id, { name })
    message.success('Đã đổi tên')
    showRenameDialog.value = false
    await loadItems(currentFolderId.value)
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể đổi tên')
  }
}

async function handleDelete() {
  if (!contextTarget.value) return
  try {
    await deleteItem(contextTarget.value.id)
    message.success('Đã xóa')
    showDeleteDialog.value = false
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể xóa')
  }
}
</script>

<style scoped>
.file-manager {
  max-width: 1200px;
}

.fm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
  gap: 1rem;
}

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
</style>
