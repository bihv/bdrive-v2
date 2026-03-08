<template>
  <div class="logo-container">
    <div class="logo-icon">
      <svg
        viewBox="0 0 40 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        class="logo-svg"
      >
        <!-- Cloud shape -->
        <path
          d="M30 20.5C30 16.08 26.42 12.5 22 12.5C18.88 12.5 16.19 14.26 14.88 16.84C14.6 16.8 14.3 16.78 14 16.78C10.69 16.78 8 19.47 8 22.78C8 26.09 10.69 28.78 14 28.78H29C31.76 28.78 34 26.54 34 23.78C34 21.28 32.19 19.22 29.81 18.84C29.93 18.08 30 17.3 30 16.5"
          stroke="url(#cloudGrad)"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <!-- Upload arrow -->
        <path
          d="M20 24V17M20 17L17 20M20 17L23 20"
          stroke="url(#arrowGrad)"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="logo-arrow"
        />
        <defs>
          <linearGradient id="cloudGrad" x1="8" y1="12" x2="34" y2="29" gradientUnits="userSpaceOnUse">
            <stop stop-color="#3b82f6" />
            <stop offset="1" stop-color="#06b6d4" />
          </linearGradient>
          <linearGradient id="arrowGrad" x1="17" y1="17" x2="23" y2="24" gradientUnits="userSpaceOnUse">
            <stop stop-color="#3b82f6" />
            <stop offset="1" stop-color="#06b6d4" />
          </linearGradient>
        </defs>
      </svg>
    </div>
    <span v-if="showText" class="logo-text">
      <span class="logo-text-one">1</span>Drive
    </span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  showText?: boolean
}>(), {
  showText: true,
})
</script>

<style scoped>
.logo-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-svg {
  width: 100%;
  height: 100%;
}

.logo-arrow {
  animation: float-arrow 3s ease-in-out infinite;
}

@keyframes float-arrow {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-2px);
  }
}

.logo-text {
  font-size: var(--font-size-2xl);
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.02em;
}

.logo-text-one {
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-accent) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-weight: 800;
}
</style>
