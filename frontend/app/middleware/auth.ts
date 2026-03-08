export default defineNuxtRouteMiddleware(async (to) => {
    const authStore = useAuthStore()

    if (authStore.isAuthenticated) return

    // Server side: skip redirect, let client handle auth
    if (import.meta.server) return

    // Client side: try refreshing token before redirecting
    const { refreshToken } = useAuth()
    const success = await refreshToken()
    if (success) return

    return navigateTo('/auth/login')
})

