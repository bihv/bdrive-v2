<template>
  <div class="settings-page">
    <SettingsHero
      :used-storage-text="usedStorageText"
      :storage-usage-percent="storageUsagePercent"
      :role-label="roleLabel"
      :is-verified="Boolean(currentUser?.is_verified)"
    />

    <section class="settings-grid">
      <SettingsProfilePanel
        :initials="initials"
        :full-name="currentUser?.full_name || 'Người dùng 1Drive'"
        :email="currentUser?.email || 'no-email@example.com'"
        :created-at-text="createdAtText"
        :last-login-text="lastLoginText"
      />

      <SettingsStoragePanel
        :draft="draft"
        :quota-text="quotaText"
        :storage-usage-percent="storageUsagePercent"
        :retention-options="retentionOptions"
        @update:storage-alerts="draft.storageAlerts = $event"
        @update:auto-cleanup-trash="draft.autoCleanupTrash = $event"
        @update:trash-retention-days="draft.trashRetentionDays = $event"
      />

      <SettingsSecurityPanel
        :draft="draft"
        :session-options="sessionOptions"
        @update:enforce2fa="draft.enforce2fa = $event"
        @update:force-suspicious-logout="draft.forceSuspiciousLogout = $event"
        @update:session-timeout="draft.sessionTimeout = $event"
      />

      <SettingsSharingPanel
        :draft="draft"
        :share-permission-options="sharePermissionOptions"
        @update:share-links-expire="draft.shareLinksExpire = $event"
        @update:default-share-permission="draft.defaultSharePermission = $event"
        @update:audit-public-downloads="draft.auditPublicDownloads = $event"
      />

      <SettingsEditorPanel
        :draft="draft"
        :open-mode-options="openModeOptions"
        @update:editor-autosave="draft.editorAutosave = $event"
        @update:lock-while-editing="draft.lockWhileEditing = $event"
        @update:default-open-mode="draft.defaultOpenMode = $event"
      />

      <SettingsSuggestionsPanel
        :suggested-groups="suggestedGroups"
        @reset="resetDraft"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useAuthStore } from '~/stores/auth'
import SettingsEditorPanel from '~/components/settings/SettingsEditorPanel.vue'
import SettingsHero from '~/components/settings/SettingsHero.vue'
import SettingsProfilePanel from '~/components/settings/SettingsProfilePanel.vue'
import SettingsSecurityPanel from '~/components/settings/SettingsSecurityPanel.vue'
import SettingsSharingPanel from '~/components/settings/SettingsSharingPanel.vue'
import SettingsStoragePanel from '~/components/settings/SettingsStoragePanel.vue'
import SettingsSuggestionsPanel from '~/components/settings/SettingsSuggestionsPanel.vue'
import type { SettingsDraft, SettingsOption, SuggestedSettingsGroup } from '~/types/settings'

definePageMeta({
  layout: 'default',
  middleware: 'auth',
})

useHead({
  title: '1Drive - Settings',
})

const authStore = useAuthStore()
const { currentUser } = storeToRefs(authStore)

const defaultDraft = (): SettingsDraft => ({
  storageAlerts: true,
  autoCleanupTrash: true,
  trashRetentionDays: 30,
  enforce2fa: false,
  forceSuspiciousLogout: true,
  sessionTimeout: 12,
  shareLinksExpire: true,
  defaultSharePermission: 'viewer',
  auditPublicDownloads: true,
  editorAutosave: true,
  lockWhileEditing: true,
  defaultOpenMode: 'preview',
})

const draft = ref<SettingsDraft>(defaultDraft())

const retentionOptions: SettingsOption<number>[] = [
  { label: '7 ngày', value: 7 },
  { label: '30 ngày', value: 30 },
  { label: '60 ngày', value: 60 },
  { label: '90 ngày', value: 90 },
]

const sessionOptions: SettingsOption<number>[] = [
  { label: '4 giờ', value: 4 },
  { label: '12 giờ', value: 12 },
  { label: '24 giờ', value: 24 },
  { label: '7 ngày', value: 168 },
]

