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
</script>

<template>
  <div
    v-if="track"
    class="h-14 relative"
    @click.self="timelineStore.clearSelection()"
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
