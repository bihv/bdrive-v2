<template>
  <form class="register-form" @submit.prevent="handleSubmit">
    <!-- Full Name -->
    <div class="form-group">
      <label class="form-label" for="register-name">Full Name</label>
      <n-input
        id="register-name"
        v-model:value="form.fullName"
        type="text"
        placeholder="John Doe"
        size="large"
        :status="errors.fullName ? 'error' : undefined"
        :disabled="loading"
        autocomplete="name"
      >
        <template #prefix>
          <n-icon size="18" color="var(--color-text-muted)">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="8" r="5"/><path d="M20 21a8 8 0 0 0-16 0"/></svg>
          </n-icon>
        </template>
      </n-input>
      <span v-if="errors.fullName" class="field-error">{{ errors.fullName }}</span>
    </div>

    <!-- Email -->
    <div class="form-group">
      <label class="form-label" for="register-email">Email</label>
      <n-input
        id="register-email"
        v-model:value="form.email"
        type="text"
        placeholder="name@example.com"
        size="large"
        :status="errors.email ? 'error' : undefined"
        :disabled="loading"
        autocomplete="email"
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
      <label class="form-label" for="register-password">Password</label>
      <n-input
        id="register-password"
        v-model:value="form.password"
        type="password"
        placeholder="At least 8 characters"
        size="large"
        show-password-on="click"
        :status="errors.password ? 'error' : undefined"
        :disabled="loading"
        autocomplete="new-password"
      >
        <template #prefix>
          <n-icon size="18" color="var(--color-text-muted)">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="18" height="11" x="3" y="11" rx="2" ry="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
          </n-icon>
        </template>
      </n-input>
      <span v-if="errors.password" class="field-error">{{ errors.password }}</span>

      <!-- Password strength indicator -->
      <div v-if="form.password" class="password-strength">
        <div class="strength-bars">
          <div
            v-for="i in 4"
            :key="i"
            class="strength-bar"
            :class="{ active: passwordStrength >= i }"
            :style="{ backgroundColor: i <= passwordStrength ? strengthColor : undefined }"
          />
        </div>
        <span class="strength-label" :style="{ color: strengthColor }">{{ strengthLabel }}</span>
      </div>
    </div>

    <!-- Confirm Password -->
    <div class="form-group">
      <label class="form-label" for="register-confirm">Confirm Password</label>
      <n-input
        id="register-confirm"
        v-model:value="form.confirmPassword"
        type="password"
        placeholder="Re-enter your password"
        size="large"
        show-password-on="click"
        :status="errors.confirmPassword ? 'error' : undefined"
        :disabled="loading"
        autocomplete="new-password"
      >
        <template #prefix>
          <n-icon size="18" color="var(--color-text-muted)">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10"/><path d="m9 12 2 2 4-4"/></svg>
          </n-icon>
        </template>
      </n-input>
      <span v-if="errors.confirmPassword" class="field-error">{{ errors.confirmPassword }}</span>
    </div>

    <!-- Terms -->
    <div class="form-group terms-group">
      <n-checkbox v-model:checked="agreeTerms" :disabled="loading">
        I agree to the
        <NuxtLink to="/terms" class="terms-link">Terms of Service</NuxtLink>
        and
        <NuxtLink to="/privacy" class="terms-link">Privacy Policy</NuxtLink>
      </n-checkbox>
      <span v-if="errors.terms" class="field-error">{{ errors.terms }}</span>
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
        Create Account
      </template>
      <template v-else>
        Creating account...
      </template>
    </n-button>

    <!-- Login link -->
    <div class="login-link">
      <span class="login-text">Already have an account?</span>
      <NuxtLink to="/auth/login" class="login-anchor">
        Sign in →
      </NuxtLink>
    </div>
  </form>
</template>

<script setup lang="ts">
const emit = defineEmits<{
  success: []
  error: [message: string]
}>()

const auth = useAuth()

const form = reactive({
  fullName: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const agreeTerms = ref(false)
const loading = ref(false)
const errorMessage = ref('')
const errors = reactive<Record<string, string>>({})

// Password strength computation
const passwordStrength = computed(() => {
  const p = form.password
  if (!p) return 0
  let score = 0
  if (p.length >= 8) score++
  if (/[A-Z]/.test(p) && /[a-z]/.test(p)) score++
  if (/[0-9]/.test(p)) score++
  if (/[^A-Za-z0-9]/.test(p)) score++
  return score
})

const strengthLabel = computed(() => {
  const labels = ['', 'Weak', 'Fair', 'Good', 'Strong']
  return labels[passwordStrength.value] || ''
})

const strengthColor = computed(() => {
  const colors = ['', '#ef4444', '#f59e0b', '#22c55e', '#06b6d4']
  return colors[passwordStrength.value] || ''
})

function validate(): boolean {
  Object.keys(errors).forEach(key => delete errors[key])

  if (!form.fullName) {
    errors.fullName = 'Full name is required'
  }
  else if (form.fullName.length < 2) {
    errors.fullName = 'Name must be at least 2 characters'
  }

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

  if (!form.confirmPassword) {
    errors.confirmPassword = 'Please confirm your password'
  }
  else if (form.password !== form.confirmPassword) {
    errors.confirmPassword = 'Passwords do not match'
  }

  if (!agreeTerms.value) {
    errors.terms = 'You must agree to the terms'
  }

  return Object.keys(errors).length === 0
}

async function handleSubmit() {
  if (loading.value) return

  errorMessage.value = ''

  if (!validate()) return

  loading.value = true

  try {
    await auth.register({
      email: form.email,
      password: form.password,
      full_name: form.fullName,
    })
    emit('success')
    navigateTo('/')
  }
  catch (err: any) {
    console.error('Register API Error:', err)
    const message = err?.data?.error || err?.message || 'Registration failed. Please try again.'
    errorMessage.value = message
    emit('error', message)
  }
  finally {
    loading.value = false
  }
}
</script>

<style scoped>
.register-form {
  width: 100%;
}

.form-group {
  margin-bottom: 1.125rem;
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

/* Password strength indicator */
.password-strength {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.strength-bars {
  display: flex;
  gap: 4px;
  flex: 1;
}

.strength-bar {
  height: 4px;
  flex: 1;
  border-radius: 2px;
  background: var(--color-border);
  transition: background-color var(--transition-base);
}

.strength-label {
  font-size: 0.75rem;
  font-weight: 500;
  min-width: 45px;
  text-align: right;
}

.terms-group {
  margin-bottom: 1.5rem;
}

.terms-link {
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.terms-link:hover {
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

.login-link {
  text-align: center;
}

.login-text {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-right: 0.5rem;
}

.login-anchor {
  font-size: 0.875rem;
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color var(--transition-fast);
}

.login-anchor:hover {
  color: var(--color-accent);
}
</style>
