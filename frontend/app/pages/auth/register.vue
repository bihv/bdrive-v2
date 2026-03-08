<template>
  <div>
    <div class="register-card glass-card animate-fade-in-up">
      <!-- Logo & Header -->
      <div class="register-header">
        <UiLogo />
        <p class="register-subtitle">Create your 1Drive account</p>
      </div>

      <!-- Register Form -->
      <AuthRegisterForm
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
.register-card {
  width: 100%;
  max-width: 420px;
  padding: 2.5rem;
}

.register-header {
  text-align: center;
  margin-bottom: 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.register-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-secondary);
  font-weight: 400;
}

/* Responsive */
@media (max-width: 480px) {
  .register-card {
    padding: 1.5rem;
    margin: 0 0.5rem;
    border-radius: var(--radius-md);
  }
}
</style>
