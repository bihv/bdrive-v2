<template>
  <form class="login-form" @submit.prevent="handleSubmit">
    <!-- Email -->
    <div class="form-group">
      <label class="form-label" for="login-email">Email</label>
      <n-input
        id="login-email"
        v-model:value="form.email"
        type="text"
        placeholder="name@example.com"
        size="large"
        :status="errors.email ? 'error' : undefined"
        :disabled="loading"
        autocomplete="email"
        @keydown.enter="handleSubmit"
      >
        <template #prefix>
          <n-icon size="18" color="var(--color-text-muted)">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="20" height="16" x="2" y="4" rx="2" /><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7" /></svg>
          </n-icon>
        </template>
      </n-input>
      <span v-if="errors.email" class="field-error">{{ errors.email }}</span>
    </div>

    <!-- Password -->
    <div class="form-group">
      <label class="form-label" for="login-password">Password</label>
      <n-input
        id="login-password"
        v-model:value="form.password"
        type="password"
        placeholder="Enter your password"
        size="large"
        show-password-on="click"
        :status="errors.password ? 'error' : undefined"
        :disabled="loading"
        autocomplete="current-password"
        @keydown.enter="handleSubmit"
      >
        <template #prefix>
          <n-icon size="18" color="var(--color-text-muted)">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="11" x="3" y="11" rx="2" ry="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
          </n-icon>
        </template>
      </n-input>
      <span v-if="errors.password" class="field-error">{{ errors.password }}</span>
    </div>

    <!-- Remember + Forgot -->
    <div class="form-row">
      <n-checkbox v-model:checked="rememberMe" :disabled="loading">
        Remember me
      </n-checkbox>
      <NuxtLink to="/auth/forgot-password" class="forgot-link">
        Forgot password?
      </NuxtLink>
    </div>

    <!-- Error message -->
    <div v-if="errorMessage" class="error-alert animate-shake">
      <n-alert type="error" :bordered="false">
        {{ errorMessage }}
      </n-alert>
    </div>

    <!-- Submit -->
    <n-button
      type="primary"
      size="large"
      block
      :loading="loading"
      :disabled="loading"
      attr-type="submit"
      class="submit-btn"
    >
      <template v-if="!loading">
        Sign In
      </template>
      <template v-else>
        Signing in...
      </template>
    </n-button>

    <!-- Register link -->
    <div class="register-link">
      <span class="register-text">Don't have an account?</span>
      <NuxtLink to="/auth/register" class="register-anchor">
        Create one →
      </NuxtLink>
    </div>
  </form>
</template>

<script setup lang="ts">
import { useAuth } from '~/composables/useAuth'

const emit = defineEmits<{
  success: []
  error: [message: string]
}>()

const auth = useAuth()

const form = reactive({
  email: '',
  password: '',
})

const rememberMe = ref(false)
const loading = ref(false)
const errorMessage = ref('')
const errors = reactive<Record<string, string>>({})

function validate(): boolean {
  // Clear previous errors
  Object.keys(errors).forEach(key => delete errors[key])

  if (!form.email) {
    errors.email = 'Email is required'
  }
  else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email)) {
    errors.email = 'Please enter a valid email'
  }

  if (!form.password) {
    errors.password = 'Password is required'
  }
  else if (form.password.length < 8) {
    errors.password = 'Password must be at least 8 characters'
  }

  return Object.keys(errors).length === 0
}

async function handleSubmit() {
  if (loading.value) return

  errorMessage.value = ''

  if (!validate()) return

  loading.value = true

  try {
    await auth.login({
      email: form.email,
      password: form.password,
    })
    emit('success')
    navigateTo('/')
  }
  catch (err: any) {
    const message = err?.data?.error || err?.message || 'Login failed. Please try again.'
    errorMessage.value = message
    emit('error', message)
  }
  finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-form {
  width: 100%;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 0.5rem;
}

.field-error {
  display: block;
  font-size: 0.75rem;
  color: var(--color-error);
  margin-top: 0.375rem;
}

.form-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.forgot-link {
  font-size: 0.875rem;
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.forgot-link:hover {
  color: var(--color-accent);
}

.error-alert {
  margin-bottom: 1.25rem;
}

.submit-btn {
  margin-bottom: 1.5rem;
  height: 48px !important;
  font-size: 1rem !important;
  font-weight: 600 !important;
  border-radius: var(--radius-md) !important;
  background: linear-gradient(135deg, var(--color-primary) 0%, #2563eb 100%) !important;
  border: none !important;
  transition: all var(--transition-base) !important;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 20px rgba(59, 130, 246, 0.4);
}

.submit-btn:active:not(:disabled) {
  transform: translateY(0);
}

.register-link {
  text-align: center;
}

.register-text {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-right: 0.5rem;
}

.register-anchor {
  font-size: 0.875rem;
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color var(--transition-fast);
}

.register-anchor:hover {
  color: var(--color-accent);
}

@media (max-width: 480px) {
  .form-group {
    margin-bottom: 1rem;
  }

  .form-label,
  .forgot-link,
  .register-text,
  .register-anchor {
    font-size: 0.8125rem;
  }

  .form-row {
    align-items: flex-start;
    gap: 0.75rem;
    margin-bottom: 1.1rem;
  }

  .submit-btn {
    margin-bottom: 1rem;
    height: 44px !important;
    font-size: 0.95rem !important;
  }
}
</style>
