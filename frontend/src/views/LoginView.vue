<script setup lang="ts">
import { authApi } from '../api/auth'

function loginWithGoogle() {
  window.location.href = authApi.googleLoginUrl()
}
</script>

<template>
  <div class="page login-page">
    <div class="card login-card">
      <h1>Cinema Ticket</h1>
      <p class="subtitle">Sign in to book your seats.</p>
      <button @click="loginWithGoogle">Continue with Google</button>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80vh;
}
.login-card {
  width: 100%;
  max-width: 360px;
  text-align: center;
}
.subtitle {
  color: var(--text-dim);
  margin-bottom: 1.5rem;
}
</style>
