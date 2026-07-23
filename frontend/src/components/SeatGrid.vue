<script setup lang="ts">
import { computed } from 'vue'
import type { ShowtimeSeat } from '../types'
import SeatCell from './SeatCell.vue'

const props = defineProps<{
  seats: Record<string, ShowtimeSeat>
  rows: number
  cols: number
  selected: Set<string>
}>()

const emit = defineEmits<{ toggle: [label: string] }>()

const rowLabels = computed(() => Array.from({ length: props.rows }, (_, i) => String.fromCharCode(65 + i)))
const colNumbers = computed(() => Array.from({ length: props.cols }, (_, i) => i + 1))
</script>

<template>
  <div class="grid-wrapper">
    <div class="screen">Screen</div>
    <div class="grid">
      <template v-for="rowLabel in rowLabels" :key="rowLabel">
        <template v-for="col in colNumbers" :key="`${rowLabel}${col}`">
          <SeatCell
            v-if="seats[`${rowLabel}${col}`]"
            :seat="seats[`${rowLabel}${col}`]"
            :is-selected="selected.has(`${rowLabel}${col}`)"
            @toggle="emit('toggle', `${rowLabel}${col}`)"
          />
        </template>
      </template>
    </div>
  </div>
</template>

<style scoped>
.grid-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}
.screen {
  width: 80%;
  text-align: center;
  padding: 0.5rem;
  border-radius: 4px;
  background: var(--surface-alt);
  color: var(--text-dim);
  font-size: 0.75rem;
  letter-spacing: 0.15em;
  text-transform: uppercase;
}
.grid {
  display: grid;
  grid-template-columns: repeat(v-bind(cols), 2.4rem);
  gap: 0.4rem;
}
</style>
