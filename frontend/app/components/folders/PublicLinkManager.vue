<template>
  <div class="share-section">
    <div class="share-section-head">
      <div>
        <div class="share-section-title">Public links</div>
        <p class="share-section-copy">Mọi preview và download từ link public sẽ luôn đi qua policy của app trước khi cấp quyền.</p>
      </div>
      <n-button size="small" tertiary :loading="loading" @click="loadLinks">
        Làm mới
      </n-button>
    </div>

    <div class="share-create-card">
      <div class="share-grid">
        <label class="share-field">
          <span class="share-label">Access mode</span>
          <div class="share-mode-row">
            <label class="share-radio active">
              <input type="radio" checked />
              <span>Anonymous + password</span>
            </label>
            <label class="share-radio disabled" title="Sẽ hỗ trợ ở giai đoạn sau">
              <input type="radio" disabled />
              <span>Authenticated external</span>
              <small>Coming soon</small>
            </label>
          </div>
        </label>

        <label class="share-field">
          <span class="share-label">Password protection</span>
          <div class="share-toggle-row">
            <n-switch v-model:value="createPasswordEnabled" />
            <span>{{ createPasswordEnabled ? 'Bật' : 'Tắt' }}</span>
          </div>
          <n-input
            v-if="createPasswordEnabled"
            v-model:value="createPassword"
            type="password"
            show-password-on="click"
            placeholder="Ít nhất 8 ký tự"
          />
        </label>

        <label class="share-field">
          <span class="share-label">Expiry time</span>
          <div class="share-toggle-row">
            <n-switch v-model:value="createExpiryEnabled" />
            <span>{{ createExpiryEnabled ? 'Có hạn' : 'Không hạn' }}</span>
          </div>
          <input
            v-if="createExpiryEnabled"
            v-model="createExpiry"
            class="share-native-input"
            type="datetime-local"
          />
        </label>
      </div>

      <div class="share-create-actions">
        <n-button type="primary" :loading="creating" @click="createLink">
          Tạo public link
        </n-button>
      </div>
    </div>

    <div v-if="links.length === 0 && !loading" class="share-empty">
      Chưa có public link nào cho mục này.
    </div>

    <div v-for="link in links" :key="link.id" class="share-link-card">
      <div class="share-link-head">
        <div class="share-link-status">
          <n-tag :type="statusTagType(link.status)" size="small" round>
            {{ link.status }}
          </n-tag>
          <span class="share-link-meta">Tạo {{ formatDate(link.created_at) }}</span>
        </div>

        <div class="share-link-actions">
          <n-button size="small" tertiary @click="copyLink(link.token)">Copy link</n-button>
          <n-button
            size="small"
            type="error"
            quaternary
            :disabled="link.status === 'revoked'"
            :loading="editing[link.id]?.revoking"
            @click="revokeLink(link)"
          >
            Revoke
          </n-button>
        </div>
      </div>

      <label class="share-field">
        <span class="share-label">URL</span>
        <div class="share-url-row">
          <n-input :value="buildPublicUrl(link.token)" readonly />
          <n-button size="small" secondary @click="openLink(link.token)">Mở</n-button>
        </div>
      </label>

      <div class="share-stats">
        <div class="share-stat">
          <span class="share-stat-label">Access count</span>
          <strong>{{ link.access_count }}</strong>
        </div>
        <div class="share-stat">
          <span class="share-stat-label">Last access</span>
          <strong>{{ formatDate(link.last_accessed_at) }}</strong>
        </div>
        <div class="share-stat">
          <span class="share-stat-label">Password</span>
          <strong>{{ link.requires_password ? 'Enabled' : 'Off' }}</strong>
        </div>
        <div class="share-stat">
          <span class="share-stat-label">Expiry</span>
          <strong>{{ link.expires_at ? formatDate(link.expires_at) : 'Never' }}</strong>
        </div>
      </div>

      <div class="share-grid share-grid--compact">
        <label class="share-field">
          <span class="share-label">Password protection</span>
          <div class="share-toggle-row">
            <n-switch v-model:value="editing[link.id].passwordEnabled" />
            <span>{{ editing[link.id].passwordEnabled ? 'Bật' : 'Tắt' }}</span>
          </div>
          <n-input
            v-if="editing[link.id].passwordEnabled"
            v-model:value="editing[link.id].password"
            type="password"
            show-password-on="click"
            placeholder="Để trống nếu giữ password hiện tại"
          />
        </label>

        <label class="share-field">
          <span class="share-label">Expiry time</span>
          <div class="share-toggle-row">
            <n-switch v-model:value="editing[link.id].expiryEnabled" />
            <span>{{ editing[link.id].expiryEnabled ? 'Có hạn' : 'Không hạn' }}</span>
          </div>
          <input
            v-if="editing[link.id].expiryEnabled"
            v-model="editing[link.id].expiry"
            class="share-native-input"
            type="datetime-local"
          />
        </label>
      </div>

      <div class="share-save-row">
        <n-button size="small" :loading="editing[link.id].saving" @click="saveLink(link)">
          Lưu thay đổi
        </n-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { PublicLink } from '~/types/public-link'

