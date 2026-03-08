import type { AuthResponse, LoginRequest, RegisterRequest, User } from '~/types/auth'

export function useAuth() {
    const authStore = useAuthStore()
    const api = useApi()
    const router = useRouter()

    async function login(credentials: LoginRequest): Promise<void> {
        const data = await api.post<AuthResponse>('/api/v1/auth/login', credentials)
        authStore.setAuth(data)
    }

    async function register(data: RegisterRequest): Promise<void> {
        const response = await api.post<AuthResponse>('/api/v1/auth/register', data)
        authStore.setAuth(response)
    }

    async function logout(): Promise<void> {
        try {
            await api.post<void>('/api/v1/auth/logout', {})
        }
        catch {
            // Ignore errors during logout
        }
        finally {
            authStore.clearAuth()
            navigateTo('/auth/login')
        }
    }

    async function refreshToken(): Promise<boolean> {
        try {
            const data = await api.post<AuthResponse>('/api/v1/auth/refresh', {})
            authStore.setAuth(data)
            return true
        }
        catch {
            authStore.clearAuth()
            return false
        }
    }

    async function fetchMe(): Promise<User | null> {
        try {
            const user = await api.get<User>('/api/v1/auth/me')
            authStore.setUser(user)
            return user
        }
        catch {
            return null
        }
    }

    return {
        login,
        register,
        logout,
        refreshToken,
        fetchMe,
        isAuthenticated: computed(() => authStore.isAuthenticated),
        currentUser: computed(() => authStore.user),
    }
}
