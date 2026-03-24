<template>
  <div class="file-manager">
    <!-- Header -->
    <FileManagerHeader
      :is-trash-view="isTrashView"
      :breadcrumbs="breadcrumbs"
      :has-trash-items="trashItems.length > 0"
      @breadcrumb-click="onBreadcrumbClick"
      @empty-trash="handleEmptyTrash"
      @upload-click="showUploadFile = true"
      @create-folder-click="showCreateFolder = true"
    />

    <!-- Items View (Grid/List/Column) -->
    <FileManagerView
      :items="displayItems"
      :is-trash-view="isTrashView"
      :actions="actions"
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
})

const showRenameDialog = actions.showRenameDialog

function onBreadcrumbClick(index: number) {
  const crumb = breadcrumbs.value[index]
  if (!crumb) return
  if (index === breadcrumbs.value.length - 1) return
  navigateToFolder(crumb.id, crumb.path)
}

async function handleCreate(name: string, parentId?: string) {
  try {
    await createFolder({ name, parent_id: parentId })
    message.success('Folder created')
    showCreateFolder.value = false
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
  } catch (e: any) {
    message.error(e?.data?.error || 'Failed to delete')
  }
}

async function handleUpload(file: File, parentId?: string, onProgress?: (progress: number) => void) {
  try {
    await uploadFile(file, parentId || undefined, onProgress)
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
  width: 100%;
  min-width: 0;
}
</style>
