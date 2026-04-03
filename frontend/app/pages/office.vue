<template>
  <div :class="['office-page', { 'office-page--shared': isPublicShare }]">
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
import type { PreviewData } from '~/composables/usePreview'
import { usePublicLinkApi } from '~/composables/usePublicLinkApi'
import { useAuthStore } from '~/stores/auth'

definePageMeta({
  layout: false,
  middleware: 'office-access',
})

const route = useRoute()
const router = useRouter()
const config = useRuntimeConfig()
const message = useMessage()
const authStore = useAuthStore()
const publicApi = usePublicLinkApi()
const { getPreviewData, getOfficeDocumentType, getFileExtension, getOnlyOfficeUrl } = usePreview()

const loading = ref(true)
const error = ref<string | null>(null)
const previewData = ref<PreviewData | null>(null)
let docEditor: any = null

const itemId = computed(() => getSingleQueryValue(route.query.id))
const shareToken = computed(() => getSingleQueryValue(route.query.share_token))
const shareSession = computed(() => getSingleQueryValue(route.query.session))
const shareName = computed(() => getSingleQueryValue(route.query.name))
const isPublicShare = computed(() => Boolean(shareToken.value))

useHead({
  title: computed(() => {
    if (previewData.value?.name) return `${previewData.value.name} - 1Drive`
    return isPublicShare.value ? '1Drive Shared Office' : '1Drive - Office'
  }),
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

function getSingleQueryValue(value: string | string[] | undefined): string {
  if (Array.isArray(value)) {
    return String(value[0] || '')
  }
  return String(value || '')
}

function normalizeDockerReachableUrl(url: string): string {
  if (url.includes('localhost') || url.includes('127.0.0.1')) {
    return url.replace('localhost', 'host.docker.internal').replace('127.0.0.1', 'host.docker.internal')
  }
  return url
}

function buildInternalCallbackUrl() {
  const apiBase = config.public.apiBase
  return normalizeDockerReachableUrl(
    `${apiBase}/api/v1/onlyoffice/callback?id=${itemId.value}&userId=${authStore.user?.id}`,
  )
}

function initEditor(data: PreviewData, options: { isPublicShare: boolean; shareToken?: string }) {
  const ext = getFileExtension(data.name)
  const documentType = getOfficeDocumentType(data.name)
  const documentKey = options.isPublicShare
    ? `${options.shareToken || 'public'}_${itemId.value}_${Date.now()}`
    : `${itemId.value}_${new Date(data.updated_at).getTime()}`

  const editorConfig = {
    document: {
      fileType: ext,
      key: documentKey,
      title: data.name,
      url: data.url,
      permissions: {
        edit: !options.isPublicShare,
        download: true,
      },
    },
    documentType: documentType,
    editorConfig: {
      mode: options.isPublicShare ? 'view' : 'edit',
      callbackUrl: options.isPublicShare ? undefined : buildInternalCallbackUrl(),
      lang: 'en',
      customization: {
        forcesave: !options.isPublicShare,
        chat: false,
        comments: false,
        compactHeader: false,
        toolbarNoTabs: false,
        hideRightMenu: false,
        uiTheme: 'theme-dark',
      },
      user: {
        id: options.isPublicShare ? 'anonymous' : (authStore.user?.id || 'anonymous'),
        name: options.isPublicShare ? 'Anonymous viewer' : (authStore.user?.full_name || 'Anonymous'),
      },
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

async function loadInternalPreviewData() {
  const data = await getPreviewData(itemId.value)
  previewData.value = data
  return {
    data,
    options: { isPublicShare: false },
  }
}

async function loadPublicPreviewData() {
  const detail = await publicApi.getPublicLink(shareToken.value, shareSession.value || undefined)
  const name = shareName.value || (detail.item?.id === itemId.value ? detail.item.name : `shared-file-${itemId.value}`)
  const updatedAt = detail.item?.id === itemId.value ? detail.item.updated_at : new Date().toISOString()
  const url = normalizeDockerReachableUrl(
    publicApi.buildStreamUrl(shareToken.value, itemId.value, shareSession.value || undefined, false),
  )

  const data: PreviewData = {
    url,
    mime_type: detail.item?.id === itemId.value ? detail.item.mime_type || '' : '',
    name,
    size: detail.item?.id === itemId.value ? detail.item.size : 0,
    updated_at: updatedAt,
  }

  previewData.value = data

  return {
    data,
    options: { isPublicShare: true, shareToken: shareToken.value },
  }
}

onMounted(async () => {
  if (!itemId.value) {
    error.value = 'Thiếu ID file'
    loading.value = false
    return
  }

  try {
    const officeUrl = getOnlyOfficeUrl()
    const previewSource = isPublicShare.value
      ? await loadPublicPreviewData()
      : await loadInternalPreviewData()

    await loadScript(`${officeUrl}/web-apps/apps/api/documents/api.js`)

    loading.value = false
    await nextTick()
    initEditor(previewSource.data, previewSource.options)
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

.office-page--shared {
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

/* Mobile */
@media (max-width: 768px) {
  .office-page {
    height: 100vh;
  }
}
</style>
