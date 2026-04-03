<template>
  <div class="share-page">
    <div class="share-shell">
      <div class="share-hero">
        <div class="share-hero-copy">
          <span class="share-kicker">1Drive Public Link</span>
          <h1>{{ heroTitle }}</h1>
          <p>{{ heroDescription }}</p>
        </div>
        <div class="share-hero-meta">
          <n-tag :type="statusTagType" round>{{ detail?.status || 'unknown' }}</n-tag>
          <span v-if="detail?.expires_at" class="share-meta-text">Hết hạn {{ formatDate(detail.expires_at) }}</span>
          <span v-if="detail?.session_expires_at" class="share-meta-text">Session tới {{ formatDate(detail.session_expires_at) }}</span>
        </div>
      </div>

      <n-spin v-if="loading" size="large" class="share-loading" />

      <n-result
        v-else-if="fatalState"
        status="warning"
        :title="fatalState.title"
        :description="fatalState.description"
        class="share-result"
      />

      <div v-else-if="detail && !detail.access_granted" class="share-lock-card">
        <div class="share-lock-copy">
          <span class="share-kicker">Protected link</span>
          <h2>Nhập password để tiếp tục</h2>
          <p>Nội dung file hoặc folder sẽ không được hiển thị cho đến khi password hợp lệ.</p>
        </div>

        <n-input
          v-model:value="password"
          type="password"
          show-password-on="click"
          placeholder="Nhập password chia sẻ"
          @keydown.enter="authenticate"
        />

        <div class="share-lock-actions">
          <n-button type="primary" :loading="authenticating" @click="authenticate">Mở link</n-button>
        </div>
      </div>

      <div v-else-if="detail?.item" class="share-content">
        <template v-if="detail.item.is_folder">
          <div class="share-browser-panel">
            <div class="share-browser-top">
              <div>
                <span class="share-kicker">Shared folder</span>
                <h2>{{ folderData?.current_folder.name || detail.item.name }}</h2>
              </div>
              <n-button size="small" tertiary :loading="folderLoading" @click="loadFolder(currentFolderId)">
                Làm mới
              </n-button>
            </div>

            <div class="share-breadcrumbs">
              <button
                v-for="crumb in folderData?.breadcrumbs || []"
                :key="crumb.id"
                type="button"
                class="share-crumb"
                :class="{ active: crumb.id === currentFolderId }"
                @click="loadFolder(crumb.id)"
              >
                {{ crumb.name }}
              </button>
            </div>

            <div v-if="folderData?.items.length" class="share-items">
              <button
                v-for="item in folderData.items"
                :key="item.id"
                type="button"
                class="share-item-row"
                :class="{ active: selectedFile?.id === item.id }"
                @click="openSharedItem(item)"
              >
                <div class="share-item-main">
                  <span class="share-item-icon">{{ item.is_folder ? '📁' : fileEmoji(item) }}</span>
                  <div class="share-item-copy">
                    <strong>{{ item.name }}</strong>
                    <small>{{ item.is_folder ? 'Folder' : (item.mime_type || 'File') }}</small>
                  </div>
                </div>
                <span class="share-item-meta">{{ item.is_folder ? '' : formatSize(item.size) }}</span>
              </button>
            </div>

            <div v-else class="share-empty-state">
              Folder này chưa có item nào.
            </div>
          </div>

          <div class="share-viewer-panel">
            <div class="share-viewer-head">
              <div>
                <span class="share-kicker">Preview</span>
                <h2>{{ selectedFile?.name || 'Chọn file để xem trước' }}</h2>
              </div>
              <div v-if="selectedFile" class="share-viewer-actions">
                <n-button tertiary @click="downloadSelected">Download</n-button>
                <n-button
                  v-if="isOfficeFile(selectedFile.name)"
                  type="primary"
                  @click="openOfficeViewer(selectedFile)"
                >
                  Open Office viewer
                </n-button>
              </div>
            </div>

            <div v-if="!selectedFile" class="share-empty-state share-empty-state--viewer">
              Chọn một file trong danh sách bên trái để preview hoặc download.
            </div>

            <template v-else>
              <div v-if="previewLoading" class="share-preview-loading">
                <n-spin size="large" />
              </div>

              <div v-else-if="selectedPreviewType === 'image'" class="share-media-stage">
                <img :src="selectedStreamUrl" :alt="selectedFile.name" class="share-image" />
              </div>

              <div v-else-if="selectedPreviewType === 'video'" class="share-media-stage">
                <video :src="selectedStreamUrl" controls class="share-video" />
              </div>

              <div v-else-if="selectedPreviewType === 'audio'" class="share-audio-stage">
                <audio :src="selectedStreamUrl" controls class="share-audio" />
              </div>

              <div v-else-if="selectedPreviewType === 'pdf'" class="share-pdf-stage">
                <iframe :src="selectedStreamUrl" class="share-pdf" title="PDF preview" />
              </div>

              <div v-else-if="selectedPreviewType === 'text' || selectedPreviewType === 'markdown'" class="share-text-stage">
                <pre>{{ textPreview }}</pre>
              </div>

              <div v-else-if="selectedPreviewType === 'office'" class="share-office-stage">
                <p>File Office được mở qua OnlyOffice ở chế độ chỉ đọc.</p>
                <n-button type="primary" @click="openOfficeViewer(selectedFile)">Mở viewer</n-button>
              </div>

              <div v-else class="share-empty-state share-empty-state--viewer">
                {{ previewUnavailableMessage }}
              </div>
            </template>
          </div>
        </template>

        <template v-else>
          <div class="share-single-panel">
            <div class="share-viewer-head">
              <div>
                <span class="share-kicker">Shared file</span>
                <h2>{{ detail.item.name }}</h2>
              </div>
              <div class="share-viewer-actions">
                <n-button tertiary @click="downloadRoot">Download</n-button>
                <n-button
                  v-if="isOfficeFile(detail.item.name)"
                  type="primary"
                  @click="openOfficeViewer(detail.item)"
                >
                  Open Office viewer
                </n-button>
              </div>
            </div>

            <div v-if="previewLoading" class="share-preview-loading">
              <n-spin size="large" />
            </div>

            <div v-else-if="selectedPreviewType === 'image'" class="share-media-stage">
              <img :src="selectedStreamUrl" :alt="detail.item.name" class="share-image" />
            </div>

            <div v-else-if="selectedPreviewType === 'video'" class="share-media-stage">
              <video :src="selectedStreamUrl" controls class="share-video" />
            </div>

            <div v-else-if="selectedPreviewType === 'audio'" class="share-audio-stage">
              <audio :src="selectedStreamUrl" controls class="share-audio" />
            </div>

            <div v-else-if="selectedPreviewType === 'pdf'" class="share-pdf-stage">
              <iframe :src="selectedStreamUrl" class="share-pdf" title="PDF preview" />
            </div>

            <div v-else-if="selectedPreviewType === 'text' || selectedPreviewType === 'markdown'" class="share-text-stage">
              <pre>{{ textPreview }}</pre>
            </div>

            <div v-else-if="selectedPreviewType === 'office'" class="share-office-stage">
              <p>File Office được mở qua OnlyOffice ở chế độ chỉ đọc.</p>
              <n-button type="primary" @click="openOfficeViewer(detail.item)">Mở viewer</n-button>
            </div>

            <div v-else class="share-empty-state share-empty-state--viewer">
              {{ previewUnavailableMessage }}
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PublicFolderListing, PublicSharedItem } from '~/types/public-link'