const props = defineProps<{
  itemId: string
}>()

type LinkEditState = {
  passwordEnabled: boolean
  password: string
  expiryEnabled: boolean
  expiry: string
  saving: boolean
  revoking: boolean
}

const api = useApi()
const message = useMessage()
const dialog = useDialog()

const loading = ref(false)
const creating = ref(false)
const links = ref<PublicLink[]>([])
const editing = reactive<Record<string, LinkEditState>>({})

const createPasswordEnabled = ref(false)
const createPassword = ref('')
const createExpiryEnabled = ref(false)
const createExpiry = ref('')

watch(() => props.itemId, () => {
  resetCreateForm()
  loadLinks()
}, { immediate: true })

function resetCreateForm() {
  createPasswordEnabled.value = false
  createPassword.value = ''
  createExpiryEnabled.value = false
  createExpiry.value = ''
}

function syncEditStates() {
  const nextIds = new Set(links.value.map(link => link.id))

  for (const link of links.value) {
    editing[link.id] = {
      passwordEnabled: link.requires_password,
      password: '',
      expiryEnabled: Boolean(link.expires_at),
      expiry: toLocalDateTime(link.expires_at),
      saving: editing[link.id]?.saving ?? false,
      revoking: editing[link.id]?.revoking ?? false,
    }
  }

  for (const id of Object.keys(editing)) {
    if (!nextIds.has(id)) {
      delete editing[id]
    }
  }
}

async function loadLinks() {
  if (!props.itemId) return
  loading.value = true
  try {
    links.value = await api.listItemPublicLinks(props.itemId)
    syncEditStates()
  } catch (error: any) {
    message.error(error?.data?.error || 'Không tải được public links')
  } finally {
    loading.value = false
  }
}

async function createLink() {
  if (createPasswordEnabled.value && createPassword.value.trim().length < 8) {
    message.error('Password phải có ít nhất 8 ký tự')
    return
  }

  const body: Record<string, string> = {}
  if (createPasswordEnabled.value && createPassword.value.trim()) {
    body.password = createPassword.value.trim()
  }
  if (createExpiryEnabled.value && createExpiry.value) {
    body.expires_at = new Date(createExpiry.value).toISOString()
  }

  creating.value = true
  try {
    const created = await api.createItemPublicLink(props.itemId, body)
    links.value = [created, ...links.value]
    syncEditStates()
    resetCreateForm()
    message.success('Đã tạo public link')
  } catch (error: any) {
    message.error(error?.data?.error || 'Không tạo được public link')
  } finally {
    creating.value = false
  }
}

async function saveLink(link: PublicLink) {
  const state = editing[link.id]
  if (!state) return

  if (state.passwordEnabled && !link.requires_password && state.password.trim().length < 8) {
    message.error('Password phải có ít nhất 8 ký tự')
    return
  }

  state.saving = true
  try {
    const payload: Record<string, string | boolean> = {
      password_enabled: state.passwordEnabled,
      clear_expiry: !state.expiryEnabled,
    }

    if (state.passwordEnabled && state.password.trim()) {
      payload.password = state.password.trim()
    }
    if (state.expiryEnabled && state.expiry) {
      payload.expires_at = new Date(state.expiry).toISOString()
      payload.clear_expiry = false
    }

    const updated = await api.updatePublicLink(link.id, payload)
    links.value = links.value.map(candidate => candidate.id === updated.id ? updated : candidate)
    syncEditStates()
    message.success('Đã lưu cấu hình link')
  } catch (error: any) {
    message.error(error?.data?.error || 'Không cập nhật được public link')
  } finally {
    state.saving = false
  }
}

