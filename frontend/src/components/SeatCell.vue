<script setup lang="ts">
import { computed } from 'vue'
import type { ShowtimeSeat } from '../types'

const props = defineProps<{
  seat: ShowtimeSeat
  isSelected: boolean
}>()

const emit = defineEmits<{ toggle: [] }>()

const classes = computed(() => ({
  selected: props.isSelected,
  locked: props.seat.status === 'LOCKED' && !props.isSelected,
  booked: props.seat.status === 'BOOKED',
}))

const disabled = computed(
  () => props.seat.status === 'BOOKED' || (props.seat.status === 'LOCKED' && !props.isSelected),
)
</script>

<template>
  <button
    class="seat"
    :class="classes"
    :disabled="disabled"
    :title="`${seat.seat_label} - ${seat.status}`"
    @click="emit('toggle')"
  >
    {{ seat.seat_label }}
  </button>
</template>

<style scoped>
.seat {
  width: 2.4rem;
  height: 2.4rem;
  padding: 0;
  font-size: 0.7rem;
  border-radius: 6px;
  background: var(--available);
  color: var(--text-dim);
  border: 1px solid var(--border);
}
.seat.selected {
  background: var(--selected);
  color: white;
  border-color: var(--selected);
}
.seat.locked {
  background: var(--locked);
  color: white;
  border-color: var(--locked);
  cursor: not-allowed;
}
.seat.booked {
  background: var(--booked);
  color: #d0d0d5;
  cursor: not-allowed;
  opacity: 0.7;
}
</style>