definePageMeta({
  layout: false,
})

const route = useRoute()
const message = useMessage()
const publicApi = usePublicLinkApi()
const { getPreviewType, isOfficeFile } = usePreview()

const token = computed(() => String(route.params.token || ''))
const password = ref('')
const loading = ref(true)
const authenticating = ref(false)
const folderLoading = ref(false)
const previewLoading = ref(false)
const detail = ref<any | null>(null)
const folderData = ref<PublicFolderListing | null>(null)
const selectedFile = ref<PublicSharedItem | null>(null)
const textPreview = ref('')
const fatalState = ref<{ title: string; description: string } | null>(null)

const sessionToken = ref('')
const currentFolderId = computed(() => folderData.value?.current_folder.id || detail.value?.item?.id || '')
const selectedPreviewType = computed(() => {
  const item = selectedFile.value || detail.value?.item
  return getPreviewType(item?.name, item?.mime_type)
})
const previewUnavailableMessage = computed(() => {
  const item = selectedFile.value || detail.value?.item
  if (!item) {
    return 'Chưa có preview inline cho định dạng này. Bạn vẫn có thể tải file xuống.'
  }

  const mimeType = item.mime_type?.toLowerCase() || ''
  const extension = item.name.split('.').pop()?.toLowerCase() || ''
  const looksLikeMedia = mimeType.startsWith('video/')
    || mimeType.startsWith('audio/')
    || ['avi', 'flv', 'mkv', 'wmv', 'wma'].includes(extension)

  if (looksLikeMedia && selectedPreviewType.value === 'unknown') {
    return 'Trình duyệt hiện tại không hỗ trợ preview inline định dạng media này. Bạn vẫn có thể tải file xuống.'
  }

  return 'Chưa có preview inline cho định dạng này. Bạn vẫn có thể tải file xuống.'
})
const selectedStreamUrl = computed(() => {
  const itemId = selectedFile.value?.id || detail.value?.item?.id
  return publicApi.buildStreamUrl(token.value, itemId, sessionToken.value || undefined, false)
})

