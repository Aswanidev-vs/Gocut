<script setup>
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { FILTER_PRESETS, cssFilterForGrade } from '../../lib/filterPresets'
import { Sparkles } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

// Layered base scene that each preset's filter is applied over, so every
// thumbnail reads as a differently-graded version of the same footage.
const BASE_SCENE = [
  'linear-gradient(135deg, rgba(56,189,248,0.85) 0%, rgba(139,92,246,0.7) 55%, rgba(236,72,153,0.8) 100%)',
  'radial-gradient(circle at 30% 25%, rgba(250,204,21,0.55) 0%, transparent 45%)',
  'linear-gradient(180deg, rgba(15,23,42,0.35) 0%, rgba(15,23,42,0) 60%)',
].join(', ')

function presetFilter(preset) {
  const { filter, opacity } = cssFilterForGrade(preset.grade)
  return {
    backgroundImage: BASE_SCENE,
    filter: Array.isArray(filter) ? filter.join(' ') : filter,
    opacity,
  }
}

function applyPreset(preset) {
  const ids = timelineStore.selectedClipIds
  if (!ids || ids.length === 0) {
    uiStore.addToast('Select a clip first', 'warn')
    return
  }
  ids.forEach((id) => {
    timelineStore.updateClipColor(id, preset.grade)
  })
  uiStore.addToast(`Applied ${preset.name} to ${ids.length} clip${ids.length > 1 ? 's' : ''}`, 'success', 1200)
}
</script>

<template>
  <div class="p-2 flex flex-col gap-2">
    <div class="grid grid-cols-2 gap-1.5">
      <button
        v-for="(p, i) in FILTER_PRESETS"
        :key="i"
        class="rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors p-1.5"
        @click="applyPreset(p)"
      >
        <div
          class="aspect-video rounded mb-1 flex items-center justify-center"
          :style="presetFilter(p)"
        >
          <Sparkles :size="14" class="text-white drop-shadow" />
        </div>
        <div class="text-[10px] text-text-primary text-center">{{ p.name }}</div>
      </button>
    </div>
  </div>
</template>
