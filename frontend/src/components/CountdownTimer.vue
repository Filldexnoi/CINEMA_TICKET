<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

const props = defineProps<{ expiresAt: string }>()
const emit = defineEmits<{ expired: [] }>()

const now = ref(Date.now())
let interval: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  interval = setInterval(() => {
    now.value = Date.now()
    if (remainingMs.value <= 0) emit('expired')
  }, 1000)
})
onBeforeUnmount(() => clearInterval(interval))

const remainingMs = computed(() => new Date(props.expiresAt).getTime() - now.value)
const label = computed(() => {
  const totalSeconds = Math.max(0, Math.floor(remainingMs.value / 1000))
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
})
</script>

<template>
  <span class="countdown" :class="{ urgent: remainingMs < 60_000 }">{{ label }}</span>
</template>

<style scoped>
.countdown {
  font-variant-numeric: tabular-nums;
  font-weight: 600;
}
.countdown.urgent {
  color: var(--danger);
}
</style>
