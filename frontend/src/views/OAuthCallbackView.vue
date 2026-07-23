<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.store'
import { authApi } from '../api/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

onMounted(async () => {
  const token = route.query.token as string | undefined
  if (!token) {
    router.replace({ name: 'login' })
    return
  }
  auth.setSession(token)
  try {
    auth.setUser(await authApi.me())
  } catch {}
  router.replace({ name: 'movies' })
})
</script>

<template>
  <div class="page">
    <p>Signing you in&hellip;</p>
  </div>
</template>
