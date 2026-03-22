<template>
  <div class="file-manager">
    <!-- Header -->
    <div class="fm-header">
      <div class="fm-breadcrumb">
        <template v-if="isTrashView">
          <span class="trash-title">Thùng rác</span>
        </template>
        <template v-else>
          <n-breadcrumb>
            <n-breadcrumb-item
              v-for="(crumb, i) in breadcrumbs"
              :key="i"
              :clickable="i < breadcrumbs.length - 1"
              @click="onBreadcrumbClick(i)"
            >
              {{ crumb.name }}
            </n-breadcrumb-item>
          </n-breadcrumb>
        </template>
      </div>
      <div class="fm-actions">
        <template v-if="isTrashView">
          <n-button
            v-if="trashItems.length > 0"
            type="warning"
            size="small"
            @click="handleEmptyTrash"
          >
            <template #icon>
              <n-icon><Icon icon="mdi:delete-sweep" /></n-icon>
            </template>
            Dọn thùng rác
          </n-button>
        </template>
        <template v-else>
          <n-button type="primary" size="small" @click="showUploadFile = true">
            <template #icon>
              <n-icon><Icon icon="mdi:upload" /></n-icon>
            </template>
            Tải lên
          </n-button>
          <n-button type="primary" size="small" @click="showCreateFolder = true">
            <template #icon>
              <n-icon><Icon icon="mdi:folder-plus" /></n-icon>
            </template>
            Tạo thư mục
          </n-button>
        </template>
      </div>
    </div>

    <!-- Items Grid/List -->
    <n-spin :show="displayLoading">
      <div v-if="displayItems.length === 0 && !displayLoading" class="fm-empty">
        <n-empty :description="isTrashView ? 'Thùng rác trống' : 'Thư mục trống'">
          <template #icon>
            <n-icon size="48" :depth="3">
              <Icon :icon="isTrashView ? 'mdi:delete-outline' : 'mdi:folder-open-outline'" />
            </n-icon>
          </template>
          <template v-if="!isTrashView" #extra>
            <n-button size="small" @click="showCreateFolder = true">
              Tạo thư mục mới
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
          @dblclick="!isTrashView && onItemDblClick(item)"
          @contextmenu.prevent="isTrashView ? onTrashContext($event, item) : onItemContext($event, item)"
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
              <span v-if="item.is_folder">{{ item.child_count }} mục</span>
              <span v-else>{{ formatSize(item.size) }}</span>
            </template>
          </div>
          <div v-if="isTrashView" class="fm-trash-actions">
            <n-button size="tiny" @click.stop="handleRestore(item)">
              <template #icon>
                <n-icon><Icon icon="mdi:restore" /></n-icon>
              </template>
              Khôi phục
            </n-button>
            <n-button size="tiny" type="error" @click.stop="handlePermanentDelete(item)">
              <template #icon>
                <n-icon><Icon icon="mdi:delete-forever" /></n-icon>
              </template>
            </n-button>
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

    <!-- Context menu for trash items -->
    <n-dropdown
      :show="showTrashContextMenu"
      :x="contextX"
      :y="contextY"
      trigger="manual"
      placement="bottom-start"
      :options="trashContextOptions"
      @select="onTrashContextSelect"
      @clickoutside="showTrashContextMenu = false"
    />

    <!-- Modals -->
    <FolderActions
      v-model:show-create="showCreateFolder"
      v-model:show-rename="showRenameDialog"
      v-model:show-delete="showDeleteDialog"
      v-model:show-upload="showUploadFile"
      :target-item="contextTarget"
      :parent-id="currentFolderId || undefined"
      @create="handleCreate"
      @rename="handleRename"
      @delete="handleDelete"
      @upload="handleUpload"
    />

    <!-- Restore Folder Dialog -->
    <n-modal
      v-model:show="showRestoreDialog"
      preset="card"
      title="Chọn thư mục khôi phục"
      style="width: 400px"
    >
      <div class="restore-folder-list">
        <div
          class="restore-folder-item"
          :class="{ active: selectedRestoreFolder === null }"
          @click="selectedRestoreFolder = null"
        >
          <n-icon><Icon icon="mdi:home-outline" /></n-icon>
          <span>Thư mục gốc</span>
        </div>
        <div
          v-for="folder in folderTree"
          :key="folder.id"
          class="restore-folder-item"
          :class="{ active: selectedRestoreFolder === folder.id }"
          :style="{ paddingLeft: (folder.depth * 20 + 24) + 'px' }"
          @click="selectedRestoreFolder = folder.id"
        >
          <n-icon><Icon icon="mdi:folder" /></n-icon>
          <span>{{ folder.name }}</span>
        </div>
      </div>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showRestoreDialog = false">Hủy</n-button>
          <n-button type="primary" @click="handleRestoreWithFolder(selectedRestoreFolder)">
            Khôi phục
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Rename Dialog -->
    <n-modal
      v-model:show="showRenameDialog"
      preset="card"
      title="Đổi tên"
      style="width: 400px"
    >
      <n-input v-model:value="newItemName" placeholder="Tên mới" />
      <template #footer>
        <n-space justify="end">
          <n-button @click="showRenameDialog = false">Hủy</n-button>
          <n-button type="primary" @click="handleRestoreWithRename">
            Khôi phục
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { storeToRefs } from 'pinia'
import type { Item, RestoreItemRequest, FolderTreeNode } from '~/types/folder'
import { useFolderStore } from '~/stores/folder'

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
  uploadFile,
  loadFolderTree,
  items,
  currentFolderId,
  breadcrumbs,
  loading,
  folderTree,
} = useFolder()

