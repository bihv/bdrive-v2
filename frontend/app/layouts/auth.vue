<template>
  <div class="auth-layout">
    <div class="auth-bg" />
    <div class="auth-content">
      <div class="auth-shell glass-card">
        <section class="auth-slot">
          <slot />
        </section>
      </div>
    </div>
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
  width: min(520px, calc(100% - 2rem));
  border: 1px solid var(--color-border);
  overflow: hidden;
}

.auth-slot {
  padding: 2.25rem;
}

.auth-slot {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

@media (max-width: 640px) {
  .auth-shell {
    width: calc(100% - 1.25rem);
  }

  .auth-slot {
    padding: 1.35rem 1rem 1rem;
  }
}
</style>
