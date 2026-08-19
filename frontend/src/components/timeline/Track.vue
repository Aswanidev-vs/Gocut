<script setup>
import { computed, ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'
import { useHistoryStore } from '../../stores/historyStore'
import Clip from './Clip.vue'
import { Volume2, VolumeX, Lock, Unlock, Trash2, Plus, Video, Music, Type, Smile, Sparkles } from 'lucide-vue-next'

const props = defineProps({
  track: { type: Object, required: true },
  trackMeta: { type: Object, default: () => ({}) },
  duration: { type: Number, default: 60 },
})

const timelineStore = useTimelineStore()
const projectStore = useProjectStore()

const typeIcons = { video: Video, audio: Music, text: Type, sticker: Smile, fx: Sparkles }
const trackIcon = computed(() => typeIcons[props.track.type] || Video)

const trackClips = computed(() => {
  return timelineStore.clips
    .filter(c => c.trackId === props.track.id)
    .map(c => ({
      ...c,
      isSelected: timelineStore.selectedClipIds.includes(c.id),
    }))
})

function toggleMute() {
  props.track.muted = !props.track.muted
  projectStore.markDirty()
}

function toggleLock() {
  props.track.locked = !props.track.locked
  projectStore.markDirty()
}

function removeTrack() {
  timelineStore.removeTrack(props.track.id)
  projectStore.markDirty()
}

const TRACK_TYPE_FOR_NEW = {
  video: 'video',
  audio: 'audio',
  text: 'text',
  sticker: 'sticker',
  fx: 'fx',
}

function addAnotherTrackOfSameType() {
  timelineStore.addTrack(props.track.type)
  projectStore.markDirty()
}

const isDragOver = ref(false)

function onDragOver(e) {
  // Only accept if dragging an asset
  if (e.dataTransfer.types.includes('application/x-gocut-asset')) {
    e.dataTransfer.dropEffect = 'copy'
    isDragOver.value = true
  }
}

function onDragLeave() {
  isDragOver.value = false
}

function onDrop(e) {
  e.preventDefault()
  isDragOver.value = false
  const assetId = e.dataTransfer.getData('application/x-gocut-asset')
  if (!assetId) return
  
  const asset = projectStore.getAsset(assetId)
  if (!asset) return

  // Calculate drop time from clientX relative to the scrollable timeline container.
  // e.offsetX can be undefined in some WebViews (Wails/WebView2 DragEvent) or
  // relative to a child element, so we always compute from clientX + scrollLeft.
  const scrollContainer = document.querySelector('.timeline-content')
  const rect = scrollContainer ? scrollContainer.getBoundingClientRect() : e.currentTarget.getBoundingClientRect()
  const scrollLeft = scrollContainer ? scrollContainer.scrollLeft : 0
  const px = Math.max(0, e.clientX - rect.left + scrollLeft)
  const time = px / timelineStore.zoom
  
  timelineStore.addClip({
    assetId: asset.id,
    trackId: props.track.id,
    startTime: time,
    duration: asset.duration || 3
  })
  if (asset.type === 'video' && props.track.type === 'video') {
    timelineStore.addClip({
      assetId: asset.id,
      trackType: 'audio',
      startTime: time,
      duration: asset.duration || 3
    })
  }
  timelineStore.setCurrentTime(time)
  projectStore.markDirty()
}

// ---- marquee drag-select on empty track area ----
const marquee = ref(null) // { startPx } | null

function clientXToTime(clientX) {
  const scrollContainer = document.querySelector('.timeline-content')
  const rect = scrollContainer ? scrollContainer.getBoundingClientRect() : null
  if (!rect) return null
  const scrollLeft = scrollContainer.scrollLeft || 0
  return Math.max(0, clientX - rect.left + scrollLeft) / timelineStore.zoom
}

function onTrackMouseDown(e) {
  if (e.button !== 0) return
  // Only act on the empty track area itself (clicks on clips
  // stopPropagation and never reach here).
  if (e.target !== e.currentTarget) return
  e.preventDefault()
  // Clicking empty track area clears selection immediately (drag then
  // builds a new marquee selection from scratch). Remember the previous
  // selection so a plain click that changes nothing doesn't dirty the
  // project or push a redundant history checkpoint.
  const selectionBefore = [...timelineStore.selectedClipIds].sort()
  timelineStore.clearSelection()
  const t = clientXToTime(e.clientX)
  if (t === null) return
  const startPx = Math.max(0, t * timelineStore.zoom)
  marquee.value = { startPx }
  let hadMove = false // distinguish click-to-deselect from a real drag

  const updateMarquee = (mv) => {
    hadMove = true
    const mt = clientXToTime(mv.clientX)
    if (mt === null || !marquee.value) return
    const curPx = Math.max(0, mt * timelineStore.zoom)
    const leftPx = Math.min(startPx, curPx)
    const rightPx = Math.max(startPx, curPx)
    marquee.value = { startPx, leftPx, rightPx }
    // Live-update the selection while dragging.
    const startTime = leftPx / timelineStore.zoom
    const endTime = rightPx / timelineStore.zoom
    const hits = timelineStore.clips
      .filter(c => c.trackId === props.track.id)
      .filter(c => c.startTime < endTime && c.startTime + c.duration > startTime)
    timelineStore.selectedClipIds = hits.map(c => c.id)
  }
  const endMarquee = () => {
    document.removeEventListener('mousemove', updateMarquee)
    document.removeEventListener('mouseup', endMarquee)
    marquee.value = null
    // Commit once on mouseup (dirty flag + history checkpoint) only if this
    // was a real drag AND the selection actually changed.
    const selectionAfter = [...timelineStore.selectedClipIds].sort()
    const changed = selectionAfter.length !== selectionBefore.length ||
      selectionAfter.some((id, i) => id !== selectionBefore[i])
    if (hadMove && changed) {
      projectStore.markDirty()
      const hs = useHistoryStore()
      if (hs && typeof hs.pushSnapshot === 'function') hs.pushSnapshot()
    }
  }
  document.addEventListener('mousemove', updateMarquee)
  document.addEventListener('mouseup', endMarquee)
}

const marqueeStyle = computed(() => {
  if (!marquee.value || marquee.value.leftPx === undefined) return {}
  return {
    left: marquee.value.leftPx + 'px',
    width: Math.max(1, (marquee.value.rightPx || 0) - marquee.value.leftPx) + 'px',
  }
})
</script>

<template>
  <div
    v-if="track"
    class="h-14 relative"
    :class="{ 'bg-accent/10 border-accent/50 border-t border-b': isDragOver }"
    @mousedown="onTrackMouseDown"
    @dragover.prevent="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <!-- Track header (the label/controls sit in a sibling element in the parent
         layout, not here, so the timeline area itself is fully scrollable). -->
    <Clip
      v-for="clip in trackClips"
      :key="clip.id"
      :clip="clip"
      :zoom="timelineStore.zoom"
      :track-type="track.type"
      :track-color="trackMeta.color"
      :selected="clip.isSelected"
    />

    <!-- Marquee drag-select rectangle -->
    <div
      v-if="marquee && marquee.leftPx !== undefined"
      class="absolute top-1 bottom-1 z-20 pointer-events-none border border-accent bg-accent/10 rounded-sm"
      :style="marqueeStyle"
    />
  </div>
</template>