const heroTitle = computed(() => {
  if (fatalState.value) return fatalState.value.title
  if (detail.value?.item?.name) return detail.value.item.name
  if (detail.value?.requires_password) return 'Protected public link'
  return 'Shared content'
})

const heroDescription = computed(() => {
  if (fatalState.value) return fatalState.value.description
  if (detail.value?.item?.is_folder) return 'Duyệt item con trong folder được chia sẻ và preview file trực tiếp qua app policy.'
  if (detail.value?.item) return 'Preview hoặc download file trực tiếp. Quyền truy cập được kiểm tra ở mọi request.'
  if (detail.value?.requires_password) return 'Password protection đang bật. Nội dung chỉ hiển thị sau khi xác thực thành công.'
  return 'Public link'
})

const statusTagType = computed(() => {
  switch (detail.value?.status) {
    case 'active':
      return 'success'
    case 'expired':
      return 'warning'
    case 'revoked':
      return 'error'
    default:
      return 'default'
  }
})

useHead({
  title: computed(() => detail.value?.item?.name ? `${detail.value.item.name} - Shared via 1Drive` : '1Drive Shared Link'),
})

onMounted(async () => {
  if (import.meta.client) {
    sessionToken.value = sessionStorage.getItem(sessionStorageKey(token.value)) || ''
  }
  await loadLink()
})

watch(selectedFile, async (file) => {
  if (!file) {
    textPreview.value = ''
    return
  }

  if (selectedPreviewType.value === 'text' || selectedPreviewType.value === 'markdown') {
    previewLoading.value = true
    try {
      textPreview.value = await fetch(selectedStreamUrl.value).then(response => response.text())
    } catch {
      textPreview.value = 'Không tải được nội dung file.'
    } finally {
      previewLoading.value = false
    }
    return
  }

  textPreview.value = ''
  previewLoading.value = false
})

async function loadLink() {
  loading.value = true
  fatalState.value = null

  try {
    detail.value = await publicApi.getPublicLink(token.value, sessionToken.value || undefined)

    if (detail.value.access_granted && detail.value.item?.is_folder) {
      await loadFolder()
      selectedFile.value = null
    } else if (detail.value.access_granted && detail.value.item && !detail.value.item.is_folder) {
      selectedFile.value = detail.value.item
    } else {
      folderData.value = null
      selectedFile.value = null
    }
  } catch (error: any) {
    fatalState.value = mapPublicError(error)
    clearSession()
  } finally {
    loading.value = false
  }
}

async function authenticate() {
  if (!password.value.trim()) {
    message.error('Nhập password để tiếp tục')
    return
  }

  authenticating.value = true
  try {
    const response = await publicApi.authenticatePublicLink(token.value, { password: password.value })
    sessionToken.value = response.session_token
    if (import.meta.client) {
      sessionStorage.setItem(sessionStorageKey(token.value), response.session_token)
    }
    password.value = ''
    await loadLink()
  } catch (error: any) {
    message.error(error?.data?.error || 'Password không đúng hoặc link không còn hiệu lực')
  } finally {
    authenticating.value = false
  }
}

