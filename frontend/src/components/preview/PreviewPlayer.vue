<script setup>
import { ref, computed, onMounted, onUnmounted, watch, watchEffect } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { usePlayerStore } from '../../stores/playerStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { Play, Pause, SkipBack, SkipForward, Maximize2, Volume2, VolumeX, Repeat, Loader2, ImageOff } from 'lucide-vue-next'

const projectStore = useProjectStore()
const playerStore = usePlayerStore()
const timelineStore = useTimelineStore()

const isFullscreen = ref(false)
const playInterval = ref(null)

// Playback ticker: when playing, advance currentTime every animation frame.
function startTicker() {
  stopTicker()
  let last = performance.now()
  const tick = (now) => {
    const dt = (now - last) / 1000
    last = now
    if (playerStore.isPlaying) {
      const next = timelineStore.currentTime + dt
      if (next >= Math.max(timelineStore.duration, 0.5)) {
        if (playerStore.loop) {
          timelineStore.setCurrentTime(0)
        } else {
          playerStore.pause()
          return
        }
      } else {
        timelineStore.setCurrentTime(next)
      }
    }
    playInterval.value = requestAnimationFrame(tick)
  }
  playInterval.value = requestAnimationFrame(tick)
}

function stopTicker() {
  if (playInterval.value) {
    cancelAnimationFrame(playInterval.value)
    playInterval.value = null
  }
}

onMounted(() => {
  startTicker()
  // Kick a first preview frame.
  setTimeout(() => playerStore.refreshPreview().catch(() => {}), 100)
})

onUnmounted(() => {
  stopTicker()
})

// Refresh the preview frame whenever the current time changes (but throttled
// to avoid hammering ffmpeg).
let refreshTimeout = null
watch(() => timelineStore.currentTime, () => {
  if (refreshTimeout) clearTimeout(refreshTimeout)
  refreshTimeout = setTimeout(() => {
    playerStore.refreshPreview().catch(() => {})
  }, 80)
})

watch(() => projectStore.project?.id, () => {
  setTimeout(() => playerStore.refreshPreview().catch(() => {}), 100)
})

const aspectStyle = computed(() => {
  const res = projectStore.project?.resolution || { width: 1920, height: 1080 }
  return { aspectRatio: `${res.width} / ${res.height}` }
})

const projectSize = computed(() => {
  const r = projectStore.project?.resolution || { width: 1920, height: 1080 }
  return `${r.width} × ${r.height} · ${projectStore.project?.fps || 30}fps`
})

const previewSrc = computed(() => {
  if (!playerStore.previewImage) return null
  // The Go side returns a bare base64 string; we wrap as JPEG.
  return `data:image/jpeg;base64,${playerStore.previewImage}`
})

// Konva-style overlay: render text and sticker clips on top of the video frame.
const overlayItems = computed(() => {
  const t = timelineStore.currentTime
  const items = []
  for (const c of timelineStore.clips) {
    if (t < c.startTime || t >= c.startTime + c.duration) continue
    if (c.textProps) items.push({ kind: 'text', clip: c, props: c.textProps })
    if (c.stickerProps) items.push({ kind: 'sticker', clip: c, props: c.stickerProps })
  }
  return items
})

function toggleFullscreen() {
  const el = document.querySelector('.preview-frame')
  if (!el) return
  if (!document.fullscreenElement) {
    el.requestFullscreen?.().catch(() => {})
    isFullscreen.value = true
  } else {
    document.exitFullscreen?.()
    isFullscreen.value = false
  }
}

function toggleMute() {
  playerStore.setVolume(playerStore.volume > 0 ? 0 : 1.0)
}

function onScrub(e) {
  const value = parseFloat(e.target.value)
  if (!isNaN(value)) {
    timelineStore.setCurrentTime(value)
  }
}

const scrubMax = computed(() => Math.max(timelineStore.duration, 5))

const visibleClips = computed(() => {
  const t = timelineStore.currentTime
  return timelineStore.clips.filter(c => t >= c.startTime && t < c.startTime + c.duration)
})
</script>

