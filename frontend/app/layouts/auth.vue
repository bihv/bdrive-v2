<template>
  <div class="auth-layout">
    <div class="auth-bg" />
    <div class="auth-content">
      <div class="auth-shell glass-card">
        <section class="auth-brand">
          <span class="auth-eyebrow">Secure workspace</span>
          <h1>1Drive</h1>
          <p class="auth-brand-copy">
            Quản lý file, chia sẻ tài liệu và cộng tác trên cùng một không gian cloud có kiểm soát.
          </p>

          <div class="auth-feature-list">
            <div class="auth-feature">
              <strong>Storage control</strong>
              <span>Quota, trash lifecycle và preview cho tài liệu/media.</span>
            </div>
            <div class="auth-feature">
              <strong>Collaboration</strong>
              <span>OnlyOffice, chia sẻ theo quyền và theo dõi hoạt động truy cập.</span>
            </div>
            <div class="auth-feature">
              <strong>Operational clarity</strong>
              <span>Thiết kế tối ưu cho dashboard nội bộ thay vì landing page chung chung.</span>
            </div>
          </div>
        </section>

        <section class="auth-slot">
          <slot />
        </section>
      </div>
    </div>
    <footer class="auth-footer">
      <span class="auth-footer-text">
        Powered by Backblaze B2 ☁️
      </span>
    </footer>
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
}

.auth-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  position: relative;
  z-index: 1;
}

.auth-shell {
  width: min(1120px, 100%);
  display: grid;
  grid-template-columns: minmax(320px, 1.05fr) minmax(340px, 0.95fr);
  overflow: hidden;
  border: 1px solid var(--color-border);
}

.auth-brand,
.auth-slot {
  padding: 2.25rem;
}

.auth-brand {
  position: relative;
  background:
    radial-gradient(circle at top left, rgba(6, 182, 212, 0.15), transparent 28%),
    radial-gradient(circle at bottom left, rgba(59, 130, 246, 0.16), transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.02), rgba(255, 255, 255, 0.01));
  border-right: 1px solid var(--color-border);
}

.auth-eyebrow {
  display: inline-flex;
  margin-bottom: 0.75rem;
  color: var(--color-accent);
  font-size: var(--font-size-xs);
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.auth-brand h1 {
  font-size: clamp(2.5rem, 4vw, 4rem);
  letter-spacing: -0.04em;
  margin-bottom: 1rem;
}

.auth-brand-copy {
  max-width: 40ch;
  color: var(--color-text-secondary);
  font-size: var(--font-size-lg);
}

.auth-feature-list {
  display: grid;
  gap: 0.85rem;
  margin-top: 2rem;
}

.auth-feature {
  padding: 1rem 1.1rem;
  border-radius: var(--radius-lg);
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
}

.auth-feature strong {
  display: block;
  margin-bottom: 0.25rem;
}

.auth-feature span {
  color: var(--color-text-secondary);
}

.auth-slot {
  display: flex;
  align-items: center;
  justify-content: center;
}

.auth-footer {
  text-align: center;
  padding: 1.5rem;
  position: relative;
  z-index: 1;
}

.auth-footer-text {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
}

@media (max-width: 900px) {
  .auth-content {
    padding: 1rem;
  }

  .auth-shell {
    grid-template-columns: 1fr;
  }

  .auth-brand {
    border-right: none;
    border-bottom: 1px solid var(--color-border);
  }
}

@media (max-width: 640px) {
  .auth-brand,
  .auth-slot {
    padding: 1.5rem;
  }
}
</style>
