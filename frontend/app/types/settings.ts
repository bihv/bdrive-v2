export interface SettingsDraft {
  storageAlerts: boolean
  autoCleanupTrash: boolean
  trashRetentionDays: number
  enforce2fa: boolean
  forceSuspiciousLogout: boolean
  sessionTimeout: number
  shareLinksExpire: boolean
  defaultSharePermission: 'viewer' | 'editor'
  auditPublicDownloads: boolean
  editorAutosave: boolean
  lockWhileEditing: boolean
  defaultOpenMode: 'preview' | 'editor'
}

export interface SettingsOption<T extends string | number = string | number> {
  label: string
  value: T
}

export interface SuggestedSettingsGroup {
  title: string
  description: string
  icon: string
}
