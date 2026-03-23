<template>
  <Transition name="fade">
    <div v-if="previewItemId" class="preview-modal-overlay">
      <div class="preview-modal-container">
        <!-- Header -->
        <div class="preview-header">
          <div class="preview-header-left">
            <n-button quaternary circle @click="closePreview" class="close-btn">
              <template #icon>
                <n-icon size="24"><Icon icon="mdi:arrow-left" /></n-icon>
              </template>
            </n-button>
            <div class="preview-file-info">
              <n-icon size="24" :style="{ color: 'var(--color-primary)' }">
                <Icon :icon="fileIcon" />
              </n-icon>
              <span class="preview-file-name" :title="previewData?.name">{{ previewData?.name || 'Loading...' }}</span>
            </div>
          </div>
          <div class="preview-header-right">
            <!-- Save button for text editor -->
            <n-button
              v-if="previewType === 'text' && isContentModified"
              type="primary"
              size="small"
              :loading="saving"
              @click="saveTextContent"
              class="save-btn"
            >
              <template #icon>
                <n-icon><Icon icon="mdi:content-save" /></n-icon>
              </template>
              Save
            </n-button>
            <n-button quaternary circle @click="downloadFile" class="action-btn" title="Download">
              <template #icon>
                <n-icon size="24"><Icon icon="mdi:download" /></n-icon>
              </template>
            </n-button>
            <n-button quaternary circle @click="closePreview" class="action-btn" title="Close">
              <template #icon>
                <n-icon size="24"><Icon icon="mdi:close" /></n-icon>
              </template>
            </n-button>
          </div>
        </div>

        <!-- Content -->
        <div class="preview-content">
          <n-spin v-if="loading" size="large" class="preview-loading" />

          <div v-else-if="error" class="preview-error">
            <n-result status="error" :title="error" description="Cannot preview this file">
              <template #footer>
                <n-button @click="closePreview">Close</n-button>
              </template>
            </n-result>
          </div>

          <!-- Image Preview -->
          <div v-else-if="previewType === 'image'" class="preview-media-wrapper">
            <n-image
              :src="previewData?.url"
              :alt="previewData?.name"
              class="preview-image"
              object-fit="contain"
            />
          </div>

          <!-- Video Preview -->
          <div v-else-if="previewType === 'video'" class="preview-media-wrapper">
            <video
              controls
              autoplay
              class="preview-video"
              :src="previewData?.url"
            >
              Your browser does not support video playback.
            </video>
          </div>

          <!-- Audio Preview -->
          <div v-else-if="previewType === 'audio'" class="preview-audio-container">
            <div class="audio-visual">
              <n-icon size="100" :style="{ color: 'var(--color-primary)' }">
                <Icon icon="mdi:music-circle" />
              </n-icon>
              <p class="audio-filename">{{ previewData?.name }}</p>
            </div>
            <audio controls autoplay class="preview-audio" :src="previewData?.url">
              Your browser does not support audio playback.
            </audio>
          </div>

          <!-- PDF Preview -->
          <div v-else-if="previewType === 'pdf'" class="preview-pdf-container">
            <iframe :src="previewData?.url" class="preview-pdf" />
          </div>

          <!-- Text/Code Editor (Monaco) -->
          <div v-else-if="previewType === 'text'" class="preview-editor-container">
            <n-spin v-if="textLoading" size="medium" />
            <MonacoEditor
              v-else
              v-model="textContent"
              :lang="monacoLanguage"
              theme="vs-dark"
              :options="monacoOptions"
              class="monaco-editor-instance"
              @load="handleEditorMount"
            />
          </div>

          <!-- Unknown File Type -->
          <div v-else class="preview-unknown">
            <n-result
              status="info"
              title="Preview not supported"
              :description="`File type ${previewData?.mime_type || 'unknown'} is not supported for preview.`"
            >
              <template #footer>
                <n-space>
                  <n-button @click="closePreview">Close</n-button>
                  <n-button type="primary" @click="downloadFile">Download</n-button>
                </n-space>
              </template>
            </n-result>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { PreviewData, PreviewType } from '~/composables/usePreview'

const message = useMessage()
const api = useApi()
const { getPreviewData, getPreviewType, getFileExtension, getMonacoLanguage, previewItemId } = usePreview()

const loading = ref(true)
const error = ref<string | null>(null)
const previewData = ref<PreviewData | null>(null)
const previewType = ref<PreviewType>('unknown')
const previewExtension = ref<string>('txt')
const textContent = ref('')
const textLoading = ref(false)
const monacoLanguage = ref('plaintext')
const saving = ref(false)
const originalContent = ref('')

const isContentModified = computed(() => textContent.value !== originalContent.value)

const monacoOptions = {
  automaticLayout: true,
  fontSize: 14,
  fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace",
  minimap: { enabled: true },
  scrollBeyondLastLine: false,
  wordWrap: 'on' as const,
  padding: { top: 16 },
  renderLineHighlight: 'all' as const,
  smoothScrolling: true,
  cursorBlinking: 'smooth' as const,
  cursorSmoothCaretAnimation: 'on' as const,
}