async function loadFolder(parentId?: string) {
  if (!detail.value?.item?.is_folder) return

  folderLoading.value = true
  try {
    folderData.value = await publicApi.listSharedItems(token.value, sessionToken.value || undefined, parentId)
    const selectedStillVisible = folderData.value.items.find(item => item.id === selectedFile.value?.id && !item.is_folder)
    if (!selectedStillVisible) {
      selectedFile.value = null
    }
  } catch (error: any) {
    const state = mapPublicError(error)
    if (state) {
      fatalState.value = state
    }
  } finally {
    folderLoading.value = false
  }
}

function openSharedItem(item: PublicSharedItem) {
  if (item.is_folder) {
    loadFolder(item.id)
    return
  }
  selectedFile.value = item
}

function downloadSelected() {
  if (!selectedFile.value) return
  window.open(publicApi.buildStreamUrl(token.value, selectedFile.value.id, sessionToken.value || undefined, true), '_blank')
}

function downloadRoot() {
  if (!detail.value?.item) return
  window.open(publicApi.buildStreamUrl(token.value, detail.value.item.id, sessionToken.value || undefined, true), '_blank')
}

function openOfficeViewer(item: PublicSharedItem) {
  const params = new URLSearchParams({
    id: item.id,
    share_token: token.value,
    name: item.name,
  })
  if (sessionToken.value) {
    params.set('session', sessionToken.value)
  }
  window.open(`/office?${params.toString()}`, '_blank')
}

function clearSession() {
  sessionToken.value = ''
  if (import.meta.client) {
    sessionStorage.removeItem(sessionStorageKey(token.value))
  }
}

function sessionStorageKey(linkToken: string) {
  return `public-link-session:${linkToken}`
}

function mapPublicError(error: any) {
  const code = error?.data?.code
  switch (code) {
    case 'PUBLIC_LINK_EXPIRED':
      return {
        title: 'Link đã hết hạn',
        description: 'Public link này không còn hiệu lực vì đã vượt quá thời hạn chia sẻ.',
      }
    case 'PUBLIC_LINK_REVOKED':
      return {
        title: 'Link đã bị thu hồi',
        description: 'Chủ sở hữu đã revoke public link này. Mọi request mới đều bị từ chối ngay.',
      }
    case 'PUBLIC_LINK_NOT_FOUND':
      return {
        title: 'Link không tồn tại',
        description: 'URL này không khớp với public link nào còn hợp lệ.',
      }
    case 'PUBLIC_LINK_ACCESS_DENIED':
      clearSession()
      return {
        title: 'Access denied',
        description: 'Session của public link không còn hợp lệ hoặc request nằm ngoài policy được phép.',
      }
    default:
      return {
        title: 'Không mở được public link',
        description: error?.data?.error || 'Đã có lỗi xảy ra khi tải nội dung chia sẻ.',
      }
  }
}

