import { defineStore } from 'pinia'
import { reactive, ref } from 'vue'
import { showtimesApi } from '../api/showtimes'
import { seatsApi } from '../api/seats'
import type { SeatEventMessage, ShowtimeSeat } from '../types'

export const useSeatMapStore = defineStore('seatMap', () => {
  const showtimeId = ref<string | null>(null)
  const seats = reactive<Record<string, ShowtimeSeat>>({})
  const selected = ref<Set<string>>(new Set())
  const lockToken = ref<string | null>(null)
  const loading = ref(false)

  async function loadSnapshot(id: string) {
    loading.value = true
    try {
      showtimeId.value = id
      const list = await showtimesApi.seats(id)
      for (const key of Object.keys(seats)) delete seats[key]
      for (const seat of list) seats[seat.seat_label] = seat
    } finally {
      loading.value = false
    }
  }

  function applyEvent(msg: SeatEventMessage) {
    if (msg.showtime_id !== showtimeId.value) return
    for (const label of msg.seat_labels) {
      const seat = seats[label]
      if (!seat) continue
      switch (msg.type) {
        case 'seat.locked':
          seat.status = 'LOCKED'
          break
        case 'seat.released':
        case 'booking.expired':
          seat.status = 'AVAILABLE'
          seat.locked_by = undefined
          seat.lock_expires_at = undefined
          seat.booking_id = undefined
          selected.value.delete(label)
          break
        case 'booking.confirmed':
          seat.status = 'BOOKED'
          break
      }
    }
  }

  function toggleSelect(label: string) {
    if (selected.value.has(label)) {
      selected.value.delete(label)
      return
    }
    const seat = seats[label]
    if (!seat || seat.status !== 'AVAILABLE') return
    selected.value.add(label)
  }

  async function lockSelected() {
    if (!showtimeId.value || selected.value.size === 0) return
    const labels = Array.from(selected.value)
    const { lock_token } = await seatsApi.lock(showtimeId.value, labels)
    lockToken.value = lock_token
    for (const label of labels) {
      const seat = seats[label]
      if (seat) seat.status = 'LOCKED'
    }
  }

  function reset() {
    showtimeId.value = null
    for (const key of Object.keys(seats)) delete seats[key]
    selected.value = new Set()
    lockToken.value = null
  }

  return { showtimeId, seats, selected, lockToken, loading, loadSnapshot, applyEvent, toggleSelect, lockSelected, reset }
})
