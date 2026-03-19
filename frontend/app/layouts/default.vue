<template>
  <n-layout has-sider class="app-layout">
    <!-- Sidebar -->
    <n-layout-sider
      bordered
      :width="260"
      :collapsed-width="0"
      collapse-mode="width"
      :show-trigger="isDesktop ? 'arrow-circle' : false"
      :collapsed="sidebarCollapsed"
      @collapse="sidebarCollapsed = true"
      @expand="sidebarCollapsed = false"
      class="app-sidebar"
      :class="{ 'is-mobile': !isDesktop }"
      :native-scrollbar="false"
    >
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <span class="logo-icon">☁️</span>
          <span class="logo-text text-gradient">1Drive</span>
        </div>
      </div>

      <!-- Folder Tree -->
      <div class="sidebar-content">
        <div class="sidebar-section-title">
          <span>Thư mục</span>
          <n-button
            quaternary
            circle
            size="tiny"
            @click="showCreateFolder = true"
          >
            <template #icon>
              <n-icon><Icon icon="mdi:folder-plus-outline" /></n-icon>
            </template>
          </n-button>
        </div>

        <!-- Root item -->
        <div
          class="tree-root-item"
          :class="{ active: !currentFolderId }"
          @click="navigateToFolder(null, '/')"
        >
          <n-icon size="18"><Icon icon="mdi:home-outline" /></n-icon>
          <span>Tất cả file</span>
        </div>

        <ClientOnly>
          <FolderTree
            :tree="folderTree"
            :selected-id="currentFolderId"
            :loading="treeLoading"
            @select="onFolderSelect"
            @create="onCreateInFolder"
            @rename="onRenameFolder"
            @delete="onDeleteFolder"
          />
        </ClientOnly>
      </div>

      <!-- User info -->
      <div class="sidebar-footer">
        <div class="user-info" v-if="currentUser">
          <n-avatar :size="32" round>
            {{ currentUser.full_name?.charAt(0) || '?' }}
          </n-avatar>
          <div class="user-details">
            <span class="user-name">{{ currentUser.full_name }}</span>
            <span class="user-email">{{ currentUser.email }}</span>
          </div>
        </div>
      </div>
    </n-layout-sider>

    <!-- Mobile overlay -->
    <div
      v-if="!isDesktop && !sidebarCollapsed"
      class="sidebar-overlay"
      @click="sidebarCollapsed = true"
    />

    <!-- Main Content -->
    <n-layout class="app-main">
      <!-- Mobile header with menu button -->
      <div class="mobile-header">
        <n-button
          quaternary
          circle
          size="medium"
          @click="sidebarCollapsed = !sidebarCollapsed"
        >
          <template #icon>
            <n-icon><Icon icon="mdi:menu" /></n-icon>
          </template>
        </n-button>
      </div>

      <div class="main-content">
        <slot />
      </div>
    </n-layout>

    <!-- Create Folder Modal -->
    <FolderActions
      v-model:show-create="showCreateFolder"
      v-model:show-rename="showRenameFolder"
      v-model:show-delete="showDeleteFolder"
      v-model:show-upload="showUploadFile"
      :target-item="targetItem"
      :parent-id="createParentId"
      @create="handleCreateFolder"
      @rename="handleRenameFolder"
      @delete="handleDeleteFolder"
    />
  </n-layout>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { FolderTreeNode } from '~/types/folder'

const { currentUser } = useAuth()
const {
  loadItems,
  loadFolderTree,
  createFolder,
  updateItem,
  deleteItem,
  navigateToFolder,
  folderTree,
  currentFolderId,
  treeLoading,
} = useFolder()

const route = useRoute()
const store = useFolderStore()

// Watch for URL folder param changes (browser back/forward)
watch(() => route.query.folder, (newFolderId) => {
  const folderId = newFolderId as string | undefined

  if (!folderId) {
    store.setCurrentFolder(null, '/')
    loadItems(null)
    return
  }

  store.setCurrentFolder(folderId, '/')
  loadItems(folderId)
})

const showCreateFolder = ref(false)
const showRenameFolder = ref(false)
const showDeleteFolder = ref(false)
const showUploadFile = ref(false)
const targetItem = ref<{ id: string; name: string } | null>(null)
const createParentId = ref<string | undefined>(undefined)

const message = useMessage()