function formatDate(value?: string | null) {
  if (!value) return 'Chưa có'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'Chưa có'
  return date.toLocaleString('vi-VN', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatSize(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const units = ['B', 'KB', 'MB', 'GB']
  const unitIndex = Math.min(units.length - 1, Math.floor(Math.log(bytes) / Math.log(k)))
  return `${(bytes / Math.pow(k, unitIndex)).toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function fileEmoji(item: Pick<PublicSharedItem, 'name' | 'mime_type'>) {
  switch (getPreviewType(item.name, item.mime_type)) {
    case 'image':
      return '🖼️'
    case 'video':
      return '🎬'
    case 'audio':
      return '🎧'
    case 'pdf':
      return '📕'
    case 'office':
      return '📄'
    case 'text':
    case 'markdown':
      return '📝'
    default:
      return '📦'
  }
}
</script>

<style scoped>
.share-page {
  min-height: 100vh;
  padding: 2rem 1.2rem 3rem;
  background:
    radial-gradient(circle at top left, rgba(248, 196, 113, 0.18), transparent 32%),
    radial-gradient(circle at top right, rgba(64, 224, 208, 0.14), transparent 28%),
    linear-gradient(180deg, #09111e 0%, #0d1728 48%, #08101b 100%);
  color: #f5f7fb;
}

.share-shell {
  width: min(1200px, 100%);
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.2rem;
}

.share-hero,
.share-lock-card,
.share-browser-panel,
.share-viewer-panel,
.share-single-panel {
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 28px;
  background: rgba(9, 17, 31, 0.72);
  box-shadow: 0 28px 60px rgba(0, 0, 0, 0.24);
  backdrop-filter: blur(18px);
}

.share-hero {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
  padding: 1.5rem 1.7rem;
}

.share-kicker {
  display: inline-block;
  margin-bottom: 0.45rem;
  font-size: 0.75rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.66);
}

.share-hero h1,
.share-lock-card h2,
.share-browser-top h2,
.share-viewer-head h2 {
  margin: 0;
  font-size: clamp(1.5rem, 2.2vw, 2.4rem);
  line-height: 1.05;
}

.share-hero p,
.share-lock-card p {
  margin: 0.45rem 0 0;
  max-width: 60ch;
  color: rgba(255, 255, 255, 0.74);
  line-height: 1.55;
}

.share-hero-meta {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  align-items: flex-end;
}

.share-meta-text {
  font-size: 0.9rem;
  color: rgba(255, 255, 255, 0.72);
}

.share-loading,
.share-result,
.share-lock-card {
  min-height: 280px;
}

.share-lock-card {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 1rem;
  padding: 1.6rem;
}

.share-lock-actions {
  display: flex;
  justify-content: flex-start;
}

.share-content {
  display: grid;
  grid-template-columns: minmax(320px, 380px) minmax(0, 1fr);
  gap: 1rem;
}

.share-browser-panel,
.share-viewer-panel,
.share-single-panel {
  padding: 1.2rem;
}

.share-single-panel {
  min-height: 520px;
  grid-column: 1 / -1;
}

.share-browser-top,
.share-viewer-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.share-breadcrumbs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  margin-bottom: 1rem;
}

.share-crumb {
  border: 0;
  border-radius: 999px;
  padding: 0.5rem 0.8rem;
  background: rgba(255, 255, 255, 0.06);
  color: inherit;
  cursor: pointer;
}

.share-crumb.active {
  background: rgba(248, 196, 113, 0.2);
}

.share-items {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.share-item-row {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  width: 100%;
  padding: 0.9rem 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.03);
  color: inherit;
  cursor: pointer;
  text-align: left;
}

.share-item-row.active,
.share-item-row:hover {
  border-color: rgba(248, 196, 113, 0.34);
  background: rgba(248, 196, 113, 0.09);
}

.share-item-main {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  min-width: 0;
}

.share-item-icon {
  font-size: 1.3rem;
}

.share-item-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.share-item-copy strong,
.share-item-copy small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.share-item-copy small,
.share-item-meta {
  color: rgba(255, 255, 255, 0.62);
}

.share-viewer-actions {
  display: flex;
  gap: 0.6rem;
}

.share-media-stage,
.share-pdf-stage,
.share-text-stage,
.share-office-stage,
.share-audio-stage,
.share-empty-state,
.share-preview-loading {
  min-height: 420px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.03);
}

.share-preview-loading,
.share-empty-state,
.share-office-stage,
.share-audio-stage {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 0.75rem;
  padding: 1.5rem;
  text-align: center;
  color: rgba(255, 255, 255, 0.72);
}

.share-image,
.share-video,
.share-pdf {
  width: 100%;
  height: 100%;
  border: 0;
  border-radius: 22px;
  background: #050913;
}

.share-media-stage {
  overflow: hidden;
}

.share-image {
  object-fit: contain;
}

.share-video {
  object-fit: contain;
}

.share-pdf-stage {
  overflow: hidden;
}

.share-pdf {
  min-height: 620px;
}

.share-text-stage {
  overflow: auto;
  padding: 1rem 1.15rem;
}

.share-text-stage pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'SFMono-Regular', Menlo, Monaco, Consolas, monospace;
  font-size: 0.92rem;
  line-height: 1.65;
  color: #f3f5f9;
}

.share-audio {
  width: min(520px, 100%);
}

@media (max-width: 900px) {
  .share-content {
    grid-template-columns: 1fr;
  }

  .share-hero,
  .share-browser-top,
  .share-viewer-head {
    flex-direction: column;
  }

  .share-hero-meta {
    align-items: flex-start;
  }

  .share-viewer-actions {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>
