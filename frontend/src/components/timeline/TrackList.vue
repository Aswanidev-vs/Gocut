<script setup>
import { computed } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import Track from './Track.vue'

const props = defineProps({
  duration: { type: Number, default: 60 },
})

const timelineStore = useTimelineStore()

const trackTypeMeta = computed(() => {
  const types = ['video', 'audio', 'text', 'sticker', 'fx']
  return types.map(type => ({
    type,
    ...timelineStore.TRACK_TYPES[type],
  }))
})
</script>

<template>
  <div class="flex flex-col">
    <div
      v-for="trackMeta in trackTypeMeta"
      :key="trackMeta.type"
      class="flex border-b border-border"
    >
      <div
        class="w-10 h-16 flex items-center justify-center border-r border-border flex-shrink-0"
        :style="{ backgroundColor: trackMeta.color + '15' }"
      >
        <span class="text-xs font-bold" :style="{ color: trackMeta.color }">
          {{ trackMeta.label }}
        </span>
      </div>
      <div class="flex-1 h-16 relative" :style="{ width: `${duration * timelineStore.zoom}px` }">
        <Track
          v-if="timelineStore.tracks.find(t => t.type === trackMeta.type)"
          :track-type="trackMeta.type"
          :track-meta="trackMeta"
          :duration="duration"
        />
        <div
          v-else
          class="h-full border border-dashed border-border rounded flex items-center justify-center text-text-secondary text-xs"
        >
          No track
        </div>
      </div>
    </div>
  </div>
</template>
