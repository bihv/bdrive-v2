import { defineStore } from 'pinia'
import type { Item, RecentItem } from '~/types/folder'

export const useQuickAccessStore = defineStore('quick-access', {
    state: () => ({
        starredIds: [] as string[],
        recentItems: [] as RecentItem[],
        starredItems: [] as Item[],
        loading: false,
    }),

    actions: {
        removeItem(itemId: string) {
            this.starredIds = this.starredIds.filter(id => id !== itemId)
            this.recentItems = this.recentItems.filter(item => item.id !== itemId)
            this.starredItems = this.starredItems.filter(item => item.id !== itemId)
        },

        setResolvedItems(payload: {
            recentItems: RecentItem[]
            starredItems: Item[]
        }) {
            this.recentItems = payload.recentItems
            this.starredItems = payload.starredItems
            this.starredIds = payload.starredItems.map(item => item.id)
        },

        setLoading(value: boolean) {
            this.loading = value
        },

        isStarred(itemId: string) {
            return this.starredIds.includes(itemId)
        },
    },
})
