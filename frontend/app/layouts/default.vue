<template>
  <n-layout has-sider class="app-layout">
    <!-- Sidebar -->
    <n-layout-sider
      bordered
      :width="260"
      :collapsed-width="0"
      collapse-mode="width"
      :show-trigger="false"
      :collapsed="sidebarCollapsed"
      @collapse="sidebarCollapsed = true"
      @expand="sidebarCollapsed = false"
      class="app-sidebar"
      :class="{ 'is-mobile': !isDesktop }"
      :native-scrollbar="true"
    >
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <span class="logo-icon">☁️</span>
          <span class="logo-text text-gradient">1Drive</span>
        </div>
      </div>

      <!-- Folder Tree -->
      <div class="sidebar-content">
        <section class="sidebar-group">
          <div class="sidebar-section-title">
            <span>Điều hướng</span>
          </div>

          <div class="sidebar-section-card">
            <div
              class="nav-item"
              :class="{ active: isAllFilesActive }"
              @click="navigateToRoot"
            >
              <n-icon size="18"><Icon icon="mdi:home-variant" /></n-icon>
              <span>Tất cả file</span>
            </div>

            <div
              class="nav-item"
              :class="{ active: isSettingsActive }"
              @click="navigateToSettings"
            >
              <n-icon size="18"><Icon icon="mdi:cog-outline" /></n-icon>
              <span>Settings</span>
            </div>

            <div
              class="trash-item"
              :class="{ active: isTrashView }"
              @click="navigateToTrash"
            >
              <n-icon size="18"><Icon icon="mdi:trash-can-outline" /></n-icon>
              <span>Thùng rác</span>
              <n-badge
                v-if="trashItems.length > 0"
                :value="trashItems.length"
                :max="99"
                type="error"
                class="trash-badge"
              />
            </div>
          </div>
        </section>

        <section class="sidebar-group">
          <div class="sidebar-section-title">
            <span>Thư mục</span>
          </div>

          <div class="sidebar-section-card sidebar-tree-card">
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
        </section>

      </div>

      <!-- Footer: User info -->
      <div class="sidebar-footer">
        <div class="sidebar-footer-card" v-if="currentUser">
          <div class="user-info">
            <n-avatar :size="32" round>
              {{ currentUser.full_name?.charAt(0) || '?' }}
            </n-avatar>
            <div class="user-details">
              <span class="user-name">{{ currentUser.full_name }}</span>
              <span class="user-email">{{ currentUser.email }}</span>
            </div>
            <n-button
              quaternary
              circle
              size="small"
              @click="handleLogout"
            >
              <template #icon>
                <n-icon><Icon icon="mdi:logout" /></n-icon>
              </template>
            </n-button>
          </div>

          <div class="storage-summary">
            <div class="storage-copy">
              <span class="storage-label">Storage</span>
              <strong>{{ storageUsageText }}</strong>
            </div>
            <n-progress
              type="line"
              :percentage="storageUsagePercent"
              :show-indicator="false"
              :height="6"
              :border-radius="999"
              status="success"
            />
            <small class="storage-meta">{{ currentUser.role }} · {{ currentUser.is_verified ? 'Đã xác minh' : 'Chưa xác minh' }}</small>
          </div>
        </div>
      </div>
    </n-layout-sider>

    <button
      v-if="isDesktop"
      type="button"
      class="sidebar-edge-toggle"
      :class="{ collapsed: sidebarCollapsed }"
      :aria-label="sidebarCollapsed ? 'Mở sidebar' : 'Thu gọn sidebar'"
      @click="sidebarCollapsed = !sidebarCollapsed"
    >
      <n-icon size="18">
        <span class="sidebar-edge-toggle-icon" :class="{ collapsed: sidebarCollapsed }">
          <Icon icon="mdi:chevron-left" />
        </span>
      </n-icon>
    </button>

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
        <div class="mobile-header-main">
          <n-button
            quaternary
            circle
            size="medium"
            class="mobile-menu-btn"
            @click="sidebarCollapsed = !sidebarCollapsed"
          >
            <template #icon>
              <n-icon><Icon icon="mdi:menu" /></n-icon>
            </template>
          </n-button>

          <div class="mobile-header-copy">
            <strong>{{ mobileHeaderTitle }}</strong>
            <span>{{ currentUser?.full_name || '1Drive' }}</span>
          </div>
        </div>

        <div class="mobile-header-actions">
          <n-button
            quaternary
            circle
            size="medium"
            class="mobile-menu-btn"
            @click="openSearchPalette"
          >
            <template #icon>
              <n-icon><Icon icon="mdi:magnify" /></n-icon>
            </template>
          </n-button>
        </div>
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

    <!-- Global File Preview Modal -->
    <FilePreviewModal />

    <!-- Global Search Palette -->
    <SearchPalette />
  </n-layout>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '~/stores/auth'
