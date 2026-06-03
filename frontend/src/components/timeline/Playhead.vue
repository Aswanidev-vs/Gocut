<script setup>
import { computed, ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'

const props = defineProps({
  duration: { type: Number, default: 60 },
})

const timelineStore = useTimelineStore()

const leftPx = computed(() => timelineStore.currentTime * timelineStore.zoom)

let dragging = false

function onMouseDown(e) {
  e.preventDefault()
  dragging = true
  seekFromEvent(e)
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

function onMove(e) {
  if (dragging) seekFromEvent(e)
}

function onUp() {
  dragging = false
  document.removeEventListener('mousemove', onMove)
  document.removeEventListener('mouseup', onUp)
}

function seekFromEvent(e) {
  const target = e.currentTarget?.parentElement || e.target
  // We measure relative to the timeline content area
  const rect = (document.querySelector('.timeline-content'))?.getBoundingClientRect()
  if (!rect) return
  const x = e.clientX - rect.left + (document.querySelector('.timeline-content'))?.scrollLeft || 0
  const t = Math.max(0, x / timelineStore.zoom)
  timelineStore.setCurrentTime(t)
}
</script>

<template>
  <div
    class="absolute top-0 bottom-0 z-30 pointer-events-none"
    :style="{ left: leftPx + 'px', width: '1px' }"
  >
    <div class="absolute -top-0 -left-1.5 w-3.5 h-3.5 bg-accent rotate-45 origin-center" style="transform: translateY(-3px) rotate(45deg);" />
    <div class="absolute top-0 bottom-0 left-0 w-px bg-accent shadow-[0_0_6px_rgba(0,212,255,0.6)]" />
    <!-- Wider hit area for click-to-seek at the top -->
    <div
      class="absolute -top-1 -left-3 w-7 h-7 cursor-ew-resize pointer-events-auto"
      @mousedown="onMouseDown"
    />
  </div>
</template>
