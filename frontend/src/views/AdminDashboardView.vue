<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { moviesApi } from '../api/movies'
import { adminApi } from '../api/admin'
import type { AuditLog, Booking, Movie } from '../types'

const movies = ref<Movie[]>([])
const bookings = ref<Booking[]>([])
const auditLogs = ref<AuditLog[]>([])
const bookingsLoading = ref(true)
const auditLoading = ref(true)

const movieFilter = ref('')
const dateFilter = ref('')
const userEmailFilter = ref('')

async function loadBookings() {
  bookingsLoading.value = true
  try {
    bookings.value = await adminApi.listBookings({
      movie_id: movieFilter.value || undefined,
      date: dateFilter.value || undefined,
      user_email: userEmailFilter.value || undefined,
    })
  } finally {
    bookingsLoading.value = false
  }
}

async function loadAuditLogs() {
  auditLoading.value = true
  try {
    auditLogs.value = await adminApi.listAuditLogs()
  } finally {
    auditLoading.value = false
  }
}

function movieTitle(movieId: string) {
  return movies.value.find((m) => m.id === movieId)?.title ?? movieId
}

function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('en-US')
}

let filterDebounce: ReturnType<typeof setTimeout> | undefined
watch([movieFilter, dateFilter, userEmailFilter], () => {
  clearTimeout(filterDebounce)
  filterDebounce = setTimeout(loadBookings, 300)
})

onMounted(async () => {
  movies.value = await moviesApi.list()
  await Promise.all([loadBookings(), loadAuditLogs()])
})
</script>

<template>
  <div class="page">
    <h1>Admin Dashboard</h1>

    <section class="card">
      <h2>Bookings</h2>
      <div class="filters">
        <select v-model="movieFilter">
          <option value="">All movies</option>
          <option v-for="m in movies" :key="m.id" :value="m.id">{{ m.title }}</option>
        </select>
        <input v-model="dateFilter" type="date" />
        <input v-model="userEmailFilter" type="text" placeholder="Filter by user email" />
      </div>

      <p v-if="bookingsLoading">Loading&hellip;</p>
      <table v-else-if="bookings.length">
        <thead>
          <tr>
            <th>User</th>
            <th>Movie</th>
            <th>Showtime date</th>
            <th>Seats</th>
            <th>Total</th>
            <th>Status</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="b in bookings" :key="b.id">
            <td>{{ b.user_email }}</td>
            <td>{{ movieTitle(b.movie_id) }}</td>
            <td>{{ b.showtime_date }}</td>
            <td>{{ b.seat_labels.join(', ') }}</td>
            <td>฿{{ b.total_amount.toFixed(0) }}</td>
            <td><span class="status" :class="b.status.toLowerCase()">{{ b.status }}</span></td>
            <td>{{ formatDateTime(b.created_at) }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">No bookings match these filters.</p>
    </section>

    <section class="card">
      <h2>Audit Log</h2>
      <p v-if="auditLoading">Loading&hellip;</p>
      <table v-else-if="auditLogs.length">
        <thead>
          <tr>
            <th>Event</th>
            <th>Message</th>
            <th>Time</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in auditLogs" :key="log.id">
            <td><span class="status" :class="log.event_type.toLowerCase()">{{ log.event_type }}</span></td>
            <td>{{ log.message }}</td>
            <td>{{ formatDateTime(log.occurred_at) }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">No audit entries yet.</p>
    </section>
  </div>
</template>

<style scoped>
section.card {
  margin-top: 1.5rem;
}
h2 {
  margin-top: 0;
}
.filters {
  display: flex;
  gap: 0.6rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}
.filters select,
.filters input {
  background: var(--surface-alt);
  border: 1px solid var(--border);
  border-radius: 6px;
  padding: 0.5rem 0.65rem;
  color: var(--text);
}
table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}
th,
td {
  text-align: left;
  padding: 0.5rem 0.6rem;
  border-bottom: 1px solid var(--border);
}
th {
  color: var(--text-dim);
  font-weight: 600;
}
.status {
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
  font-size: 0.8rem;
  background: var(--surface-alt);
}
.status.confirmed,
.status.booking_success {
  color: #3ecf8e;
}
.status.expired,
.status.booking_timeout,
.status.seat_released {
  color: var(--locked);
}
.status.failed,
.status.system_error {
  color: var(--danger);
}
.empty {
  color: var(--text-dim);
}
</style>