function revokeLink(link: PublicLink) {
  dialog.warning({
    title: 'Revoke public link',
    content: `Thu hồi link cho "${link.token.slice(0, 8)}..." ngay bây giờ? Mọi request mới sẽ bị từ chối.`,
    positiveText: 'Revoke',
    negativeText: 'Hủy',
    onPositiveClick: async () => {
      const state = editing[link.id]
      if (state) state.revoking = true
      try {
        const updated = await api.revokePublicLink(link.id)
        links.value = links.value.map(candidate => candidate.id === updated.id ? updated : candidate)
        syncEditStates()
        message.success('Đã revoke public link')
      } catch (error: any) {
        message.error(error?.data?.error || 'Không revoke được public link')
      } finally {
        if (state) state.revoking = false
      }
    },
  })
}

async function copyLink(token: string) {
  try {
    await navigator.clipboard.writeText(buildPublicUrl(token))
    message.success('Đã copy public link')
  } catch {
    message.error('Không copy được link')
  }
}

function openLink(token: string) {
  window.open(buildPublicUrl(token), '_blank')
}

function buildPublicUrl(token: string) {
  return `${window.location.origin}/share/${token}`
}

function toLocalDateTime(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const offset = date.getTimezoneOffset()
  const local = new Date(date.getTime() - offset * 60_000)
  return local.toISOString().slice(0, 16)
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

function statusTagType(status: PublicLink['status']) {
  switch (status) {
    case 'active':
      return 'success'
    case 'expired':
      return 'warning'
    case 'revoked':
      return 'error'
    default:
      return 'default'
  }
}
</script>

<style scoped>
.share-section {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
}

.share-section-head {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
}

.share-section-title {
  font-size: var(--font-size-sm);
  font-weight: 600;
  color: var(--color-text-primary);
}

.share-section-copy {
  margin: 0.25rem 0 0;
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  line-height: 1.5;
}

.share-create-card,
.share-link-card {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  padding: 0.95rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.02);
}

.share-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.share-grid--compact {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.share-field {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.share-label {
  font-size: var(--font-size-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.share-toggle-row,
.share-mode-row,
.share-url-row,
.share-create-actions,
.share-save-row,
.share-link-head,
.share-link-status,
.share-link-actions {
  display: flex;
  gap: 0.65rem;
  align-items: center;
}

.share-mode-row {
  flex-direction: column;
  align-items: stretch;
}

.share-radio {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.7rem 0.8rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: var(--font-size-sm);
}

.share-radio.active {
  border-color: rgba(24, 160, 88, 0.35);
  background: rgba(24, 160, 88, 0.08);
}

.share-radio.disabled {
  opacity: 0.6;
}

.share-radio small {
  margin-left: auto;
  color: var(--color-text-muted);
}

.share-link-head {
  justify-content: space-between;
}

.share-link-meta {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
}

.share-url-row {
  align-items: stretch;
}

.share-url-row :deep(.n-input) {
  flex: 1;
}

.share-native-input {
  width: 100%;
  min-height: 40px;
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.04);
  color: var(--color-text-primary);
}

.share-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.7rem;
}

.share-stat {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.75rem;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.03);
}

.share-stat-label {
  font-size: var(--font-size-xs);
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.share-empty {
  padding: 0.95rem;
  border: 1px dashed var(--color-border);
  border-radius: var(--radius-lg);
  color: var(--color-text-muted);
  text-align: center;
}

@media (max-width: 720px) {
  .share-grid,
  .share-grid--compact,
  .share-stats {
    grid-template-columns: 1fr;
  }

  .share-link-head,
  .share-link-status,
  .share-link-actions,
  .share-url-row,
  .share-section-head {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
