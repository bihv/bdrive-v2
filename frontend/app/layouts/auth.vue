<template>
  <div class="auth-layout">
    <div class="auth-bg" />
    <div class="auth-content">
      <slot />
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
  padding: 2rem 1rem;
  position: relative;
  z-index: 1;
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
</style>
