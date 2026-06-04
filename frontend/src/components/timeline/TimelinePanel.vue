<script setup>
import { computed, ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import { Plus, ZoomIn, ZoomOut, Maximize2, Link2, Scissors, Trash2, Volume2, VolumeX, Lock, Unlock, Video, Music, Type, Smile, Sparkles, Magnet } from 'lucide-vue-next'
import TimelineRuler from './TimelineRuler.vue'
import Track from './Track.vue'
import Playhead from './Playhead.vue'

const timelineStore = useTimelineStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

// Pixels of scrollable timeline area. The container is the entire bottom panel
// minus the toolbar (40px) minus the ruler (28px).
const TRACK_HEIGHT = 56
const TRACK_HEADER_WIDTH = 96

const visibleDuration = computed(() => Math.max(timelineStore.duration + 10, 30))
const totalWidth = computed(() => visibleDuration.value * timelineStore.zoom)

const trackOrder = ['video', 'audio', 'text', 'sticker', 'fx']

const tracksByType = computed(() => {
  const map = {}
  for (const t of timelineStore.tracks) {
    if (!map[t.type]) map[t.type] = []
    map[t.type].push(t)
  }
  return map
})

function addTrackOfType(type) {
  timelineStore.addTrack(type)
  projectStore.markDirty()
  uiStore.addToast(`Added ${type} track`, 'success', 1000)
}

function removeTrack(track) {
  timelineStore.removeTrack(track.id)
  projectStore.markDirty()
}

function toggleMute(track) {
  track.muted = !track.muted
  projectStore.markDirty()
}

function toggleLock(track) {
  track.locked = !track.locked
  projectStore.markDirty()
}

// Wheel-zoom on the timeline
const timelineBody = ref(null)
function onWheel(e) {
  if (e.ctrlKey || e.metaKey) {
    e.preventDefault()
    const factor = e.deltaY < 0 ? 1.15 : 0.87
    timelineStore.zoomBy(factor)
  }
}

// Drop onto empty track rows (auto-create track + add clip)
const emptyRowDragOver = ref(null) // which track type is hovered

function onEmptyRowDragOver(e, type) {
  if (e.dataTransfer.types.includes('application/x-gocut-asset')) {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'copy'
    emptyRowDragOver.value = type
  }
}

function onEmptyRowDragLeave(type) {
  if (emptyRowDragOver.value === type) emptyRowDragOver.value = null
}

function onEmptyRowDrop(e, type) {
  emptyRowDragOver.value = null
  const assetId = e.dataTransfer.getData('application/x-gocut-asset')
  if (!assetId) return
  const asset = projectStore.getAsset(assetId)
  if (!asset) return
  // Compute drop time from X offset relative to this element
  const time = Math.max(0, e.offsetX / timelineStore.zoom)
  // Auto-create the track if it doesn't exist
  let track = timelineStore.getTrackByType(type)
  if (!track) track = timelineStore.addTrack(type)
  timelineStore.addClip({
    assetId: asset.id,
    trackId: track.id,
    startTime: time,
    duration: asset.duration || 3,
  })
  if (asset.type === 'video' && type === 'video') {
    timelineStore.addClip({
      assetId: asset.id,
      trackType: 'audio',
      startTime: time,
      duration: asset.duration || 3,
    })
  }
  projectStore.markDirty()
  timelineStore.setCurrentTime(time)
  uiStore.addToast(`Added to ${type} track`, 'success', 1200)
}

function splitSelectedAtPlayhead() {
  if (timelineStore.selectedClips.length === 0) {
    uiStore.addToast('Select a clip to split', 'warn')
    return
  }
  for (const c of timelineStore.selectedClips) {
    timelineStore.splitClipAt(c.id, timelineStore.currentTime)
  }
  uiStore.addToast('Split', 'success', 1000)
}

function deleteSelected() {
  if (timelineStore.selectedClips.length === 0) return
  timelineStore.removeSelected()
}

const typeIcons = { video: Video, audio: Music, text: Type, sticker: Smile, fx: Sparkles }
</script>

<template>
  <div class="h-64 bg-panel border-t border-border flex flex-col flex-shrink-0 select-none">
    <!-- Toolbar -->
    <div class="h-10 bg-panel border-b border-border flex items-center px-2 gap-1 flex-shrink-0">
      <div class="flex items-center gap-0.5">
        <button
          v-for="(type, i) in trackOrder"
          :key="type"
          class="flex items-center gap-1 px-1.5 py-1 rounded text-[10px] text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
          :title="`Add ${type} track`"
          @click="addTrackOfType(type)"
        >
          <component :is="typeIcons[type]" :size="11" />
          <Plus :size="9" />
        </button>
      </div>

      <div class="h-5 w-px bg-border mx-1" />

      <button
        class="flex items-center gap-1 px-2 py-1 rounded text-[10px] text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="splitSelectedAtPlayhead"
        title="Split (S)"
      >
        <Scissors :size="11" /> Split
      </button>
      <button
        class="flex items-center gap-1 px-2 py-1 rounded text-[10px] text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="deleteSelected"
        title="Delete (Del)"
      >
        <Trash2 :size="11" /> Delete
      </button>

      <div class="h-5 w-px bg-border mx-1" />

      <button
        class="p-1 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="timelineStore.setZoom(timelineStore.zoom - 10)"
        title="Zoom out"
      >
        <ZoomOut :size="12" />
      </button>
      <div class="font-mono text-[10px] text-text-secondary w-12 text-center">{{ Math.round(timelineStore.zoom) }}px/s</div>
      <button
        class="p-1 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="timelineStore.setZoom(timelineStore.zoom + 10)"
        title="Zoom in"
      >
        <ZoomIn :size="12" />
      </button>
      <button
        class="p-1 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="timelineStore.setZoom(50)"
        title="Zoom to fit"
      >
        <Maximize2 :size="12" />
      </button>

      <div class="h-5 w-px bg-border mx-1" />

      <button
        class="p-1 rounded transition-colors"
        :class="timelineStore.snapEnabled ? 'text-accent bg-accent/10' : 'text-text-secondary hover:text-text-primary hover:bg-border'"
        @click="timelineStore.snapEnabled = !timelineStore.snapEnabled"
        title="Snap (S)"
      >
        <Magnet :size="12" />
      </button>
      <button
        class="p-1 rounded transition-colors"
        :class="timelineStore.rippleDelete ? 'text-accent bg-accent/10' : 'text-text-secondary hover:text-text-primary hover:bg-border'"
        @click="timelineStore.rippleDelete = !timelineStore.rippleDelete"
        title="Ripple delete"
      >
        <Link2 :size="12" />
      </button>

      <div class="flex-1" />

      <div class="text-[10px] text-text-secondary font-mono">
        {{ timelineStore.clips.length }} clip{{ timelineStore.clips.length === 1 ? '' : 's' }}
        · {{ timelineStore.tracks.length }} track{{ timelineStore.tracks.length === 1 ? '' : 's' }}
      </div>
    </div>

    <!-- Timeline body: header column + scrollable content -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Track headers column -->
      <div class="w-24 bg-bg/40 border-r border-border flex-shrink-0 flex flex-col">
        <div class="h-7 border-b border-border" />
        <div
          v-for="type in trackOrder"
          :key="type"
          class="flex items-center gap-1 px-2 border-b border-border text-[10px] text-text-secondary"
          :style="{ minHeight: TRACK_HEIGHT + 'px' }"
        >
          <component :is="typeIcons[type]" :size="11" />
          <span class="capitalize">{{ type }}</span>
          <span class="ml-auto text-text-secondary/60">{{ tracksByType[type]?.length || 0 }}</span>
        </div>
      </div>

      <!-- Scrollable content: ruler + tracks -->
      <div
        ref="timelineBody"
        class="timeline-content flex-1 overflow-x-auto overflow-y-hidden"
        @wheel="onWheel"
      >
        <div
          class="relative"
          :style="{ width: totalWidth + 'px', minWidth: '100%' }"
        >
          <TimelineRuler :duration="visibleDuration" />

          <div class="relative">
            <div
              v-for="type in trackOrder"
              :key="type"
              :style="{ height: TRACK_HEIGHT + 'px' }"
              class="relative border-b border-border"
            >
              <template v-if="tracksByType[type]?.length">
                <Track
                  v-for="track in tracksByType[type]"
                  :key="track.id"
                  :track="track"
                  :track-meta="timelineStore.TRACK_TYPES[type]"
                  :duration="visibleDuration"
                />
              </template>
              <div
                v-else
                class="absolute inset-0 flex items-center justify-center text-[10px] text-text-secondary/40 italic transition-colors"
                :class="emptyRowDragOver === type ? 'bg-accent/10 text-accent' : ''"
                @dragover="(e) => onEmptyRowDragOver(e, type)"
                @dragleave="onEmptyRowDragLeave(type)"
                @drop="(e) => onEmptyRowDrop(e, type)"
              >
                {{ emptyRowDragOver === type ? `Drop here to add to ${type} track` : `${type} track — drag a clip here or use + above` }}
              </div>
            </div>
            <Playhead :duration="visibleDuration" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
