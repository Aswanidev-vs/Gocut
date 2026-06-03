<script setup>
import { computed, ref } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'
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
  isDragOver.value = false
  const assetId = e.dataTransfer.getData('application/x-gocut-asset')
  if (!assetId) return
  
  const asset = projectStore.getAsset(assetId)
  if (!asset) return

  // Calculate drop time based on X coordinate
  const rect = e.currentTarget.getBoundingClientRect()
  const offsetX = Math.max(0, e.clientX - rect.left)
  
  // Account for timeline scrolling (scrollX is managed by TimelinePanel/parent, 
  // but timelineStore.zoom is px per second. 
  // Wait, if the Track is inside a scrollable container, e.clientX is relative to the viewport.
  // We need to use offsetX on the target element itself (which stretches to totalWidth)
  const time = (e.offsetX || offsetX) / timelineStore.zoom
  
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
}
</script>

<template>
  <div
    v-if="track"
    class="h-14 relative"
    :class="{ 'bg-accent/10 border-accent/50 border-t border-b': isDragOver }"
    @click.self="timelineStore.clearSelection()"
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
  </div>
</template>
