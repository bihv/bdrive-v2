<template>
  <div class="office-page">
    <!-- OnlyOffice Editor Container -->
    <div class="office-content">
      <n-spin v-if="loading" size="large" class="office-loading">
        <template #description>
          Đang khởi tạo trình soạn thảo...
        </template>
      </n-spin>

      <div v-if="error" class="office-error">
        <n-result status="error" :title="error" description="Không thể mở trình soạn thảo Office">
          <template #footer>
            <n-button @click="goBack">Quay lại</n-button>
          </template>
        </n-result>
      </div>

      <div
        id="onlyoffice-editor"
        v-show="!loading && !error"
        class="office-editor"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import type { PreviewData } from '~/composables/usePreview'
import { useAuthStore } from '~/stores/auth'

definePageMeta({
  layout: false,
  middleware: 'auth',
})

const route = useRoute()
const router = useRouter()
const config = useRuntimeConfig()
const message = useMessage()
const authStore = useAuthStore()
const { getPreviewData, getOfficeDocumentType, getFileExtension, getOnlyOfficeUrl } = usePreview()

const loading = ref(true)
const error = ref<string | null>(null)
const previewData = ref<PreviewData | null>(null)
let docEditor: any = null

const fileIcon = computed(() => {
  if (!previewData.value) return 'mdi:file-document-outline'
  const ext = getFileExtension(previewData.value.name)
  if (['doc', 'docx'].includes(ext)) return 'mdi:file-word-box'
  if (['xls', 'xlsx'].includes(ext)) return 'mdi:file-excel-box'
  if (['ppt', 'pptx'].includes(ext)) return 'mdi:file-powerpoint-box'
  return 'mdi:file-document-outline'
})

const iconColor = computed(() => {
  const ext = previewData.value ? getFileExtension(previewData.value.name) : ''
  if (['doc', 'docx'].includes(ext)) return '#2b579a'
  if (['xls', 'xlsx'].includes(ext)) return '#217346'
  if (['ppt', 'pptx'].includes(ext)) return '#d24726'
  return 'var(--color-primary)'
})

useHead({
  title: computed(() => previewData.value?.name ? `${previewData.value.name} - 1Drive` : '1Drive - Office'),
})

function goBack() {
  if (docEditor) {
    try { docEditor.destroyEditor() } catch (_) {}
  }
  window.close()
  // Fallback if window.close() doesn't work (not opened by script)
  router.back()
}

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    // Check if already loaded
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve()
      return
    }

    const script = document.createElement('script')
    script.src = src
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error(`Failed to load OnlyOffice script from ${src}`))
    document.head.appendChild(script)
  })
}

function initEditor(data: PreviewData, onlyofficeUrl: string) {
  const ext = getFileExtension(data.name)
  const documentType = getOfficeDocumentType(data.name)
  const apiBase = config.public.apiBase

  // Fix callback URL for Docker (OnlyOffice inside Docker needs to reach host machine)
  let callbackUrlStr = `${apiBase}/api/v1/onlyoffice/callback?id=${route.query.id}&userId=${authStore.user?.id}`
  if (callbackUrlStr.includes('localhost') || callbackUrlStr.includes('127.0.0.1')) {
    callbackUrlStr = callbackUrlStr.replace('localhost', 'host.docker.internal').replace('127.0.0.1', 'host.docker.internal')
  }

  const editorConfig = {
    document: {
      fileType: ext,
      key: `${route.query.id}_${new Date(data.updated_at).getTime()}`,
      title: data.name,
      url: data.url,
      permissions: {
        edit: true,
        download: true,
      }
    },
    documentType: documentType,
    editorConfig: {
      mode: 'edit',
      callbackUrl: callbackUrlStr,
      lang: 'en',
      customization: {
        forcesave: true,
        chat: false,
        comments: false,
        compactHeader: true,
        toolbarNoTabs: true,
        hideRightMenu: true,
        uiTheme: 'theme-dark',
      },
      user: {
        id: authStore.user?.id || 'anonymous',
        name: authStore.user?.full_name || 'Anonymous',
      }
    },
    type: 'desktop',
    height: '100%',
    width: '100%',
  }

  try {
    // @ts-ignore - DocsAPI is loaded dynamically
    docEditor = new DocsAPI.DocEditor('onlyoffice-editor', editorConfig)
  } catch (e: any) {
    console.error('Failed to initialize OnlyOffice editor:', e)
    error.value = 'Không thể khởi tạo trình soạn thảo'
  }
}

onMounted(async () => {
  const itemId = route.query.id as string
  if (!itemId) {
    error.value = 'Thiếu ID file'
    loading.value = false
    return
  }

  try {
    // Step 1: Get preview data
    const data = await getPreviewData(itemId)
    previewData.value = data

    // Step 2: Load OnlyOffice JS API
    const onlyofficeUrl = getOnlyOfficeUrl()
    await loadScript(`${onlyofficeUrl}/web-apps/apps/api/documents/api.js`)

    // Step 3: Initialize editor
    loading.value = false
    await nextTick()
    initEditor(data, onlyofficeUrl)
  } catch (e: any) {
    error.value = e?.data?.error || e?.message || 'Không thể tải trình soạn thảo Office'
    message.error(error.value!)
    loading.value = false
  }
})

onUnmounted(() => {
  if (docEditor) {
    try { docEditor.destroyEditor() } catch (_) {}
  }
})
</script>

<style scoped>
.office-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  margin: 0;
  padding: 0;
  overflow: hidden;
}

.office-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 0;
  position: relative;
}

.office-loading {
  margin: auto;
}

.office-error {
  padding: 2rem;
}

.office-editor {
  width: 100%;
  height: 100%;
}

/* Mobile */
@media (max-width: 768px) {
  .office-page {
    height: 100vh;
  }
}
</style>
