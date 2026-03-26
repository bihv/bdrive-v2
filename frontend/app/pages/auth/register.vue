<template>
  <div class="auth-page">
    <div class="register-card animate-fade-in-up">
      <div class="register-header">
        <span class="register-kicker">Tạo tài khoản</span>
        <Logo />
        <h2>Tạo tài khoản</h2>
        <p class="register-subtitle">Bắt đầu với 1Drive.</p>
      </div>

      <RegisterForm
        @success="onRegisterSuccess"
        @error="onRegisterError"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
})

useHead({
  title: 'Create Account - 1Drive',
  meta: [
    { name: 'description', content: 'Create your 1Drive account - Secure cloud file management powered by Backblaze B2' },
  ],
})

function onRegisterSuccess() {
  // Redirect handled by RegisterForm
}

function onRegisterError(message: string) {
  console.error('Registration failed:', message)
}
</script>

<style scoped>
.auth-page {
  width: 100%;
}

.register-card {
  width: 100%;
  max-width: 480px;
}

.register-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 2rem;
  text-align: center;
}

.register-kicker {
  color: var(--color-accent);
  font-size: var(--font-size-xs);
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.register-header h2 {
  letter-spacing: -0.03em;
  font-size: clamp(1.75rem, 3vw, 2.25rem);
}

.register-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

@media (max-width: 480px) {
  .register-card {
    max-width: none;
  }

  .register-header h2 {
    font-size: 1.5rem;
    line-height: 1.05;
  }

  .register-header {
    gap: 0.45rem;
    margin-bottom: 1.1rem;
  }

  .register-kicker,
  .register-header :deep(.logo-container) {
    display: none;
  }

  .register-subtitle {
    font-size: 0.9rem;
    line-height: 1.45;
  }
}
</style>
