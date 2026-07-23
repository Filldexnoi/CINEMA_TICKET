<script setup lang="ts">
import { onMounted } from 'vue'
import NavBar from './components/NavBar.vue'
import { useAuthStore } from './stores/auth.store'
import { authApi } from './api/auth'

const auth = useAuthStore()

onMounted(async () => {
  if (auth.token && !auth.user) {
    try {
      auth.setUser(await authApi.me())
    } catch {}
  }
})
</script>

<template>
  <NavBar />
  <RouterView />
</template>
