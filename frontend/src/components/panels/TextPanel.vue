<script setup>
import { ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { Type, Plus } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const presets = [
  { name: 'Heading',     text: 'Heading',      size: 72, weight: 'bold',   color: '#FFFFFF', bgColor: 'transparent', effect: 'solid', animation: 'none' },
  { name: 'Subtitle',    text: 'Subtitle',     size: 42, weight: '600',    color: '#FFFFFF', bgColor: 'transparent', effect: 'soft', animation: 'none' },
  { name: 'Body',        text: 'Body text',    size: 24, weight: 'normal', color: '#E8E8E8', bgColor: 'transparent', effect: 'soft', animation: 'none' },
  { name: 'Caption',     text: 'Caption',      size: 18, weight: 'normal', color: '#888888', bgColor: 'transparent', effect: 'dim', animation: 'none' },
  { name: 'Callout',     text: 'Callout!',     size: 36, weight: 'bold',   color: '#0F0F0F', bgColor: '#00D4FF', effect: 'badge', animation: 'none' },
  { name: 'Quote',       text: '"Quote"',      size: 30, weight: 'italic', color: '#F59E0B', bgColor: 'transparent', effect: 'accent', animation: 'none' },
  { name: 'Neon',        text: 'NEON',         size: 64, weight: 'bold',   color: '#EC4899', bgColor: 'transparent', effect: 'glow', animation: 'pulse' },
  { name: 'Mono',        text: 'mono.txt',     size: 22, weight: 'normal', color: '#10B981', bgColor: 'transparent', effect: 'mono', animation: 'none' },
  { name: 'Lower Third', text: 'Lower Third',  size: 28, weight: '500',    color: '#FFFFFF', bgColor: 'rgba(0,0,0,0.6)', effect: 'panel', animation: 'none' },
  { name: 'Glow Up',     text: 'Glow Up',      size: 58, weight: 'bold',   color: '#FFFFFF', bgColor: 'transparent', effect: 'glow', animation: 'float' },
  { name: 'Cinematic',   text: 'Cinematic',    size: 52, weight: '700',    color: '#F8FAFC', bgColor: 'rgba(15,23,42,0.72)', effect: 'panel', animation: 'none' },
  { name: 'Sunset',      text: 'Sunset',       size: 48, weight: 'bold',   color: '#FFF7ED', bgColor: 'linear-gradient(135deg, #FB7185, #F59E0B)', effect: 'gradient', animation: 'none' },
  { name: 'Cyber',       text: 'CYBER',        size: 60, weight: '800',    color: '#A78BFA', bgColor: 'rgba(8,15,35,0.75)', effect: 'neon', animation: 'pulse' },
  { name: 'Spotlight',   text: 'SPOTLIGHT',    size: 54, weight: 'bold',   color: '#FFFDF6', bgColor: 'rgba(120,53,15,0.55)', effect: 'spotlight', animation: 'float' },
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
            textShadow: p.effect === 'glow' || p.effect === 'neon' ? '0 0 10px rgba(56, 189, 248, 0.7), 0 0 18px rgba(236, 72, 153, 0.25)' : p.effect === 'spotlight' ? '0 2px 8px rgba(15, 23, 42, 0.85)' : 'none',
            boxShadow: p.effect === 'panel' ? 'inset 0 0 0 1px rgba(255,255,255,0.08)' : 'none',
            filter: p.effect === 'gradient' ? 'drop-shadow(0 4px 10px rgba(251, 146, 60, 0.35))' : 'none',
            animation: p.animation === 'pulse' ? 'textPulse 1.8s ease-in-out infinite' : p.animation === 'float' ? 'textFloat 2.4s ease-in-out infinite' : 'none',
          }"
        >
          {{ p.text }}
        </div>
        <div class="text-[10px] text-text-secondary mt-1">{{ p.name }}</div>
      </button>
    </div>
  </div>
</template>

<style scoped>
@keyframes textPulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.03); }
}

@keyframes textFloat {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-2px); }
}
</style>
