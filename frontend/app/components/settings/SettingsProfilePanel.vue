<template>
  <SettingsPanelShell
    kicker="Tài khoản"
    title="Hồ sơ người dùng"
    tag-label="Đang dùng"
    tag-type="info"
  >
    <div class="profile-card">
      <div class="profile-avatar">{{ initials }}</div>
      <div>
        <h3>{{ fullName }}</h3>
        <p>{{ email }}</p>
      </div>
    </div>

    <div class="field-grid">
      <div class="field-item">
        <span class="field-label">Ngày tạo</span>
        <strong>{{ createdAtText }}</strong>
      </div>
      <div class="field-item">
        <span class="field-label">Lần đăng nhập cuối</span>
        <strong>{{ lastLoginText }}</strong>
      </div>
    </div>
  </SettingsPanelShell>
</template>

<script setup lang="ts">
import SettingsPanelShell from '~/components/settings/SettingsPanelShell.vue'

defineProps<{
  initials: string
  fullName: string
  email: string
  createdAtText: string
  lastLoginText: string
}>()
</script>

<style scoped>
.profile-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  margin-bottom: 1rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.profile-card p {
  color: var(--color-text-secondary);
}

.profile-avatar {
  width: 3rem;
  height: 3rem;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.28), rgba(6, 182, 212, 0.28));
  font-size: 1.125rem;
  font-weight: 700;
}

.field-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.field-item {
  padding: 0.9rem 1rem;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.field-label {
  display: block;
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 0.4rem;
}

.field-item strong {
  font-size: var(--font-size-xl);
}

@media (max-width: 768px) {
  .field-grid {
    grid-template-columns: 1fr;
  }
}
</style>
