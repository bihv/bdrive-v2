<template>
  <div class="auth-page">
    <div class="register-card animate-fade-in-up">
      <div class="register-header">
        <span class="register-kicker">Get started</span>
        <Logo />
        <h2>Tạo tài khoản 1Drive</h2>
        <p class="register-subtitle">Thiết lập không gian lưu trữ an toàn cho nhóm và tài liệu nội bộ.</p>
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
  align-items: flex-start;
  gap: 0.75rem;
  margin-bottom: 2rem;
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
  .register-header h2 {
    font-size: 1.5rem;
  }
}
</style>
