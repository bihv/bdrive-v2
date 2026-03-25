<template>
  <section class="settings-hero glass-card">
    <div>
      <span class="eyebrow">System settings</span>
      <h1>Settings</h1>
      <p>
        Trang này gom các thiết lập nên có cho hệ thống lưu trữ file: tài khoản, bảo mật,
        quota, chia sẻ, chỉnh sửa tài liệu và thông báo vận hành.
      </p>
    </div>

    <div class="hero-stats">
      <div class="stat-card">
        <span class="stat-label">Dung lượng đã dùng</span>
        <strong>{{ usedStorageText }}</strong>
        <small>{{ storageUsagePercent }}% quota</small>
      </div>
      <div class="stat-card">
        <span class="stat-label">Vai trò</span>
        <strong>{{ roleLabel }}</strong>
        <small>{{ isVerified ? 'Đã xác minh' : 'Chưa xác minh' }}</small>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  usedStorageText: string
  storageUsagePercent: number
  roleLabel: string
  isVerified: boolean
}>()
</script>

<style scoped>
.settings-hero {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(280px, 1fr);
  gap: 1.25rem;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  border: 1px solid var(--color-border);
}

.eyebrow {
  display: inline-flex;
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--color-accent);
  margin-bottom: 0.65rem;
}

.settings-hero h1 {
  font-size: clamp(2rem, 4vw, 3rem);
  letter-spacing: -0.03em;
  margin-bottom: 0.75rem;
}

.settings-hero p {
  color: var(--color-text-secondary);
}

.hero-stats {
  display: grid;
  gap: 1rem;
  grid-template-columns: 1fr;
}

.stat-card {
  padding: 1rem;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.stat-label {
  display: block;
  color: var(--color-text-muted);
  font-size: var(--font-size-xs);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-bottom: 0.4rem;
}

.stat-card strong {
  font-size: var(--font-size-xl);
}

@media (max-width: 1024px) {
  .settings-hero {
    grid-template-columns: 1fr;
  }
}
</style>
