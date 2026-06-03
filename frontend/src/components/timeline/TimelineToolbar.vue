<script setup>
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import {
  Plus,
  ZoomIn,
  ZoomOut,
  Maximize,
  Link,
  Scissors,
  Trash2,
} from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()
</script>

<template>
  <div class="h-10 bg-panel border-b border-border flex items-center px-2 gap-2">
    <button
      class="flex items-center gap-1 px-2 py-1 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="timelineStore.initializeDefaultTracks()"
      title="Add Track"
    >
      <Plus :size="14" />
      Track
    </button>
    <div class="h-5 w-px bg-border" />
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="timelineStore.setZoom(timelineStore.zoom - 10)"
      title="Zoom Out"
    >
      <ZoomOut :size="14" />
    </button>
    <span class="text-xs text-text-secondary w-12 text-center font-mono">
      {{ timelineStore.zoom }}
    </span>
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="timelineStore.setZoom(timelineStore.zoom + 10)"
      title="Zoom In"
    >
      <ZoomIn :size="14" />
    </button>
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="timelineStore.setZoom(50)"
      title="Zoom to Fit"
    >
      <Maximize :size="14" />
    </button>
    <div class="h-5 w-px bg-border" />
    <button
      class="p-1.5 rounded transition-colors"
      :class="timelineStore.snapEnabled
        ? 'bg-accent/10 text-accent'
        : 'text-text-secondary hover:text-text-primary hover:bg-border'"
      @click="timelineStore.snapEnabled = !timelineStore.snapEnabled"
      title="Snap (S)"
    >
      <Link :size="14" />
    </button>
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      title="Razor Tool"
    >
      <Scissors :size="14" />
    </button>
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      title="Ripple Delete"
    >
      <Trash2 :size="14" />
    </button>
  </div>
</template>
