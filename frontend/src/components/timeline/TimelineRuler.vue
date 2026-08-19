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

function seekFromClientX(clientX) {
  // Anchor to the fixed-position scroll viewport (.timeline-content), which
  // does NOT move with horizontal scroll, then add the scroll offset to land
  // in content space. (Anchoring to the ruler element instead would already
  // bake in the scroll and double-count it.)
  const container = document.querySelector('.timeline-content')
  const rect = container?.getBoundingClientRect()
  if (!rect) return
  const scrollLeft = container?.scrollLeft ?? 0
  const x = clientX - rect.left + scrollLeft
  const t = Math.max(0, x / timelineStore.zoom)
  timelineStore.setCurrentTime(t)
}

// Press (or press + drag) on the ruler seeks continuously. Seek updates
// are throttled to ~100ms so frame extraction isn't spammed.
let lastSeekAt = 0
function onMouseDown(e) {
  if (e.button !== 0) return
  e.preventDefault()
  seekFromClientX(e.clientX)
  lastSeekAt = performance.now()
  const onMove = (mv) => {
    const now = performance.now()
    if (now - lastSeekAt < 100) return
    lastSeekAt = now
    seekFromClientX(mv.clientX)
  }
  const onUp = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}
</script>

<template>
  <div
    class="relative h-7 bg-bg/40 border-b border-border cursor-pointer select-none"
    @mousedown="onMouseDown"
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
