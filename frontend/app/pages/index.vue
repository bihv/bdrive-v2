<template>
  <div class="file-manager">
    <!-- Header -->
    <FileManagerHeader
      :is-trash-view="isTrashView"
      :breadcrumbs="breadcrumbs"
      :has-trash-items="trashItems.length > 0"
      :active-view="activeView"
      :view-title="viewTitle"
      @breadcrumb-click="onBreadcrumbClick"
      @view-change="onViewChange"
      @empty-trash="handleEmptyTrash"
      @upload-click="showUploadFile = true"
      @create-folder-click="showCreateFolder = true"
    />

    <!-- Items View (Grid/List/Column) -->
    <FileManagerView
      :items="displayItems"
      :display-loading="displayLoading"
      :is-trash-view="isTrashView"
      :show-create-folder-action="activeView === 'all'"
      :actions="actions"
      @create-folder-click="showCreateFolder = true"
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

    <!-- File Properties Modal -->
    <FileProperties
      v-model:show="showPropertiesModal"
      :item-id="propertiesItemId"
    />

    <!-- Trash Actions Modals -->
    <TrashActions
      v-model:show-restore-dialog="showRestoreDialog"
      v-model:show-rename-dialog="showRenameDialog"
      v-model:selected-restore-folder="selectedRestoreFolder"
      v-model:new-item-name="newItemName"
      :folder-tree="folderTree"
      @restore-with-folder="handleRestoreWithFolder(selectedRestoreFolder)"
      @restore-with-rename="handleRestoreWithRename"
    />
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import type { Item, RestoreItemRequest } from '~/types/folder'
import { useFolderStore } from '~/stores/folder'
import FileProperties from '~/components/folders/FileProperties.vue'
import FileManagerHeader from '~/components/folders/FileManagerHeader.vue'
import FileManagerView from '~/components/folders/FileManagerView.vue'
import TrashActions from '~/components/folders/TrashActions.vue'
import { useItemActions } from '~/composables/useItemActions'

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
const {
  refresh: refreshQuickAccess,
  track,
  toggleStar,
  removeItem: removeQuickAccessItem,
  recentItems,
  starredItems,
  loading: quickAccessLoading,
  isStarred,
} = useQuickAccess()

const api = useApi()
const message = useMessage()
const dialog = useDialog()
const route = useRoute()
const router = useRouter()

const activeView = computed<'all' | 'recent' | 'starred' | 'trash'>(() => {
  const view = route.query.view
  if (view === 'recent' || view === 'starred' || view === 'trash') return view
  return 'all'
})

const viewTitle = computed(() => {
  switch (activeView.value) {
    case 'recent':
      return 'Recent'
    case 'starred':
      return 'Starred'
    case 'trash':
      return 'Trash'
    default:
      return 'All files'
  }
})

const displayItems = computed(() => {
  if (activeView.value === 'trash') return trashItems.value
  if (activeView.value === 'recent') return recentItems.value
  if (activeView.value === 'starred') return starredItems.value
  return items.value
})

const displayLoading = computed(() => {
  if (activeView.value === 'trash') return trashLoading.value
  if (activeView.value === 'recent' || activeView.value === 'starred') return quickAccessLoading.value
  return loading.value
})

const showCreateFolder = ref(false)

const showDeleteDialog = ref(false)
const showUploadFile = ref(false)
const showPropertiesModal = ref(false)
const propertiesItemId = ref<string | null>(null)

// Trash-specific state
const showRestoreDialog = ref(false)
const selectedRestoreFolder = ref<string | null>(null)
const newItemName = ref('')
const restoreTarget = ref<{ id: string; name: string } | null>(null)
const contextTarget = ref<{ id: string; name: string } | null>(null)

const actions = useItemActions({
  displayItems: computed(() => displayItems.value),
  isTrashView,
  onNavigate: (id, path) => navigateToFolder(id, path),
  onDeleteRequest: (item) => {
    contextTarget.value = item
    showDeleteDialog.value = true
  },
  onRestore: (item) => handleRestore(item),
  onPermanentDelete: (item) => handlePermanentDelete(item),
  onTrackAccess: async (itemId) => {
    const item = displayItems.value.find(candidate => candidate.id === itemId)
    if (!item || item.is_folder) return
    await track(itemId, 'open')
  },
  onToggleStar: async (item) => {
    await toggleStar(item.id)
    message.success(isStarred(item.id) ? 'Added to starred' : 'Removed from starred')
  },
  isStarred,
})

const showRenameDialog = actions.showRenameDialog

function onBreadcrumbClick(index: number) {
  const crumb = breadcrumbs.value[index]
  if (!crumb) return
  if (index === breadcrumbs.value.length - 1) return
  navigateToFolder(crumb.id, crumb.path)
}

function onViewChange(view: 'all' | 'recent' | 'starred') {
  if (view === 'all') {
    router.push({ path: '/', query: currentFolderId.value ? { folder: currentFolderId.value } : {} })
    return
  }
  router.push({ path: '/', query: { view } })
}

async function handleCreate(name: string, parentId?: string) {
  try {
    await createFolder({ name, parent_id: parentId })
    message.success('Folder created')
    showCreateFolder.value = false
    await refreshQuickAccess()
  } catch (e: any) {
    message.error(e?.data?.error || 'Failed to create folder')
  }
}

