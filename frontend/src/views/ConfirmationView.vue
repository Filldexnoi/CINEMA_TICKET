<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { bookingsApi } from '../api/bookings'
import type { Booking } from '../types'

const props = defineProps<{ bookingId: string }>()
const booking = ref<Booking | null>(null)

onMounted(async () => {
  booking.value = await bookingsApi.get(props.bookingId)
})
</script>

<template>
  <div class="page">
    <div v-if="booking" class="card confirmation">
      <h1>Booking confirmed</h1>
      <p class="checkmark">&#10003;</p>
      <p>Seats: <strong>{{ booking.seat_labels.join(', ') }}</strong></p>
      <p>Total paid: <strong>฿{{ booking.total_amount.toFixed(0) }}</strong></p>
      <p class="meta">Booking ID: {{ booking.id }}</p>
      <RouterLink to="/movies">Book another movie</RouterLink>
    </div>
  </div>
</template>

<style scoped>
.confirmation {
  max-width: 420px;
  margin: 2rem auto;
  text-align: center;
}
.checkmark {
  font-size: 2.5rem;
  color: #3ecf8e;
}
.meta {
  color: var(--text-dim);
  font-size: 0.85rem;
}
</style>
