<template>
  <div class="auth-page">
    <div class="login-card animate-fade-in-up">
      <div class="login-header">
        <span class="login-kicker">Welcome back</span>
        <Logo />
        <h2>Đăng nhập để tiếp tục làm việc</h2>
        <p class="login-subtitle">Truy cập file, preview và editor trong cùng một workspace.</p>
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
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 2rem;
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
  .login-header h2 {
    font-size: 1.5rem;
  }
}
</style>
