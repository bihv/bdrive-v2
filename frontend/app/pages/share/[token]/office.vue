<script setup lang="ts">
definePageMeta({
  layout: false,
})

const route = useRoute()
const itemId = Array.isArray(route.query.item) ? route.query.item[0] : route.query.item
const session = Array.isArray(route.query.session) ? route.query.session[0] : route.query.session
const name = Array.isArray(route.query.name) ? route.query.name[0] : route.query.name

await navigateTo({
  path: '/office',
  query: {
    ...(itemId ? { id: String(itemId) } : {}),
    share_token: String(route.params.token || ''),
    ...(session ? { session: String(session) } : {}),
    ...(name ? { name: String(name) } : {}),
  },
}, { replace: true })
</script>
