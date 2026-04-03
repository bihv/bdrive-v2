import { useAuth } from '~/composables/useAuth'
import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware(async (to) => {
  const shareToken = Array.isArray(to.query.share_token) ? to.query.share_token[0] : to.query.share_token
  if (typeof shareToken === 'string' && shareToken.trim().length > 0) {
    return
  }

  const authStore = useAuthStore()
  if (authStore.isAuthenticated) return

  if (import.meta.server) return

  const { refreshToken } = useAuth()
  const success = await refreshToken()
  if (success) return

  return navigateTo('/auth/login')
})
