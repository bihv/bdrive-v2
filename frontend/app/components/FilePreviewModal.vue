<template>
  <Transition name="fade">
    <div v-show="previewItemId && previewType !== 'image'" class="preview-modal-overlay">
      <div class="preview-modal-container">
        <!-- Header -->
        <div v-show="previewType !== 'pdf'" class="preview-header">
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
              v-if="(previewType === 'text' || previewType === 'markdown') && isContentModified"
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
            <img
              :src="previewData?.url"
              :alt="previewData?.name"
              class="preview-image clickable-image"
              @click="openLightbox"
              title="Click to zoom"
            />
          </div>

          <!-- Video Preview -->
          <div v-else-if="previewType === 'video'" class="preview-media-wrapper">
            <video
              ref="videoRef"
              class="preview-video"
              :src="previewData?.url"
              crossorigin="anonymous"
              playsinline
            ></video>
          </div>

          <!-- Audio Preview -->
          <div v-else-if="previewType === 'audio'" class="preview-audio-container">
            <div class="audio-visual">
              <n-icon size="100" :style="{ color: 'var(--color-primary)' }">
                <Icon icon="mdi:music-circle" />
              </n-icon>
              <p class="audio-filename">{{ previewData?.name }}</p>
            </div>
            <div class="waveform-wrapper">
              <div ref="audioContainerRef"></div>
              <div class="audio-controls">
                <n-button circle @click="toggleAudio" type="primary" size="large">
                  <template #icon><Icon :icon="isAudioPlaying ? 'mdi:pause' : 'mdi:play'" /></template>
                </n-button>
              </div>
            </div>
          </div>

          <div v-else-if="previewType === 'pdf'" class="preview-pdf-container">
            <iframe
              :src="previewData?.url ? previewData.url : ''"
              class="preview-pdf"
              title="PDF Preview"
            ></iframe>
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

          <!-- Markdown Editor -->
          <div v-else-if="previewType === 'markdown'" class="preview-editor-container markdown-bg">
            <n-spin v-if="textLoading" size="medium" />
            <div v-else ref="markdownEditorRef" class="toast-editor-wrapper"></div>
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

        <!-- Floating Close Button for PDF -->
        <div v-if="previewType === 'pdf'" class="floating-close-container">
          <n-button round type="primary" class="pdf-close-btn" size="large" @click="closePreview">
            <template #icon>
              <n-icon size="20" color="#fff"><Icon icon="mdi:close" /></n-icon>
            </template>
            <span style="color: #fff; margin-left: 4px;">Close</span>
          </n-button>
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
const { getPreviewData, getPreviewType, getFileExtension, getMonacoLanguage, previewItemId, previewContext, previewForceType } = usePreview()

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

// DOM Refs
const videoRef = ref<HTMLVideoElement | null>(null)
const audioContainerRef = ref<HTMLElement | null>(null)
const markdownEditorRef = ref<HTMLElement | null>(null)
const isAudioPlaying = ref(false)

// Instance Refs
let plyrInstance: any = null
let wavesurferInstance: any = null
let toastEditorInstance: any = null
let lightboxInstance: any = null

function cleanupInstances() {
  if (plyrInstance) {
    plyrInstance.destroy()
    plyrInstance = null
  }
  if (wavesurferInstance) {
    wavesurferInstance.destroy()
    wavesurferInstance = null
  }
  if (toastEditorInstance) {
    toastEditorInstance.destroy()
    toastEditorInstance = null
  }
  if (lightboxInstance) {
    lightboxInstance.destroy()
    lightboxInstance = null
  }
  isAudioPlaying.value = false
}

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
  previewForceType.value = null
  cleanupInstances()
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

async function initPlyr() {
  if (!videoRef.value) return
  const Plyr = (await import('plyr')).default
  await import('plyr/dist/plyr.css')
  plyrInstance = new Plyr(videoRef.value)
}

async function initWavesurfer(url: string) {
  if (!audioContainerRef.value) return
  const WaveSurfer = (await import('wavesurfer.js')).default
  wavesurferInstance = WaveSurfer.create({
    container: audioContainerRef.value,
    waveColor: '#4ade80',
    progressColor: '#22c55e',
    cursorColor: '#fff',
    barWidth: 2,
    barRadius: 2,
    height: 60,
  })
  wavesurferInstance.load(url)
  wavesurferInstance.on('play', () => isAudioPlaying.value = true)
  wavesurferInstance.on('pause', () => isAudioPlaying.value = false)
}

function toggleAudio() {
  if (wavesurferInstance) {
    wavesurferInstance.playPause()
  }
}

async function initMarkdownEditor() {
  if (!markdownEditorRef.value) return
  // @ts-ignore
  const Editor = (await import('@toast-ui/editor')).default
  await import('@toast-ui/editor/dist/toastui-editor.css')
  
  toastEditorInstance = new Editor({
    el: markdownEditorRef.value,
    height: '100%',
    initialEditType: 'markdown',
    previewStyle: 'vertical',
    initialValue: textContent.value,
    events: {
      change: () => {
        if (toastEditorInstance) {
          textContent.value = toastEditorInstance.getMarkdown()
        }
      }
    }
  })
}

