<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { showtimesApi } from '../api/showtimes'
import { bookingsApi } from '../api/bookings'
import { useSeatMapStore } from '../stores/seatMap.store'
import { connectSeatMapSocket } from '../composables/useWebSocket'
import SeatGrid from '../components/SeatGrid.vue'
import type { Showtime } from '../types'

const props = defineProps<{ showtimeId: string }>()
const router = useRouter()
const store = useSeatMapStore()

const showtime = ref<Showtime | null>(null)
const phase = ref<'selecting' | 'processing'>('selecting')
const error = ref('')
let disconnect: (() => void) | null = null

const selectedLabels = computed(() => Array.from(store.selected))
const total = computed(() =>
  selectedLabels.value.reduce((sum, label) => sum + (store.seats[label]?.price ?? 0), 0),
)

onMounted(async () => {
  showtime.value = await showtimesApi.get(props.showtimeId)
  await store.loadSnapshot(props.showtimeId)
  disconnect = connectSeatMapSocket(
    props.showtimeId,
    (msg) => store.applyEvent(msg),
    () => store.loadSnapshot(props.showtimeId),
  )
})

onBeforeUnmount(() => {
  disconnect?.()
  if (phase.value !== 'processing') {
    store.reset()
  }
})

function toggleSeat(label: string) {
  if (phase.value !== 'selecting') return
  store.toggleSelect(label)
}

async function handleContinue() {
  if (selectedLabels.value.length === 0) return
  error.value = ''
  phase.value = 'processing'
  try {
    await store.lockSelected()
    const booking = await bookingsApi.create(props.showtimeId, selectedLabels.value, store.lockToken!)
    router.push({ name: 'checkout', params: { bookingId: booking.id } })
  } catch {
    error.value = 'One or more seats were just taken. Please choose again.'
    phase.value = 'selecting'
    await store.loadSnapshot(props.showtimeId)
    store.selected.clear()
  }
}
</script>

<template>
  <div class="page">
    <p v-if="!showtime">Loading&hellip;</p>
    <template v-else>
      <h1>{{ showtime.hall_name }}</h1>
      <p class="meta">{{ new Date(showtime.start_time).toLocaleString('en-US') }}</p>

      <SeatGrid
        :seats="store.seats"
        :rows="showtime.rows"
        :cols="showtime.cols"
        :selected="store.selected"
        @toggle="toggleSeat"
      />

      <div class="legend">
        <span><i class="dot available" /> Available</span>
        <span><i class="dot selected" /> Your selection</span>
        <span><i class="dot locked" /> Held by another user</span>
        <span><i class="dot booked" /> Booked</span>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <div class="summary card">
        <div>
          <strong>{{ selectedLabels.length }}</strong> seat(s) selected &middot;
          <strong>฿{{ total.toFixed(0) }}</strong>
        </div>
        <div class="actions">
          <button :disabled="selectedLabels.length === 0 || phase === 'processing'" @click="handleContinue">
            {{ phase === 'processing' ? 'Processing…' : 'Continue' }}
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.meta {
  color: var(--text-dim);
  margin-bottom: 1.5rem;
}
.legend {
  display: flex;
  gap: 1.25rem;
  justify-content: center;
  margin-top: 1.5rem;
  font-size: 0.8rem;
  color: var(--text-dim);
}
.dot {
  display: inline-block;
  width: 0.7rem;
  height: 0.7rem;
  border-radius: 3px;
  margin-right: 0.3rem;
}
.dot.available {
  background: var(--available);
}
.dot.selected {
  background: var(--selected);
}
.dot.locked {
  background: var(--locked);
}
.dot.booked {
  background: var(--booked);
}
.error {
  color: var(--danger);
  text-align: center;
  margin-top: 1rem;
}
.summary {
  margin-top: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}
.actions {
  display: flex;
  gap: 0.6rem;
}
</style>