import { useFolderStore } from '~/stores/folder'
import type { FolderTreeNode } from '~/types/folder'
import SearchPalette from '~/components/search/SearchPalette.vue'
import '~/assets/css/search-palette.css'

const authStore = useAuthStore()
const { currentUser } = storeToRefs(authStore)
const { logout } = useAuth()
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
const {
  refresh: refreshQuickAccess,
} = useQuickAccess()
const { open: openSearchPalette } = useSearchPalette()

const route = useRoute()
const store = useFolderStore()
const { isTrashView, trashItems } = storeToRefs(store)

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

// Watch for trash view changes
watch(() => route.query.view, (view) => {
  store.setTrashView(view === 'trash')
}, { immediate: true })

watch(currentUser, async () => {
  await refreshQuickAccess()
}, { immediate: true })

const router = useRouter()
const isSettingsActive = computed(() => route.path === '/settings')
const isAllFilesActive = computed(() => route.path === '/' && !isTrashView.value)
const mobileHeaderTitle = computed(() => {
  if (route.path === '/settings') return 'Settings'
  if (route.path === '/office') return 'Office'
  if (isTrashView.value) return 'Thùng rác'
  return '1Drive'
})
const storageUsagePercent = computed(() => {
  const used = currentUser.value?.storage_used_bytes || 0
  const quota = currentUser.value?.storage_quota_bytes || 0
  if (!quota) return 0
  return Math.min(100, Math.round((used / quota) * 100))
})
const storageUsageText = computed(() =>
  `${formatBytes(currentUser.value?.storage_used_bytes || 0)} / ${formatBytes(currentUser.value?.storage_quota_bytes || 0)}`,
)

function navigateToRoot() {
  router.push({ path: '/', query: currentFolderId.value ? { folder: currentFolderId.value } : {} })
}

function navigateToTrash() {
  router.push({ path: '/', query: { view: 'trash' } })
}

function navigateToSettings() {
  router.push({ path: '/settings' })
}

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
  await refreshQuickAccess()

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

async function handleLogout() {
  await logout()
}

function formatBytes(bytes: number) {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let index = 0
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024
    index += 1
  }
  return `${value >= 10 || index === 0 ? value.toFixed(0) : value.toFixed(1)} ${units[index]}`
}
</script>

<style scoped>
.app-layout {
  position: relative;
  height: 100vh;
  overflow: hidden;
  background: var(--color-bg-primary);
}

.app-sidebar {
  position: relative;
  overflow: visible !important;
  background: var(--color-bg-secondary) !important;
  border-right: 1px solid var(--color-border) !important;
}

.app-sidebar :deep(.n-layout-sider-scroll-container) {
  display: flex !important;
  flex-direction: column !important;
  height: 100% !important;
  overflow: hidden !important;
}

.sidebar-header {
  padding: 1.1rem 1rem 0.9rem;
  border-bottom: 1px solid var(--color-border);
}

.nav-item:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
}

.nav-item.active {
  background: rgba(59, 130, 246, 0.14);
  color: var(--color-primary);
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
  min-height: 0;
  overflow-y: auto;
  padding: 0.85rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
}

.sidebar-group {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.sidebar-section-card,
.sidebar-footer-card {
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-lg);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.035), rgba(255, 255, 255, 0.02));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
}

.sidebar-tree-card {
  padding: 0.35rem 0;
}

.sidebar-section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0.25rem;
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-top: 0.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  border-radius: var(--radius-sm);
  margin: 0.3rem;
  font-size: var(--font-size-sm);
}

.nav-item:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
}

.nav-item.active {
  background: var(--color-primary-light);
  color: var(--color-primary);
  box-shadow: inset 3px 0 0 var(--color-primary);
}

