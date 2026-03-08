<template>
  <div>
    <div class="login-card glass-card animate-fade-in-up">
      <!-- Logo & Header -->
      <div class="login-header">
        <UiLogo />
        <p class="login-subtitle">Your Cloud, Your Control</p>
      </div>

      <!-- Login Form -->
      <AuthLoginForm
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
.login-card {
  width: 100%;
  max-width: 420px;
  padding: 2.5rem;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.login-subtitle {
  font-size: var(--font-size-sm);
  color: var(--color-text-muted);
  letter-spacing: 0.05em;
  text-transform: uppercase;
  font-weight: 400;
}

/* Responsive */
@media (max-width: 480px) {
  .login-card {
    padding: 1.5rem;
    margin: 0 0.5rem;
    border-radius: var(--radius-md);
  }
}
</style>
