<script setup>
  import { useTimelineStore } from '../../stores/timelineStore'
  import { usePlayerStore } from '../../stores/playerStore'

  const timelineStore = useTimelineStore()
  const playerStore = usePlayerStore()
</script>

<template>
  <div class="relative w-full h-full">
    <div
      class="absolute inset-0 opacity-20"
      :style="{
        backgroundPositionY: `${playerStore.currentTime * 50}px`,
      }"
    />
  </div>
</template>
