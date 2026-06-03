<script setup>
import { computed } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'

const props = defineProps({
  duration: { type: Number, default: 60 },
})

const timelineStore = useTimelineStore()

// Decide tick interval based on zoom (px/s). At low zoom we want coarser ticks.
const tickInterval = computed(() => {
  const z = timelineStore.zoom
  if (z >= 100) return 1
  if (z >= 50) return 2
  if (z >= 20) return 5
  if (z >= 10) return 10
  return 30
})

const majorInterval = computed(() => tickInterval.value * 5)

const ticks = computed(() => {
  const dur = Math.max(props.duration, 30)
  const minor = tickInterval.value
  const major = majorInterval.value
  const result = []
  for (let t = 0; t <= dur; t += minor) {
    result.push({
      t,
      left: t * timelineStore.zoom,
      major: Math.abs((t / major) - Math.round(t / major)) < 0.001,
      label: t % major === 0 ? formatLabel(t) : '',
    })
  }
  return result
})

function formatLabel(seconds) {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

function onClick(e) {
  const rect = e.currentTarget.getBoundingClientRect()
  const x = e.clientX - rect.left
  const t = Math.max(0, x / timelineStore.zoom)
  timelineStore.setCurrentTime(t)
}
</script>

<template>
  <div
    class="relative h-7 bg-bg/40 border-b border-border cursor-pointer select-none"
    @click="onClick"
  >
    <div
      v-for="tick in ticks"
      :key="tick.t"
      class="absolute top-0 bottom-0"
      :style="{ left: tick.left + 'px' }"
    >
      <div
        class="absolute bottom-0 w-px"
        :class="tick.major ? 'h-2.5 bg-text-secondary' : 'h-1.5 bg-border'"
      />
      <div
        v-if="tick.label"
        class="absolute top-0.5 left-1 text-[9px] text-text-secondary font-mono whitespace-nowrap"
      >
        {{ tick.label }}
      </div>
    </div>
  </div>
</template>
