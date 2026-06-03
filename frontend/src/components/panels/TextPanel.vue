<script setup>
import { ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { Type, Plus } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const presets = [
  { name: 'Heading',   text: 'Heading',   size: 72, weight: 'bold',   color: '#FFFFFF', bgColor: 'transparent' },
  { name: 'Subtitle',  text: 'Subtitle',  size: 42, weight: '600',    color: '#FFFFFF', bgColor: 'transparent' },
  { name: 'Body',      text: 'Body text', size: 24, weight: 'normal', color: '#E8E8E8', bgColor: 'transparent' },
  { name: 'Caption',   text: 'Caption',   size: 18, weight: 'normal', color: '#888888', bgColor: 'transparent' },
  { name: 'Callout',   text: 'Callout!',  size: 36, weight: 'bold',   color: '#0F0F0F', bgColor: '#00D4FF' },
  { name: 'Quote',     text: '"Quote"',   size: 30, weight: 'italic', color: '#F59E0B', bgColor: 'transparent' },
  { name: 'Neon',      text: 'NEON',      size: 64, weight: 'bold',   color: '#EC4899', bgColor: 'transparent' },
  { name: 'Mono',      text: 'mono.txt',  size: 22, weight: 'normal', color: '#10B981', bgColor: 'transparent' },
  { name: 'Subtitle2', text: 'Lower Third', size: 28, weight: '500', color: '#FFFFFF', bgColor: 'rgba(0,0,0,0.6)' },
]

function addPreset(p) {
  const t = timelineStore.getTrackByType('text') || timelineStore.addTrack('text')
  const clip = timelineStore.addClip({
    assetId: '',
    trackId: t.id,
    trackType: 'text',
    startTime: timelineStore.currentTime,
    duration: 3,
    textProps: {
      text: p.text,
      fontFamily: 'DM Sans',
      fontSize: p.size,
      bold: p.weight === 'bold',
      italic: p.weight === 'italic',
      color: p.color,
      bgColor: p.bgColor,
      align: 'center',
    },
  })
  timelineStore.selectClip(clip.id)
  uiStore.addToast(`Added ${p.name} text`, 'success', 1500)
}
</script>

<template>
  <div class="p-2 flex flex-col gap-2">
    <button
      class="flex items-center justify-center gap-1.5 w-full py-2 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors"
      @click="addPreset({ name: 'Heading', text: 'New Text', size: 48, weight: 'bold', color: '#FFFFFF', bgColor: 'transparent' })"
    >
      <Plus :size="12" /> Add Text
    </button>

    <div class="grid grid-cols-1 gap-1.5">
      <button
        v-for="(p, i) in presets"
        :key="i"
        class="rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors p-2 text-left"
        @dblclick="addPreset(p)"
        @click="addPreset(p)"
      >
        <div
          class="truncate"
          :style="{
            color: p.color,
            background: p.bgColor,
            fontSize: Math.min(p.size / 2.5, 24) + 'px',
            fontWeight: p.weight,
            fontStyle: p.weight === 'italic' ? 'italic' : 'normal',
            padding: p.bgColor === 'transparent' ? '0' : '2px 8px',
            borderRadius: p.bgColor === 'transparent' ? '0' : '4px',
            lineHeight: 1.1,
          }"
        >
          {{ p.text }}
        </div>
        <div class="text-[10px] text-text-secondary mt-1">{{ p.name }}</div>
      </button>
    </div>
  </div>
</template>
