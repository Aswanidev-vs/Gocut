<script setup>
import { computed, ref } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { OpenFilePicker } from '../../lib/wails'
import { Music, Plus, X } from 'lucide-vue-next'

const projectStore = useProjectStore()
const timelineStore = useTimelineStore()

const audioAssets = computed(() =>
  (projectStore.project?.assets || []).filter(a => a.type === 'audio')
)

async function pickAudio() {
  try {
    const paths = await OpenFilePicker([{ name: 'Audio', extensions: ['mp3', 'wav', 'aac', 'flac'] }])
    if (!paths || paths.length === 0) return
    await projectStore.importMedia(paths)
  } catch (e) {
    // ignore
  }
}

function addToTimeline(asset) {
  timelineStore.addClip({
    assetId: asset.id,
    trackType: 'audio',
    startTime: timelineStore.currentTime,
    duration: asset.duration || 30,
  })
}
</script>

<template>
  <div class="p-2 flex flex-col gap-2">
    <button
      class="flex items-center justify-center gap-1.5 w-full py-2 rounded border border-dashed border-border text-xs text-text-secondary hover:text-text-primary hover:border-accent transition-colors"
      @click="pickAudio"
    >
      <Plus :size="12" /> Add Audio
    </button>

    <div v-if="audioAssets.length === 0" class="text-[11px] text-text-secondary text-center py-4">
      No audio imported yet
    </div>

    <div v-else class="flex flex-col gap-1.5">
      <div
        v-for="a in audioAssets"
        :key="a.id"
        class="rounded bg-bg/60 border border-border/60 p-2 hover:border-accent/50 cursor-pointer"
        @dblclick="addToTimeline(a)"
      >
        <div class="flex items-center gap-1.5 mb-1.5">
          <Music :size="12" class="text-text-secondary flex-shrink-0" />
          <span class="text-[10px] text-text-primary truncate flex-1" :title="a.path">{{ (a.path || '').split(/[\\/]/).pop() }}</span>
        </div>
        <!-- Inline mini waveform -->
        <div class="h-6 flex items-center gap-px">
          <div
            v-for="(v, i) in (a.waveform || []).slice(0, 64)"
            :key="i"
            class="flex-1 bg-accent/50 rounded-sm"
            :style="{ height: Math.max(2, Math.abs(v) * 100) + '%' }"
          />
          <div v-if="!a.waveform?.length" class="text-[9px] text-text-secondary italic w-full text-center">
            generating…
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
