<template>
  <div class="auth-page">
    <div class="login-card animate-fade-in-up">
      <div class="login-header">
        <span class="login-kicker">Đăng nhập</span>
        <Logo />
        <h2>Đăng nhập 1Drive</h2>
        <p class="login-subtitle">Tiếp tục làm việc.</p>
      </div>

      <LoginForm
        @success="onLoginSuccess"
        @error="onLoginError"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'auth',
})

useHead({
  title: 'Sign In - 1Drive',
  meta: [
    { name: 'description', content: 'Sign in to 1Drive - Secure cloud file management' },
  ],
})

function onLoginSuccess() {
  // Redirect handled by LoginForm
}

function onLoginError(message: string) {
  console.error('Login failed:', message)
}
</script>

<style scoped>
.auth-page {
  width: 100%;
}

.login-card {
  width: 100%;
  max-width: 460px;
}

.login-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 2rem;
  text-align: center;
}

.login-kicker {
  color: var(--color-accent);
  font-size: var(--font-size-xs);
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.login-header h2 {
  letter-spacing: -0.03em;
  font-size: clamp(1.75rem, 3vw, 2.25rem);
}

.login-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
}

@media (max-width: 480px) {
  .login-card {
    max-width: none;
  }

  .login-header h2 {
    font-size: 1.5rem;
    line-height: 1.05;
  }

  .login-header {
    gap: 0.45rem;
    margin-bottom: 1.1rem;
  }

  .login-kicker,
  .login-header :deep(.logo-container) {
    display: none;
  }

  .login-subtitle {
    font-size: 0.9rem;
    line-height: 1.45;
  }
}
</style>
