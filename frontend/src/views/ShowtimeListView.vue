<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { moviesApi } from '../api/movies'
import type { Movie, Showtime } from '../types'

const props = defineProps<{ movieId: string }>()

const movie = ref<Movie | null>(null)
const showtimes = ref<Showtime[]>([])
const loading = ref(true)

onMounted(async () => {
  ;[movie.value, showtimes.value] = await Promise.all([
    moviesApi.get(props.movieId),
    moviesApi.showtimes(props.movieId),
  ])
  loading.value = false
})

function formatTime(iso: string) {
  return new Date(iso).toLocaleString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="page">
    <RouterLink to="/movies" class="back">&larr; All movies</RouterLink>
    <p v-if="loading">Loading&hellip;</p>
    <template v-else-if="movie">
      <h1>{{ movie.title }}</h1>
      <p class="meta">{{ movie.genre }} &middot; {{ movie.duration_minutes }} min &middot; {{ movie.rating }}</p>

      <div class="showtime-list">
        <RouterLink
          v-for="st in showtimes"
          :key="st.id"
          :to="{ name: 'seat-map', params: { showtimeId: st.id } }"
          class="card showtime-card"
        >
          <div>
            <strong>{{ formatTime(st.start_time) }}</strong>
            <div class="meta">{{ st.hall_name }}</div>
          </div>
          <div class="price">฿{{ st.base_price.toFixed(0) }}</div>
        </RouterLink>
      </div>
    </template>
  </div>
</template>

<style scoped>
.back {
  display: inline-block;
  margin-bottom: 1rem;
  font-size: 0.9rem;
  text-decoration: none;
}
.meta {
  color: var(--text-dim);
  font-size: 0.85rem;
}
.showtime-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
.showtime-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  text-decoration: none;
  color: inherit;
}
.showtime-card:hover {
  border-color: var(--accent);
}
.price {
  font-weight: 600;
}
</style>
