<template>
  <div>
    <h1>Welcome to 1Drive</h1>
    <p>Dashboard coming soon...</p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: 'auth',
})

useHead({
  title: '1Drive - Dashboard',
})
</script>
