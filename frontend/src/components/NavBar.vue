<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth.store'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <nav v-if="auth.token" class="navbar">
    <RouterLink to="/movies" class="brand">Cinema Ticket</RouterLink>
    <RouterLink v-if="auth.user?.role === 'ADMIN'" to="/admin" class="admin-link">Admin</RouterLink>
    <div class="spacer" />
    <span v-if="auth.user" class="user">{{ auth.user.name }}</span>
    <button class="secondary" @click="logout">Log out</button>
  </nav>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.9rem 1.5rem;
  border-bottom: 1px solid var(--border);
}
.brand {
  font-weight: 700;
  text-decoration: none;
  color: var(--text);
}
.admin-link {
  color: var(--text-dim);
  text-decoration: none;
  font-size: 0.9rem;
}
.admin-link:hover {
  color: var(--text);
}
.spacer {
  flex: 1;
}
.user {
  color: var(--text-dim);
}
</style>
