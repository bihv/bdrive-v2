import type { Item, RecentItem } from '~/types/folder'
import { useAuthStore } from '~/stores/auth'

type RecentEventType = 'open' | 'download' | 'update'

export function useQuickAccess() {
    const store = useQuickAccessStore()
    const authStore = useAuthStore()
    const currentUser = computed(() => authStore.user)
    const api = useApi()

    async function refresh() {
        if (!currentUser.value?.id) {
            store.setResolvedItems({ recentItems: [], starredItems: [] })
            return
        }

        try {
            store.setLoading(true)
            const [recentItems, starredItems] = await Promise.all([
                api.getRecentItems(),
                api.getStarredItems(),
            ])

            store.setResolvedItems({
                recentItems,
                starredItems,
            })
        } finally {
            store.setLoading(false)
        }
    }

    async function track(itemId: string, type: RecentEventType) {
        await api.trackItemActivity(itemId, type)
        await refresh()
    }

    async function toggleStar(itemId: string) {
        if (store.starredIds.includes(itemId)) {
            await api.unstarItem(itemId)
        } else {
            await api.starItem(itemId)
        }
        await refresh()
    }

    return {
        refresh,
        track,
        toggleStar,
        removeItem: (itemId: string) => store.removeItem(itemId),
        loading: computed(() => store.loading),
        recentItems: computed(() => store.recentItems),
        starredItems: computed(() => store.starredItems),
        starredIds: computed(() => store.starredIds),
        isStarred: (itemId: string) => store.isStarred(itemId),
    }
}