const folderStore = useFolderStore()
const { isTrashView, trashItems, trashLoading } = storeToRefs(folderStore)

const api = useApi()
const message = useMessage()
const dialog = useDialog()

const displayItems = computed(() => {
  return isTrashView.value ? trashItems.value : items.value
})

const displayLoading = computed(() => {
  return isTrashView.value ? trashLoading.value : loading.value
})

const showCreateFolder = ref(false)
const showRenameDialog = ref(false)
const showDeleteDialog = ref(false)
const showUploadFile = ref(false)
const showContextMenu = ref(false)
const contextX = ref(0)
const contextY = ref(0)
const contextTarget = ref<{ id: string; name: string } | null>(null)

// Trash-specific state
const showTrashContextMenu = ref(false)
const trashContextTarget = ref<Item | null>(null)
const showRestoreDialog = ref(false)
const selectedRestoreFolder = ref<string | null>(null)
const newItemName = ref('')

const contextMenuOptions = computed(() => [
  { label: 'Đổi tên', key: 'rename' },
  { type: 'divider', key: 'd1' },
  { label: 'Xóa', key: 'delete' },
])

const trashContextOptions = computed(() => [
  { label: 'Khôi phục', key: 'restore' },
  { type: 'divider', key: 'd1' },
  { label: 'Xóa vĩnh viễn', key: 'permanent-delete' },
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

function formatDeletedDate(dateStr?: string): string {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  return date.toLocaleDateString('vi-VN', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
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

function onTrashContext(e: MouseEvent, item: Item) {
  contextX.value = e.clientX
  contextY.value = e.clientY
  trashContextTarget.value = item
  showTrashContextMenu.value = true
}

function onTrashContextSelect(key: string) {
  showTrashContextMenu.value = false
  if (!trashContextTarget.value) return
  if (key === 'restore') handleRestore(trashContextTarget.value)
  if (key === 'permanent-delete') handlePermanentDelete(trashContextTarget.value)
}

function onBreadcrumbClick(index: number) {
  const crumb = breadcrumbs.value[index]
  if (!crumb) return
  if (index === breadcrumbs.value.length - 1) return
  navigateToFolder(crumb.id, crumb.path)
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

async function handleUpload(file: File, parentId?: string, onProgress?: (progress: number) => void) {
  try {
    await uploadFile(file, parentId || undefined, onProgress)
    message.success('Đã tải lên file')
    showUploadFile.value = false
  } catch (e: any) {
    console.error('Upload error:', e)
    message.error(e?.data?.error || 'Không thể tải lên file')
  }
}

// Trash handlers
async function loadTrash() {
  folderStore.setTrashLoading(true)
  try {
    const data = await api.getTrash()
    folderStore.setTrashItems(data)
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể tải thùng rác')
  } finally {
    folderStore.setTrashLoading(false)
  }
}

async function handleRestore(item: Item) {
  try {
    await api.restoreItem(item.id)
    message.success('Đã khôi phục')
    folderStore.removeTrashItem(item.id)
  } catch (e: any) {
    const code = e?.data?.code
    if (code === 'PARENT_DELETED') {
      restoreTarget.value = { id: item.id, name: item.name }
      showRestoreDialog.value = true
    } else if (code === 'NAME_CONFLICT') {
      restoreTarget.value = { id: item.id, name: item.name }
      newItemName.value = item.name
      showRenameDialog.value = true
    } else {
      message.error(e?.data?.error || 'Không thể khôi phục')
    }
  }
}

const restoreTarget = ref<{ id: string; name: string } | null>(null)

async function handleRestoreWithFolder(folderId: string | null) {
  if (!restoreTarget.value) return
  try {
    const body: RestoreItemRequest = { targetParentID: folderId || undefined }
    await api.restoreItem(restoreTarget.value.id, body)
    message.success('Đã khôi phục')
    folderStore.removeTrashItem(restoreTarget.value.id)
    showRestoreDialog.value = false
    restoreTarget.value = null
  } catch (e: any) {
    if (e?.data?.code === 'NAME_CONFLICT' && restoreTarget.value) {
      newItemName.value = restoreTarget.value.name
      showRenameDialog.value = true
      showRestoreDialog.value = false
    } else {
      message.error(e?.data?.error || 'Không thể khôi phục')
    }
  }
}

async function handleRestoreWithRename() {
  if (!restoreTarget.value || !newItemName.value.trim()) return
  try {
    const body: RestoreItemRequest = { newName: newItemName.value.trim() }
    await api.restoreItem(restoreTarget.value.id, body)
    message.success('Đã khôi phục')
    folderStore.removeTrashItem(restoreTarget.value.id)
    showRenameDialog.value = false
    restoreTarget.value = null
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể khôi phục')
  }
}

async function handlePermanentDelete(item: Item) {
  dialog.warning({
    title: 'Xóa vĩnh viễn',
    content: `Bạn có chắc muốn xóa vĩnh viễn "${item.name}"? Hành động này không thể hoàn tác.`,
    positiveText: 'Xóa vĩnh viễn',
    negativeText: 'Hủy',
    onPositiveClick: async () => {
      try {
        await api.permanentDeleteItem(item.id)
        message.success('Đã xóa vĩnh viễn')
        folderStore.removeTrashItem(item.id)
      } catch (e: any) {
        message.error(e?.data?.error || 'Không thể xóa vĩnh viễn')
      }
    },
  })
}

async function handleEmptyTrash() {
  dialog.warning({
    title: 'Dọn thùng rác',
    content: `Bạn có chắc muốn xóa vĩnh viễn tất cả ${trashItems.value.length} item trong thùng rác?`,
    positiveText: 'Xóa tất cả',
    negativeText: 'Hủy',
    onPositiveClick: async () => {
      let successCount = 0
      let failCount = 0
      const failedIds: string[] = []
      
      for (const item of trashItems.value) {
        try {
          await api.permanentDeleteItem(item.id)
          successCount++
        } catch (e) {
          failCount++
          failedIds.push(item.name)
        }
      }
      
      // Reload trash list to get accurate state
      await loadTrash()
      
      if (failCount === 0) {
        message.success('Đã dọn thùng rác')
      } else if (successCount === 0) {
        message.error(`Không thể xóa ${failCount} item`)
      } else {
        message.warning(`Đã xóa ${successCount} item, ${failCount} item thất bại`)
      }
    },
  })
}

// Watch trash view
watch(isTrashView, async (isTrash) => {
  if (isTrash) {
    await loadTrash()
  }
}, { immediate: true })
</script>

<style scoped>
.file-manager {
  width: 100%;
  min-width: 0;
}

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

.trash-title {
  font-size: var(--font-size-lg);
  font-weight: 600;
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

.restore-folder-list {
  max-height: 300px;
  overflow-y: auto;
}

.restore-folder-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  cursor: pointer;
  border-radius: 6px;
  transition: all var(--transition-base);
}

.restore-folder-item:hover {
  background: var(--color-surface-hover);
}

.restore-folder-item.active {
  background: rgba(64, 158, 255, 0.15);
  color: var(--color-primary);
}

/* Mobile Responsive */
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

  .fm-actions :deep(.n-button) {
    font-size: var(--font-size-xs);
    padding: 0 0.75rem;
  }
}
</style>