// Responsive state
const isDesktop = ref(true)
const sidebarCollapsed = ref(false)

const checkScreenSize = () => {
  isDesktop.value = window.innerWidth > 768
  if (!isDesktop.value) {
    sidebarCollapsed.value = true
  }
}

onMounted(async () => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)

  // Load tree first
  await loadFolderTree()

  // Load items - root or folder from URL
  const folderId = route.query.folder as string | undefined
  if (folderId) {
    store.setCurrentFolder(folderId, '/')
    await loadItems(folderId)
  } else {
    await loadItems()
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize)
})

function onFolderSelect(node: FolderTreeNode) {
  navigateToFolder(node.id, node.path)
}

function onCreateInFolder(parentId: string) {
  createParentId.value = parentId
  showCreateFolder.value = true
}

function onRenameFolder(node: FolderTreeNode) {
  targetItem.value = { id: node.id, name: node.name }
  showRenameFolder.value = true
}

function onDeleteFolder(node: FolderTreeNode) {
  targetItem.value = { id: node.id, name: node.name }
  showDeleteFolder.value = true
}

async function handleCreateFolder(name: string, parentId?: string) {
  try {
    await createFolder({ name, parent_id: parentId })
    message.success('Đã tạo thư mục')
    showCreateFolder.value = false
    createParentId.value = undefined
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể tạo thư mục')
  }
}

async function handleRenameFolder(name: string) {
  if (!targetItem.value) return
  try {
    await updateItem(targetItem.value.id, { name })
    message.success('Đã đổi tên')
    showRenameFolder.value = false
    // Reload current items if we're inside the renamed folder
    await loadItems(currentFolderId.value)
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể đổi tên')
  }
}

async function handleDeleteFolder() {
  if (!targetItem.value) return
  try {
    await deleteItem(targetItem.value.id)
    message.success('Đã xóa thư mục')
    showDeleteFolder.value = false
    // If deleted the current folder, go back to root
    if (currentFolderId.value === targetItem.value.id) {
      navigateToFolder(null, '/')
    }
  } catch (e: any) {
    message.error(e?.data?.error || 'Không thể xóa')
  }
}
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  background: var(--color-bg-primary);
}

.app-sidebar {
  background: var(--color-bg-secondary) !important;
  border-right: 1px solid var(--color-border) !important;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid var(--color-border);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.logo-icon {
  font-size: 1.5rem;
}

.logo-text {
  font-size: var(--font-size-xl);
  font-weight: 700;
  letter-spacing: -0.02em;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem 0;
}

.sidebar-section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 1rem;
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.tree-root-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  border-radius: var(--radius-sm);
  margin: 0 0.5rem;
  font-size: var(--font-size-sm);
}

.tree-root-item:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
}

.tree-root-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.sidebar-footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--color-border);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-details {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-name {
  font-size: var(--font-size-sm);
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-email {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.app-main {
  background: var(--color-bg-primary) !important;
}

.mobile-header {
  display: none;
}

.main-content {
  padding: 1.5rem 2rem;
  min-height: 100vh;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .app-sidebar {
    position: fixed !important;
    z-index: 1000;
    height: 100vh !important;
    transition: transform var(--transition-base);
  }

  .app-sidebar.is-mobile {
    transform: translateX(-100%);
  }

  .app-sidebar.is-mobile:not(.n-layout-sider--collapsed) {
    transform: translateX(0);
  }

  .sidebar-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 999;
  }

  .mobile-header {
    display: flex;
    align-items: center;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--color-border);
  }

  .main-content {
    padding: 1rem;
  }

  .sidebar-header {
    padding: 1rem;
  }

  .sidebar-logo {
    justify-content: center;
  }

  .logo-text {
    font-size: var(--font-size-lg);
  }

  .sidebar-content {
    padding: 0.5rem 0;
  }

  .sidebar-section-title,
  .tree-root-item {
    padding: 0.5rem 0.75rem;
  }

  .sidebar-footer {
    padding: 0.5rem 0.75rem;
  }

  .user-info {
    flex-direction: column;
    text-align: center;
    gap: 0.5rem;
  }

  .user-details {
    align-items: center;
  }
}

@media (max-width: 480px) {
  .main-content {
    padding: 0.75rem;
  }

  .logo-icon {
    font-size: 1.25rem;
  }

  .logo-text {
    font-size: var(--font-size-base);
  }
}
</style>