<template>
  <div class="flex-1 flex flex-col bg-bg/50 min-h-0 overflow-hidden">
    <!-- Aspect-ratio preview frame -->
    <div class="flex-1 flex items-center justify-center p-4 min-h-0">
      <div
        class="preview-frame relative bg-black rounded-md overflow-hidden shadow-2xl shadow-black/50 ring-1 ring-border max-h-full"
        :style="{ ...aspectStyle, maxWidth: '100%', maxHeight: '100%', width: '100%' }"
      >
        <!-- Frame from ffmpeg -->
        <img
          v-if="previewSrc"
          :src="previewSrc"
          class="absolute inset-0 w-full h-full object-contain bg-black"
          alt="preview"
        />
        <div
          v-else
          class="absolute inset-0 flex flex-col items-center justify-center text-text-secondary text-sm gap-2"
        >
          <ImageOff :size="32" class="opacity-40" />
          <div v-if="playerStore.isPreviewLoading" class="flex items-center gap-2 text-xs">
            <Loader2 :size="14" class="animate-spin" />
            Loading frame…
          </div>
          <div v-else-if="playerStore.previewError" class="text-xs text-red-400 max-w-md text-center px-4">
            {{ playerStore.previewError }}
          </div>
          <div v-else class="text-xs">
            {{ (projectStore.project?.assets || []).length === 0 ? 'Import media to see preview' : 'No clip at playhead' }}
          </div>
        </div>

        <!-- Text overlay (no Konva, just CSS) -->
        <div class="absolute inset-0 pointer-events-none">
          <div
            v-for="(it, idx) in overlayItems"
            :key="idx"
            class="absolute"
            :style="{
              left: '50%',
              top: '50%',
              transform: 'translate(-50%, -50%)',
            }"
          >
            <div
              v-if="it.kind === 'text'"
              :style="{
                color: it.props.color || '#FFFFFF',
                background: it.props.bgColor && it.props.bgColor !== 'transparent' ? it.props.bgColor : 'transparent',
                fontSize: (it.props.fontSize || 32) * 0.5 + 'px',
                fontWeight: it.props.bold ? 'bold' : it.props.fontWeight || 'normal',
                fontStyle: it.props.italic ? 'italic' : 'normal',
                textDecoration: it.props.underline ? 'underline' : 'none',
                padding: it.props.bgColor && it.props.bgColor !== 'transparent' ? '4px 12px' : '0',
                borderRadius: (it.props.bgBorderRadius || 0) + 'px',
                fontFamily: it.props.fontFamily || 'DM Sans',
                whiteSpace: 'pre',
                textAlign: it.props.align || 'center',
              }"
            >
              {{ it.props.text }}
            </div>
            <div
              v-else-if="it.kind === 'sticker'"
              :style="{
                opacity: it.props.opacity ?? 1,
                transform: `scale(${it.props.width || 0.2}) rotate(${it.props.rotation || 0}deg)`,
                transformOrigin: 'center',
              }"
              v-html="it.props.svg"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Transport controls -->
    <div class="px-3 py-2 flex items-center gap-2 border-t border-border bg-panel flex-shrink-0">
      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="playerStore.stepBackward()"
        title="Previous frame (←)"
      >
        <SkipBack :size="14" />
      </button>

      <button
        class="p-2 rounded-full bg-accent text-bg hover:bg-accent-hover transition-colors"
        @click="playerStore.togglePlay()"
        :title="playerStore.isPlaying ? 'Pause (Space)' : 'Play (Space)'"
      >
        <Pause v-if="playerStore.isPlaying" :size="14" />
        <Play v-else :size="14" />
      </button>

      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="playerStore.stepForward()"
        title="Next frame (→)"
      >
        <SkipForward :size="14" />
      </button>

      <div class="font-mono text-xs text-text-secondary w-32 text-center">
        {{ playerStore.formattedTime }} / {{ timelineStore.duration ? playerStore.formatTimecode(timelineStore.duration) : '00:00:00.000' }}
      </div>

      <div class="flex-1 px-2">
        <input
          type="range"
          :min="0"
          :max="scrubMax"
          step="0.01"
          :value="timelineStore.currentTime"
          @input="onScrub"
          class="w-full h-1 bg-border rounded appearance-none cursor-pointer accent-accent"
        />
      </div>

      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="toggleMute"
        :title="playerStore.volume > 0 ? 'Mute' : 'Unmute'"
      >
        <Volume2 v-if="playerStore.volume > 0" :size="14" />
        <VolumeX v-else :size="14" />
      </button>

      <button
        class="p-1.5 rounded transition-colors"
        :class="playerStore.loop ? 'text-accent' : 'text-text-secondary hover:text-text-primary hover:bg-border'"
        @click="playerStore.toggleLoop()"
        title="Loop"
      >
        <Repeat :size="14" />
      </button>

      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="toggleFullscreen"
        title="Fullscreen"
      >
        <Maximize2 :size="14" />
      </button>

      <div class="text-[10px] text-text-secondary font-mono hidden md:block">{{ projectSize }}</div>
    </div>
  </div>
</template>
