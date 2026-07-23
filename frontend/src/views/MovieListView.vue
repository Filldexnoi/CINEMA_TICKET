<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { moviesApi } from '../api/movies'
import type { Movie } from '../types'

const movies = ref<Movie[]>([])
const loading = ref(true)

onMounted(async () => {
  movies.value = await moviesApi.list()
  loading.value = false
})
</script>

<template>
  <div class="page">
    <h1>Now Showing</h1>
    <p v-if="loading">Loading movies&hellip;</p>
    <div v-else class="grid">
      <RouterLink
        v-for="movie in movies"
        :key="movie.id"
        :to="{ name: 'showtimes', params: { movieId: movie.id } }"
        class="card movie-card"
      >
        <h2>{{ movie.title }}</h2>
        <p class="meta">{{ movie.genre }} &middot; {{ movie.duration_minutes }} min &middot; {{ movie.rating }}</p>
        <p class="description">{{ movie.description }}</p>
      </RouterLink>
    </div>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
  margin-top: 1.5rem;
}
.movie-card {
  text-decoration: none;
  color: inherit;
  display: block;
  transition: border-color 0.15s;
}
.movie-card:hover {
  border-color: var(--accent);
}
.meta {
  color: var(--text-dim);
  font-size: 0.85rem;
  margin: 0.35rem 0 0.75rem;
}
.description {
  font-size: 0.9rem;
  color: var(--text-dim);
}
</style>
