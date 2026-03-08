import { defineStore } from 'pinia'
import type { User, AuthResponse } from '~/types/auth'

export const useAuthStore = defineStore('auth', {
    state: () => ({
        user: null as User | null,
        accessToken: null as string | null,
        isAuthenticated: false,
    }),

    getters: {
        currentUser: (state) => state.user,
        token: (state) => state.accessToken,
    },

    actions: {
        setAuth(data: AuthResponse) {
            this.accessToken = data.access_token
            this.user = data.user
            this.isAuthenticated = true
        },

        clearAuth() {
            this.accessToken = null
            this.user = null
            this.isAuthenticated = false
        },

        setUser(user: User) {
            this.user = user
        },
    },
})
