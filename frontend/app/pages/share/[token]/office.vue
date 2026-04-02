<template>
  <div class="office-page">
    <div class="office-content">
      <n-spin v-if="loading" size="large" class="office-loading">
        <template #description>
          Đang khởi tạo OnlyOffice viewer...
        </template>
      </n-spin>

      <div v-if="error" class="office-error">
        <n-result status="error" :title="error" description="Không thể mở Office viewer">
          <template #footer>
            <n-button @click="goBack">Quay lại</n-button>
          </template>
        </n-result>
      </div>

      <div
        id="public-onlyoffice-editor"
        v-show="!loading && !error"
        class="office-editor"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { usePublicLinkApi } from '~/composables/usePublicLinkApi'

definePageMeta({
  layout: false,
})

const route = useRoute()
const router = useRouter()
const { getPublicLink } = usePublicLinkApi()
const { getOfficeDocumentType, getFileExtension, getOnlyOfficeUrl } = usePreview()

const loading = ref(true)
const error = ref<string | null>(null)
let docEditor: any = null

useHead({
  title: 'OnlyOffice Viewer - 1Drive Shared Link',
})

function goBack() {
  if (docEditor) {
    try { docEditor.destroyEditor() } catch (_) {}
  }
  window.close()
  router.back()
}

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
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

function toDocumentUrl(token: string, itemId: string, session?: string) {
  const publicApi = usePublicLinkApi()
  let streamUrl = publicApi.buildStreamUrl(token, itemId, session, false)
  if (streamUrl.includes('localhost') || streamUrl.includes('127.0.0.1')) {
    streamUrl = streamUrl.replace('localhost', 'host.docker.internal').replace('127.0.0.1', 'host.docker.internal')
  }
  return streamUrl
}

function initEditor(fileName: string, documentUrl: string) {
  const ext = getFileExtension(fileName)
  const documentType = getOfficeDocumentType(fileName)

  const editorConfig = {
    document: {
      fileType: ext,
      key: `${route.params.token}_${route.query.item}_${Date.now()}`,
      title: fileName,
      url: documentUrl,
      permissions: {
        edit: false,
        download: true,
      },
    },
    documentType,
    editorConfig: {
      mode: 'view',
      lang: 'en',
      customization: {
        compactHeader: false,
        toolbarNoTabs: false,
        hideRightMenu: false,
        uiTheme: 'theme-dark',
      },
      user: {
        id: 'anonymous',
        name: 'Anonymous viewer',
      },
    },
    type: 'desktop',
    height: '100%',
    width: '100%',
  }

  try {
    // @ts-ignore
    docEditor = new DocsAPI.DocEditor('public-onlyoffice-editor', editorConfig)
  } catch (cause: any) {
    console.error('Failed to initialize public OnlyOffice editor:', cause)
    error.value = 'Không thể khởi tạo OnlyOffice viewer'
  }
}

onMounted(async () => {
  const token = String(route.params.token || '')
  const itemId = String(route.query.item || '')
  const session = String(route.query.session || '')

  if (!token || !itemId) {
    error.value = 'Thiếu thông tin file'
    loading.value = false
    return
  }

  try {
    const detail = await getPublicLink(token, session || undefined)
    const fileName = String(route.query.name || '') || (detail.item?.id === itemId ? detail.item.name : `shared-file-${itemId}`)

    await loadScript(`${getOnlyOfficeUrl()}/web-apps/apps/api/documents/api.js`)

    loading.value = false
    await nextTick()
    initEditor(fileName, toDocumentUrl(token, itemId, session || undefined))
  } catch (cause: any) {
    error.value = cause?.data?.error || cause?.message || 'Không tải được Office viewer'
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
  width: 100vw;
  height: 100vh;
  background: linear-gradient(180deg, #0a1220 0%, #060b13 100%);
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
</style>
