<script setup>
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { Smile, Plus } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

// SVG stickers rendered inline so we don't need to ship PNG assets.
const stickers = [
  { name: 'Heart',    svg: '<path d="M50 85 L20 55 a18 18 0 0 1 30 -10 a18 18 0 0 1 30 10 Z" fill="#EC4899" />' },
  { name: 'Star',     svg: '<polygon points="50,15 61,40 88,40 66,57 75,84 50,67 25,84 34,57 12,40 39,40" fill="#F59E0B" />' },
  { name: 'Thumb',    svg: '<path d="M30 50 h15 v30 h-15 z M45 80 v-30 l10 -25 a8 8 0 0 1 8 8 v15 h15 a8 8 0 0 1 8 10 l-5 22 a8 8 0 0 1 -8 6 h-28 z" fill="#10B981" />' },
  { name: 'Sparkle',  svg: '<path d="M50 10 L55 40 L85 45 L55 50 L50 80 L45 50 L15 45 L45 40 Z" fill="#00D4FF" />' },
  { name: 'Check',    svg: '<path d="M20 50 L40 70 L80 30" stroke="#10B981" stroke-width="12" fill="none" stroke-linecap="round" stroke-linejoin="round" />' },
  { name: 'X',        svg: '<path d="M25 25 L75 75 M75 25 L25 75" stroke="#EF4444" stroke-width="12" stroke-linecap="round" />' },
  { name: 'Arrow',    svg: '<path d="M10 50 L70 50 L60 35 M70 50 L60 65" stroke="#F59E0B" stroke-width="8" fill="none" stroke-linecap="round" stroke-linejoin="round" />' },
  { name: 'Circle',   svg: '<circle cx="50" cy="50" r="35" fill="#8B5CF6" />' },
  { name: 'Triangle', svg: '<polygon points="50,15 88,80 12,80" fill="#F59E0B" />' },
  { name: 'Square',   svg: '<rect x="20" y="20" width="60" height="60" fill="#3B82F6" />' },
  { name: 'Lightning',svg: '<polygon points="55,10 25,55 45,55 35,90 75,40 55,40 65,10" fill="#FACC15" />' },
  { name: 'Crown',    svg: '<path d="M15 70 L25 30 L40 55 L50 20 L60 55 L75 30 L85 70 Z" fill="#F59E0B" />' },
  { name: 'Fire',     svg: '<path d="M50 90 c-20 0 -25 -15 -10 -30 c-2 -15 5 -25 10 -30 c5 5 12 15 10 30 c15 15 10 30 -10 30 z" fill="#EF4444" />' },
  { name: 'Eye',      svg: '<ellipse cx="50" cy="50" rx="35" ry="20" fill="white" stroke="#888" stroke-width="3" /><circle cx="50" cy="50" r="10" fill="#0F0F0F" />' },
  { name: 'Bubble',   svg: '<path d="M15 20 h70 a8 8 0 0 1 8 8 v30 a8 8 0 0 1 -8 8 h-50 l-15 12 v-12 a8 8 0 0 1 -8 -8 v-30 a8 8 0 0 1 8 -8 z" fill="#00D4FF" />' },
  { name: '100',      svg: '<text x="50" y="62" text-anchor="middle" font-size="38" font-weight="900" fill="#FACC15">100</text>' },
]

function addSticker(s) {
  const t = timelineStore.getTrackByType('sticker') || timelineStore.addTrack('sticker')
  const clip = timelineStore.addClip({
    assetId: '',
    trackId: t.id,
    trackType: 'sticker',
    startTime: timelineStore.currentTime,
    duration: 3,
    stickerProps: {
      x: 0, y: 0, width: 0.2, height: 0.2, rotation: 0, opacity: 1,
      flipH: false, flipV: false, svg: s.svg, name: s.name,
    },
  })
  timelineStore.selectClip(clip.id)
  uiStore.addToast(`Added ${s.name} sticker`, 'success', 1500)
}
</script>

<template>
  <div class="p-2 flex flex-col gap-2">
    <div class="grid grid-cols-3 gap-1.5">
      <button
        v-for="(s, i) in stickers"
        :key="i"
        class="aspect-square rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors flex items-center justify-center group"
        @click="addSticker(s)"
        :title="s.name"
      >
        <svg viewBox="0 0 100 100" class="w-10 h-10" v-html="s.svg" />
      </button>
    </div>
  </div>
</template>
