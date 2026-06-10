<script setup>
import { computed } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { usePlayerStore } from '../../stores/playerStore'
import { useProjectStore } from '../../stores/projectStore'
import { Play, Pause, SkipBack, SkipForward, Monitor, Eye, HelpCircle } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const playerStore = usePlayerStore()
const projectStore = useProjectStore()

const previewSrc = computed(() => {
  if (!playerStore.previewImage) return null
  return `data:image/jpeg;base64,${playerStore.previewImage}`
})
</script>

<template>
  <div class="flex-1 flex flex-col bg-zinc-950 min-h-0 overflow-hidden relative border-b border-border">
    <!-- Single Viewer (Clean & Minimal) -->
    <div class="flex-1 flex flex-col p-2 min-h-0 bg-black/60">
      <div class="flex flex-col h-full border border-zinc-800 rounded-lg overflow-hidden bg-zinc-900/60 relative">
        <div class="h-6 px-3 bg-zinc-950 border-b border-zinc-800 flex items-center justify-between text-[10px] text-zinc-400 font-mono">
          <div class="flex items-center gap-1.5">
            <Monitor :size="10" class="text-accent" />
            <span>Fusion Render Viewer</span>
          </div>
          <span class="text-[8px] uppercase px-1 rounded bg-accent/20 text-accent font-bold">Active Compositing</span>
        </div>
        
        <!-- Frame area -->
        <div class="flex-1 flex items-center justify-center p-2 min-h-0 relative bg-zinc-950">
          <img
            v-if="previewSrc"
            :src="previewSrc"
            class="max-w-full max-h-full object-contain rounded bg-black shadow-2xl"
            alt="fusion output"
          />
          <div v-else class="text-zinc-600 text-[10px] flex flex-col items-center gap-2">
            <Eye :size="20" class="opacity-40" />
            <span>Select a clip or shape to preview composition</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Fusion Transport controls -->
    <div class="h-9 px-3 border-t border-border bg-panel/40 flex items-center gap-2 flex-shrink-0">
      <button
        class="p-1 rounded text-text-secondary hover:text-text-primary hover:bg-border/40 transition-colors"
        @click="playerStore.stepBackward()"
        title="Prev Frame"
      >
        <SkipBack :size="12" />
      </button>

      <button
        class="p-1.5 rounded-full bg-accent text-bg hover:bg-accent-hover transition-colors"
        @click="playerStore.togglePlay()"
      >
        <Pause v-if="playerStore.isPlaying" :size="11" />
        <Play v-else :size="11" />
      </button>

      <button
        class="p-1 rounded text-text-secondary hover:text-text-primary hover:bg-border/40 transition-colors"
        @click="playerStore.stepForward()"
        title="Next Frame"
      >
        <SkipForward :size="12" />
      </button>

      <!-- Scrub slider -->
      <div class="flex-1 px-3 flex items-center gap-2">
        <input
          type="range"
          :min="0"
          :max="Math.max(timelineStore.duration, 5)"
          step="0.01"
          :value="timelineStore.currentTime"
          @input="(e) => timelineStore.setCurrentTime(parseFloat(e.target.value))"
          class="w-full h-1 bg-zinc-800 rounded appearance-none cursor-pointer accent-accent"
        />
      </div>

      <div class="font-mono text-[10px] text-zinc-400">
        {{ playerStore.formattedTime }}
      </div>
    </div>
  </div>
</template>
