<script setup>
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { ArrowRightLeft } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const transitions = [
  { name: 'None',      type: 'none',      duration: 0,   preview: null },
  { name: 'Fade',      type: 'fade',      duration: 0.5, preview: 'fade' },
  { name: 'Dissolve',  type: 'dissolve',  duration: 0.5, preview: 'dissolve' },
  { name: 'Wipe L',    type: 'wipeleft',  duration: 0.5, preview: 'wipe-l' },
  { name: 'Wipe R',    type: 'wiperight', duration: 0.5, preview: 'wipe-r' },
  { name: 'Slide L',   type: 'slideleft', duration: 0.5, preview: 'slide-l' },
  { name: 'Slide R',   type: 'slideright',duration: 0.5, preview: 'slide-r' },
  { name: 'Zoom In',   type: 'zoomin',    duration: 0.5, preview: 'zoom' },
  { name: 'Flip H',    type: 'hflip',     duration: 0.5, preview: 'flip' },
  { name: 'Circle',    type: 'circleopen',duration: 0.5, preview: 'circle' },
  { name: 'Pixelize',  type: 'pixelize',  duration: 0.5, preview: 'pixel' },
  { name: 'Blur',      type: 'blur',      duration: 0.5, preview: 'blur' },
]

function applyTransition(t) {
  if (timelineStore.selectedClips.length === 0) {
    uiStore.addToast('Select a clip first', 'warn')
    return
  }
  const clip = timelineStore.selectedClips[0]
  if (t.type === 'none') {
    timelineStore.updateClip(clip.id, { transition: null })
  } else {
    timelineStore.updateClip(clip.id, {
      transition: { type: t.type, duration: t.duration }
    })
  }
  uiStore.addToast(`Applied ${t.name}`, 'success', 1200)
}

function transitionStyle(t) {
  // Render a small CSS-based preview of the transition
  if (!t.preview) return {}
  const base = 'linear-gradient(135deg, #00D4FF 0%, #00D4FF 50%, #EC4899 50%, #EC4899 100%)'
  const map = {
    'fade':       { background: 'linear-gradient(90deg, #00D4FF, #EC4899)' },
    'dissolve':   { background: 'linear-gradient(90deg, #00D4FF, #EC4899)' },
    'wipe-l':     { background: 'linear-gradient(90deg, #EC4899 0%, #00D4FF 100%)' },
    'wipe-r':     { background: 'linear-gradient(90deg, #00D4FF 0%, #EC4899 100%)' },
    'slide-l':    { background: 'linear-gradient(90deg, #EC4899 0%, #00D4FF 100%)' },
    'slide-r':    { background: 'linear-gradient(90deg, #00D4FF 0%, #EC4899 100%)' },
    'zoom':       { background: 'radial-gradient(circle, #00D4FF, #EC4899)' },
    'flip':       { background: 'linear-gradient(90deg, #00D4FF 0%, #00D4FF 50%, #EC4899 50%, #EC4899 100%)' },
    'circle':     { background: 'radial-gradient(circle at center, #00D4FF 0%, #00D4FF 30%, #EC4899 30%)' },
    'pixel':      { backgroundImage: 'repeating-linear-gradient(45deg, #00D4FF 0 6px, #EC4899 6px 12px)' },
    'blur':       { background: 'linear-gradient(90deg, #00D4FF, #EC4899)', filter: 'blur(2px)' },
  }
  return map[t.preview] || { background: base }
}
</script>

<template>
  <div class="p-2 flex flex-col gap-2">
    <div class="grid grid-cols-2 gap-1.5">
      <button
        v-for="(t, i) in transitions"
        :key="i"
        class="rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors p-1.5"
        @click="applyTransition(t)"
      >
        <div
          class="aspect-video rounded mb-1 flex items-center justify-center"
          :style="transitionStyle(t)"
        >
          <ArrowRightLeft v-if="t.preview" :size="14" class="text-white drop-shadow" />
          <span v-else class="text-[10px] text-text-secondary">—</span>
        </div>
        <div class="text-[10px] text-text-primary text-center">{{ t.name }}</div>
      </button>
    </div>
  </div>
</template>
