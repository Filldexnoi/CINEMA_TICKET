<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { bookingsApi } from '../api/bookings'
import CountdownTimer from '../components/CountdownTimer.vue'
import type { Booking } from '../types'

const props = defineProps<{ bookingId: string }>()
const router = useRouter()

const booking = ref<Booking | null>(null)
const paying = ref(false)
const error = ref('')

onMounted(async () => {
  booking.value = await bookingsApi.get(props.bookingId)
})

async function pay(result: 'success' | 'fail') {
  paying.value = true
  error.value = ''
  try {
    const updated = await bookingsApi.pay(props.bookingId, result)
    booking.value = updated
    if (updated.status === 'CONFIRMED') {
      router.push({ name: 'confirmation', params: { bookingId: updated.id } })
    } else {
      error.value = 'Payment failed. Your seats have been released.'
    }
  } catch {
    error.value = 'Your seat hold expired before payment completed.'
  } finally {
    paying.value = false
  }
}

function onHoldExpired() {
  error.value = 'Your seat hold expired before payment completed.'
}
</script>

<template>
  <div class="page">
    <p v-if="!booking">Loading&hellip;</p>
    <template v-else>
      <h1>Checkout</h1>
      <div class="card">
        <p>Seats: <strong>{{ booking.seat_labels.join(', ') }}</strong></p>
        <p>Total: <strong>฿{{ booking.total_amount.toFixed(0) }}</strong></p>
        <p v-if="booking.status === 'PENDING'">
          Complete payment within
          <CountdownTimer :expires-at="booking.expires_at" @expired="onHoldExpired" />
        </p>

        <p v-if="error" class="error">{{ error }}</p>

        <div class="actions" v-if="booking.status === 'PENDING'">
          <button :disabled="paying" @click="pay('success')">Pay now (simulated)</button>
          <button class="secondary" :disabled="paying" @click="pay('fail')">Simulate declined payment</button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.card {
  margin-top: 1.5rem;
  max-width: 420px;
}
.error {
  color: var(--danger);
}
.actions {
  display: flex;
  gap: 0.6rem;
  margin-top: 1rem;
}
</style>