const sharePermissionOptions: SettingsOption<SettingsDraft['defaultSharePermission']>[] = [
  { label: 'Chỉ xem', value: 'viewer' },
  { label: 'Cho phép chỉnh sửa', value: 'editor' },
]

const openModeOptions: SettingsOption<SettingsDraft['defaultOpenMode']>[] = [
  { label: 'Preview trước', value: 'preview' },
  { label: 'Mở editor trực tiếp', value: 'editor' },
]

const suggestedGroups: SuggestedSettingsGroup[] = [
  {
    title: 'Tài khoản và hồ sơ',
    description: 'Tên hiển thị, avatar, xác minh email, ngôn ngữ giao diện, múi giờ.',
    icon: 'mdi:account-circle-outline',
  },
  {
    title: 'Bảo mật',
    description: 'Đổi mật khẩu, 2FA, session timeout, thiết bị tin cậy, nhật ký đăng nhập.',
    icon: 'mdi:shield-lock-outline',
  },
  {
    title: 'Lưu trữ',
    description: 'Quota, cảnh báo dung lượng, versioning, trash retention, lifecycle policy.',
    icon: 'mdi:database-cog-outline',
  },
  {
    title: 'Chia sẻ và phân quyền',
    description: 'Quyền mặc định, hạn dùng link public, watermark, audit download/share.',
    icon: 'mdi:account-key-outline',
  },
  {
    title: 'Editor và preview',
    description: 'OnlyOffice, autosave, lock file, chế độ mở mặc định, preview media/PDF.',
    icon: 'mdi:file-document-edit-outline',
  },
  {
    title: 'Thông báo và vận hành',
    description: 'Email cảnh báo quota, webhook, thông báo file bị xóa/chia sẻ/thất bại upload.',
    icon: 'mdi:bell-badge-outline',
  },
]

const storageUsagePercent = computed(() => {
  const used = currentUser.value?.storage_used_bytes || 0
  const quota = currentUser.value?.storage_quota_bytes || 0
  if (!quota) return 0
  return Math.min(100, Math.round((used / quota) * 100))
})

const usedStorageText = computed(() =>
  `${formatBytes(currentUser.value?.storage_used_bytes || 0)} / ${formatBytes(currentUser.value?.storage_quota_bytes || 0)}`,
)

const quotaText = computed(() => formatBytes(currentUser.value?.storage_quota_bytes || 0))
const initials = computed(() => (currentUser.value?.full_name || 'U').trim().charAt(0).toUpperCase())
const roleLabel = computed(() => currentUser.value?.role || 'user')
const createdAtText = computed(() => formatDate(currentUser.value?.created_at))
const lastLoginText = computed(() => formatDate(currentUser.value?.last_login_at))

onMounted(() => {
  const saved = localStorage.getItem('onedrive.settings.draft')
  if (!saved) return

  try {
    draft.value = {
      ...defaultDraft(),
      ...JSON.parse(saved),
    }
  } catch {
    draft.value = defaultDraft()
  }
})

watch(draft, (value) => {
  localStorage.setItem('onedrive.settings.draft', JSON.stringify(value))
}, { deep: true })

function resetDraft() {
  draft.value = defaultDraft()
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

function formatDate(value: string | null | undefined) {
  if (!value) return 'Chưa có'

  return new Intl.DateTimeFormat('vi-VN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}
</script>

<style scoped>
.settings-page {
  position: relative;
  min-height: 100%;
  padding: 1.5rem;
  isolation: isolate;
}

.settings-page::before {
  content: '';
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  background:
    radial-gradient(circle at top right, rgba(6, 182, 212, 0.14), transparent 26%),
    radial-gradient(circle at 18% 72%, rgba(59, 130, 246, 0.16), transparent 32%),
    radial-gradient(circle at 80% 100%, rgba(14, 165, 233, 0.08), transparent 22%),
    linear-gradient(180deg, rgba(10, 10, 15, 0.92), rgba(10, 10, 15, 0.98)),
    var(--color-bg-primary);
}

.settings-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

@media (max-width: 1024px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .settings-page {
    padding: 1rem;
  }
}
</style>