async function handleRename(name: string) {
  if (!contextTarget.value) return
  try {
    await updateItem(contextTarget.value.id, { name })
    message.success('Renamed successfully')
    actions.showRenameDialog.value = false
    await loadItems(currentFolderId.value)
    await refreshQuickAccess()
  } catch (e: any) {
    message.error(e?.data?.error || 'Failed to rename')
  }
}

async function handleDelete() {
  if (!contextTarget.value) return
  try {
    await deleteItem(contextTarget.value.id)
    message.success('Deleted successfully')
    showDeleteDialog.value = false
    removeQuickAccessItem(contextTarget.value.id)
    await refreshQuickAccess()
  } catch (e: any) {
    message.error(e?.data?.error || 'Failed to delete')
  }
}

async function handleUpload(file: File, parentId?: string, onProgress?: (progress: number) => void) {
  try {
    const uploadedItem = await uploadFile(file, parentId || undefined, onProgress)
    if (uploadedItem?.id) {
      await track(uploadedItem.id, 'update')
    }
    message.success('File uploaded successfully')
    showUploadFile.value = false
  } catch (e: any) {
    console.error('Upload error:', e)
    message.error(e?.data?.error || 'Failed to upload file')
  }
}

// Trash handlers
async function loadTrash() {
  folderStore.setTrashLoading(true)
  try {
    const data = await api.getTrash()
    folderStore.setTrashItems(data)
  } catch (e: any) {
    message.error(e?.data?.error || 'Failed to load trash')
  } finally {
    folderStore.setTrashLoading(false)
  }
}

async function handleRestore(item: Item) {
  try {
    await api.restoreItem(item.id)
    message.success('Restored successfully')
    folderStore.removeTrashItem(item.id)
    await refreshQuickAccess()
  } catch (e: any) {
    const code = e?.data?.code
    if (code === 'PARENT_DELETED') {
      restoreTarget.value = { id: item.id, name: item.name }
      showRestoreDialog.value = true
    } else if (code === 'NAME_CONFLICT') {
      restoreTarget.value = { id: item.id, name: item.name }
      newItemName.value = item.name
      actions.showRenameDialog.value = true
    } else {
      message.error(e?.data?.error || 'Failed to restore')
    }
  }
}

async function handleRestoreWithFolder(folderId: string | null) {
  if (!restoreTarget.value) return
  try {
    const body: RestoreItemRequest = { targetParentID: folderId || undefined }
    await api.restoreItem(restoreTarget.value.id, body)
    message.success('Restored successfully')
    folderStore.removeTrashItem(restoreTarget.value.id)
    showRestoreDialog.value = false
    restoreTarget.value = null
    await refreshQuickAccess()
  } catch (e: any) {
    if (e?.data?.code === 'NAME_CONFLICT' && restoreTarget.value) {
      newItemName.value = restoreTarget.value.name
      actions.showRenameDialog.value = true
      showRestoreDialog.value = false
    } else {
      message.error(e?.data?.error || 'Failed to restore')
    }
  }
}

async function handleRestoreWithRename() {
  if (!restoreTarget.value || !newItemName.value.trim()) return
  try {
    const body: RestoreItemRequest = { newName: newItemName.value.trim() }
    await api.restoreItem(restoreTarget.value.id, body)
    message.success('Restored successfully')
    folderStore.removeTrashItem(restoreTarget.value.id)
    actions.showRenameDialog.value = false
    restoreTarget.value = null
    await refreshQuickAccess()
  } catch (e: any) {
    message.error(e?.data?.error || 'Failed to restore')
  }
}

async function handlePermanentDelete(item: Item) {
  dialog.warning({
    title: 'Delete permanently',
    content: `Are you sure you want to permanently delete "${item.name}"? This action cannot be undone.`,
    positiveText: 'Delete permanently',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        await api.permanentDeleteItem(item.id)
        message.success('Deleted permanently')
        folderStore.removeTrashItem(item.id)
        removeQuickAccessItem(item.id)
        await refreshQuickAccess()
      } catch (e: any) {
        message.error(e?.data?.error || 'Failed to delete permanently')
      }
    },
  })
}

async function handleEmptyTrash() {
  dialog.warning({
    title: 'Empty Trash',
    content: `Are you sure you want to permanently delete all ${trashItems.value.length} items in the trash?`,
    positiveText: 'Delete all',
    negativeText: 'Cancel',
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

      await loadTrash()

      if (failCount === 0) {
        message.success('Trash emptied')
      } else if (successCount === 0) {
        message.error(`Failed to delete ${failCount} items`)
      } else {
        message.warning(`Deleted ${successCount} items, ${failCount} failed`)
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

watch(activeView, async (view) => {
  if (view === 'recent' || view === 'starred') {
    await refreshQuickAccess()
  }
}, { immediate: true })

// Sync properties modal from actions composable
watch(actions.propertiesItemId, (id) => {
  propertiesItemId.value = id
})

watch(actions.showPropertiesModal, (v) => {
  showPropertiesModal.value = v
})
</script>

<style scoped>
.file-manager {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 1.5rem 2rem;
}

@media (max-width: 768px) {
  .file-manager {
    padding: 0.75rem;
  }
}
</style>