function handleEditorMount(editor: any) {
  // Register Ctrl+S / Cmd+S shortcut to save
  editor.addCommand(
    // Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.KeyS
    2048 | 49, // CtrlCmd = 2048, KeyS = 49
    () => {
      if (isContentModified.value && !saving.value) {
        saveTextContent()
      }
    }
  )
}

const fileIcon = computed(() => {
  switch (previewType.value) {
    case 'image': return 'mdi:file-image'
    case 'video': return 'mdi:file-video'
    case 'audio': return 'mdi:file-music'
    case 'pdf': return 'mdi:file-pdf-box'
    case 'text': return 'mdi:file-code'
    default: return 'mdi:file-document-outline'
  }
})

function closePreview() {
  if (isContentModified.value) {
    if (!confirm('You have unsaved changes. Close anyway?')) return
  }
  previewItemId.value = null
  // Reset states
  setTimeout(() => {
    loading.value = true
    error.value = null
    previewData.value = null
    previewType.value = 'unknown'
    textContent.value = ''
    originalContent.value = ''
  }, 300)
}

function downloadFile() {
  if (!previewData.value?.url) return
  const a = document.createElement('a')
  a.href = previewData.value.url
  a.download = previewData.value.name
  a.target = '_blank'
  a.click()
}

async function loadTextContent(url: string) {
  textLoading.value = true
  try {
    const response = await fetch(url)
    const text = await response.text()
    // Limit display to 1MB
    textContent.value = text.length > 1048576 ? text.substring(0, 1048576) + '\n\n... (File too large, truncated)' : text
    originalContent.value = textContent.value
  } catch (e) {
    textContent.value = 'Failed to load file content.'
    originalContent.value = textContent.value
  } finally {
    textLoading.value = false
  }
}

async function saveTextContent() {
  if (!previewItemId.value) return
  saving.value = true
  try {
    const blob = new Blob([textContent.value], { type: 'text/plain' })
    const formData = new FormData()
    formData.append('file', blob, previewData.value?.name || 'file.txt')

    await api.put(`/api/v1/items/${previewItemId.value}/content`, formData)

    originalContent.value = textContent.value
    message.success('Changes saved successfully!')
  } catch (e: any) {
    message.error(e?.data?.error || e.message || 'Failed to save file')
  } finally {
    saving.value = false
  }
}

watch(previewItemId, async (newId) => {
  if (!newId) return

  loading.value = true
  error.value = null

  try {
    const data = await getPreviewData(newId)
    previewData.value = data
    previewType.value = getPreviewType(data.name)
    previewExtension.value = getFileExtension(data.name)
    monacoLanguage.value = getMonacoLanguage(data.name)

    if (previewType.value === 'text') {
      await loadTextContent(data.url)
    }
  } catch (e: any) {
    error.value = e?.data?.error || 'Failed to load preview'
    message.error(error.value!)
  } finally {
    loading.value = false
  }
}, { immediate: true })

</script>

<style scoped>
.preview-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background-color: rgba(15, 15, 15, 0.95);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-modal-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  max-width: 100vw;
  max-height: 100vh;
}

.preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1.5rem;
  background: linear-gradient(to bottom, rgba(0,0,0,0.6) 0%, transparent 100%);
  color: white;
  z-index: 10;
}

.preview-header-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  min-width: 0;
}

.close-btn {
  color: white !important;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1) !important;
}

.action-btn {
  color: white !important;
  margin-left: 0.5rem;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1) !important;
}

.save-btn {
  margin-right: 0.5rem;
}

.preview-file-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
}

.preview-file-name {
  font-weight: 500;
  font-size: 1.1rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 60vw;
}

.preview-header-right {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.preview-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 0;
  position: relative;
  overflow: hidden;
  padding: 0;
}

.preview-loading {
  margin: auto;
}

.preview-error {
  padding: 2rem;
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
}

/* Media (Image & Video) */
.preview-media-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  padding: 1rem;
}

:deep(.preview-image) {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  filter: drop-shadow(0 0 20px rgba(0,0,0,0.5));
}

.preview-video {
  max-width: 100%;
  max-height: 100%;
  filter: drop-shadow(0 0 20px rgba(0,0,0,0.5));
  outline: none;
}

/* Audio */
.preview-audio-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2rem;
  width: 100%;
  height: 100%;
  padding: 1rem;
}

.audio-visual {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.audio-filename {
  font-size: 1.2rem;
  font-weight: 500;
  color: white;
  text-align: center;
  max-width: 80vw;
  word-wrap: break-word;
}

.preview-audio {
  width: 100%;
  max-width: 600px;
}

/* PDF */
.preview-pdf-container {
  width: 100%;
  height: 100%;
  max-width: 1200px;
  background: white;
  box-shadow: 0 0 30px rgba(0,0,0,0.3);
}

.preview-pdf {
  width: 100%;
  height: 100%;
  border: none;
}

/* Monaco Editor */
.preview-editor-container {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.monaco-editor-instance {
  width: 100%;
  height: 100%;
  min-height: 0;
}

/* Unknown */
.preview-unknown {
  padding: 2rem;
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .preview-header {
    padding: 0.75rem 1rem;
  }
  
  .preview-header-left {
    gap: 0.75rem;
  }
  
  .preview-file-name {
    max-width: 45vw;
    font-size: 1rem;
  }
}
</style>