function openLightbox() {
  if (!previewData.value?.url) return

  const imageItems = previewContext.value.filter(item => getPreviewType(item.name) === 'image')
  
  if (imageItems.length === 0) {
    imageItems.push({ id: previewItemId.value!, name: previewData.value.name })
  }

  if (!imageItems.find(i => i.id === previewItemId.value)) {
    imageItems.unshift({ id: previewItemId.value!, name: previewData.value.name })
  }

  const startIndex = Math.max(0, imageItems.findIndex(i => i.id === previewItemId.value))

  const firstImg = new Image()
  firstImg.onload = () => {
    const dataSource = imageItems.map((item: any) => {
      const isCurrent = item.id === previewItemId.value
      return {
        id: item.id,
        name: item.name,
        type: 'image',
        width: isCurrent ? firstImg.naturalWidth : 0,
        height: isCurrent ? firstImg.naturalHeight : 0,
        src: isCurrent && previewData.value?.url ? previewData.value.url : ''
      }
    })

    import('photoswipe/lightbox').then(({ default: PhotoSwipeLightbox }) => {
      import('photoswipe/style.css').then(() => {
        const lightbox = new PhotoSwipeLightbox({
          dataSource,
          pswpModule: () => import('photoswipe'),
          showHideAnimationType: 'fade',
          preload: [1, 1]
        })

        lightbox.on('contentLoad', (e) => {
          const { content } = e
          if (content.type === 'image' && !content.data.src) {
            e.preventDefault()
            content.state = 'loading'
            
            getPreviewData(content.data.id).then(data => {
              const img = new Image()
              img.onload = () => {
                content.data.src = data.url
                content.data.width = img.naturalWidth
                content.data.height = img.naturalHeight
                content.width = img.naturalWidth
                content.height = img.naturalHeight
                
                content.element = img
                content.element.className = 'pswp__img'
                content.state = 'loaded'
                
                // CRITICAL FOR LAZY: Appends to DOM and computes scale!
                // @ts-ignore
                if (typeof content.updatePosition === 'function') content.updatePosition()
                content.onLoaded()

                if (lightbox.pswp) {
                  lightbox.pswp.refreshSlideContent(content.index)
                }
              }
              img.onerror = () => {
                content.state = 'error'
                content.onError()
              }
              img.src = data.url
            }).catch(() => {
              content.state = 'error'
              content.onError()
            })
          }
        })
        
        lightbox.on('destroy', () => {
          previewItemId.value = null
          lightboxInstance = null
        })
        
        lightbox.init()
        lightbox.loadAndOpen(startIndex)
        lightboxInstance = lightbox
      })
    })
  }
  firstImg.src = previewData.value.url
}

watch(previewItemId, async (newId) => {
  if (!newId) return

  loading.value = true
  error.value = null
  cleanupInstances()

  try {
    const data = await getPreviewData(newId)
    previewData.value = data
    previewType.value = previewForceType.value || getPreviewType(data.name)
    previewExtension.value = getFileExtension(data.name)
    monacoLanguage.value = getMonacoLanguage(data.name)

    if (previewType.value === 'text' || previewType.value === 'markdown') {
      await loadTextContent(data.url)
    }

    // Force DOM update so that container refs become available
    loading.value = false
    await nextTick()

    if (previewType.value === 'markdown') {
      initMarkdownEditor()
    } else if (previewType.value === 'video') {
      initPlyr()
    } else if (previewType.value === 'audio') {
      initWavesurfer(data.url)
    } else if (previewType.value === 'image') {
      if (!lightboxInstance) {
        openLightbox()
      }
    }
  } catch (e: any) {
    error.value = e?.data?.error || 'Failed to load preview'
    message.error(error.value!)
    loading.value = false
  }
})

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
  background: white;
}

.preview-pdf {
  width: 100%;
  height: 100%;
  border: none;
}

.floating-close-container {
  position: absolute;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
}

.pdf-close-btn {
  background-color: rgba(0, 0, 0, 0.4) !important;
  color: #fff !important;
  border: none !important;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3) !important;
  font-weight: bold !important;
  transition: all 0.3s ease !important;
}

.pdf-close-btn:hover {
  background-color: rgba(0, 0, 0, 0.9) !important;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(0,0,0,0.4) !important;
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

/* Monaco on iOS: floating keyboard button (editor.contrib.iPadShowKeyboard) */
.preview-editor-container :deep(textarea.iPadShowKeyboard) {
  display: none !important;
}

/* Unknown */
.preview-unknown {
  padding: 2rem;
  background: var(--color-bg-primary);
  border-radius: var(--radius-lg);
}

.clickable-image {
  cursor: pointer;
  transition: transform 0.2s;
}
.clickable-image:hover {
  transform: scale(1.02);
}

.pdf-wrapper {
  overflow-y: auto;
  height: 100%;
  width: 100%;
}

.toast-editor-wrapper {
  height: 100%;
  width: 100%;
  overflow: hidden;
}

.markdown-bg {
  background-color: white !important;
  color: #333;
}

.waveform-wrapper {
  width: 100%;
  max-width: 800px;
  background: rgba(0, 0, 0, 0.4);
  padding: 1.5rem;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-top: 1rem;
}

.audio-controls {
  display: flex;
  justify-content: center;
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