.trash-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all var(--transition-fast);
  border-radius: var(--radius-sm);
  margin: 0.3rem;
  font-size: var(--font-size-sm);
  position: relative;
}

.trash-item:hover {
  background: var(--color-surface-hover);
  color: var(--color-text-primary);
}

.trash-item.active {
  background: var(--color-error-light);
  color: var(--color-error);
  box-shadow: inset 3px 0 0 var(--color-error);
}

.trash-badge {
  margin-left: auto;
}

.sidebar-footer {
  margin-top: auto;
  padding: 0.75rem;
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar-footer-card {
  padding: 0.8rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.8rem;
}

.user-details {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
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

.storage-summary {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.storage-copy {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
}

.storage-label {
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.storage-copy strong {
  font-size: var(--font-size-sm);
}

.storage-meta {
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
}

.app-main {
  background: var(--color-bg-primary) !important;
  height: 100% !important;
}

.sidebar-edge-toggle {
  position: absolute;
  top: 50%;
  left: 244px;
  transform: translateY(-50%);
  z-index: 30;
  width: 32px;
  height: 56px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-left: 0;
  border-radius: 0 16px 16px 0;
  background:
    linear-gradient(180deg, rgba(36, 39, 49, 0.95), rgba(22, 25, 34, 0.92));
  color: var(--color-text-secondary);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 10px 30px rgba(3, 7, 18, 0.38),
    inset 0 1px 0 rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(16px);
  transition:
    left var(--transition-base),
    color var(--transition-fast),
    background var(--transition-fast),
    box-shadow var(--transition-fast),
    transform var(--transition-fast);
}

.sidebar-edge-toggle:hover {
  color: var(--color-text-primary);
  background:
    linear-gradient(180deg, rgba(48, 54, 68, 0.98), rgba(27, 31, 42, 0.96));
  box-shadow:
    0 12px 34px rgba(3, 7, 18, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.sidebar-edge-toggle:active {
  transform: translateY(-50%) scale(0.98);
}

.sidebar-edge-toggle.collapsed {
  left: 0;
  border-left: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0 16px 16px 0;
}

.sidebar-edge-toggle-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: transform 180ms ease;
}

.sidebar-edge-toggle-icon.collapsed {
  transform: rotate(180deg);
}

.app-main :deep(.n-layout-scroll-container) {
  height: 100% !important;
  display: flex !important;
  flex-direction: column !important;
  overflow: hidden !important;
}

.mobile-header {
  display: none;
}

.main-content {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  overflow-x: hidden;
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .app-sidebar {
    position: fixed !important;
    z-index: 1000;
    height: 100vh !important;
    transition: transform var(--transition-base);
  }

  .sidebar-edge-toggle {
    display: none;
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
    justify-content: space-between;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--color-border);
    position: sticky;
    top: 0;
    z-index: 20;
    background: rgba(10, 10, 15, 0.82);
    backdrop-filter: blur(18px);
  }

  .mobile-header-main {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    min-width: 0;
    flex: 1;
  }

  .mobile-menu-btn {
    flex-shrink: 0;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  .mobile-header-copy {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .mobile-header-copy strong {
    font-size: var(--font-size-base);
    line-height: 1.1;
    letter-spacing: -0.02em;
  }

  .mobile-header-copy span {
    color: var(--color-text-muted);
    font-size: var(--font-size-xs);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .mobile-header-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
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
    padding: 0.6rem;
  }

  .sidebar-section-title,
  .tree-root-item {
    padding: 0.5rem 0.75rem;
  }

  .sidebar-footer {
    padding: 0.6rem;
  }

  .sidebar-footer-card {
    padding: 0.7rem;
  }

  .user-info {
    align-items: center;
    gap: 0.6rem;
    margin-bottom: 0.6rem;
  }

  .user-details {
    align-items: flex-start;
  }

  .storage-copy {
    gap: 0.5rem;
  }

  .user-name {
    font-size: 0.9rem;
  }

  .user-email {
    max-width: 140px;
  }

  .storage-summary {
    gap: 0.35rem;
  }

  .storage-label,
  .storage-meta {
    font-size: 11px;
  }
}

@media (max-width: 480px) {
  .logo-icon {
    font-size: 1.25rem;
  }

  .logo-text {
    font-size: var(--font-size-base);
  }
}
</style>
