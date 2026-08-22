<script setup>
import { computed } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { OpenFilePicker } from '../../lib/wails'
import { Music, Plus, Volume2, VolumeX, Waves, Sparkles, Repeat, AudioLines, ArrowDownToLine } from 'lucide-vue-next'
import Waveform from '../common/Waveform.vue'

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

// ---- Selected audio clip ----
const selectedClip = computed(() => timelineStore.selectedClips[0] || null)
const selectedTrack = computed(() => {
  const c = selectedClip.value
  if (!c) return null
  return timelineStore.tracks.find(t => t.id === c.trackId) || null
})
const selectedAsset = computed(() => {
  const c = selectedClip.value
  if (!c) return null
  return projectStore.getAsset(c.assetId)
})
const isAudioClip = computed(() => selectedTrack.value?.type === 'audio')

function setClip(key, val) {
  if (!selectedClip.value) return
  timelineStore.updateClip(selectedClip.value.id, { [key]: val })
}

const volumePct = computed({
  get: () => Math.round((selectedClip.value?.volume ?? 1) * 100),
  set: (v) => setClip('volume', Math.max(0, Math.min(2, (Number(v) || 0) / 100))),
})

function toggleMute() {
  const t = selectedTrack.value
  if (!t) return
  t.muted = !t.muted
  projectStore.markDirty()
}

const fadeInSeconds = computed({
  get: () => {
    const tr = selectedClip.value?.transition
    if (tr && tr.type && tr.type !== 'none') return tr.duration || 0
    return 0
  },
  set: (v) => {
    if (!selectedClip.value) return
    const d = Math.max(0, Number(v) || 0)
    if (d <= 0) {
      timelineStore.updateClip(selectedClip.value.id, { transition: { type: 'none' } })
    } else {
      timelineStore.updateClip(selectedClip.value.id, { transition: { type: 'fade', duration: d } })
    }
  },
})

function toggleClass(on) {
  return on
    ? 'border-accent text-accent bg-accent/10'
    : 'border-border text-text-secondary hover:border-accent/40'
}
</script>

<template>
  <div class="p-2 flex flex-col gap-3">
    <!-- Keep existing audio import behavior intact -->
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
        <!-- Asset-level waveform (SVG) -->
        <div class="h-6 flex items-center bg-bg/50 border border-border/40 rounded p-0.5">
          <Waveform v-if="a.waveform && a.waveform.length" :samples="a.waveform" />
          <div v-else class="text-[9px] text-text-secondary italic w-full text-center">generating…</div>
        </div>
      </div>
    </div>

    <!-- Selected audio clip editing controls -->
    <template v-if="isAudioClip">
      <hr class="border-border" />

      <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider flex items-center gap-1.5">
        <AudioLines :size="11" /> Selected Audio Clip
      </h4>

      <!-- Waveform of the clip's asset (500 samples) -->
      <div class="h-20 bg-bg/50 border border-border rounded p-1">
        <Waveform v-if="selectedAsset && selectedAsset.waveform && selectedAsset.waveform.length" :samples="selectedAsset.waveform" />
        <div v-else class="w-full h-full flex items-center justify-center text-[10px] text-text-secondary italic">No waveform</div>
      </div>

      <!-- Volume -->
      <div>
        <div class="flex justify-between text-[10px] text-text-secondary mb-1">
          <span>Volume</span>
          <span class="font-mono">{{ volumePct }}%</span>
        </div>
        <div class="flex items-center gap-2">
          <Volume2 :size="12" class="text-text-secondary flex-shrink-0" />
          <input type="range" min="0" max="200" step="1" v-model.number="volumePct" class="flex-1 accent-accent" />
          <input type="number" min="0" max="200" step="1" v-model.number="volumePct"
            class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
        </div>
      </div>

      <!-- Mute (track-level) -->
      <button
        class="w-full flex items-center justify-center gap-1.5 px-3 py-1.5 rounded text-xs font-medium border transition-all"
        :class="toggleClass(selectedTrack?.muted)"
        @click="toggleMute"
      >
        <VolumeX v-if="selectedTrack?.muted" :size="12" />
        <Volume2 v-else :size="12" />
        {{ selectedTrack?.muted ? 'Muted (track)' : 'Mute track' }}
      </button>

      <!-- Toggles -->
      <div class="grid grid-cols-2 gap-1.5">
        <button
          class="flex items-center justify-center gap-1 px-2 py-1.5 rounded text-[11px] font-medium border transition-all"
          :class="toggleClass(selectedClip?.normalize)"
          @click="setClip('normalize', !selectedClip?.normalize)"
        >
          <Waves :size="12" /> Normalize
        </button>
        <button
          class="flex items-center justify-center gap-1 px-2 py-1.5 rounded text-[11px] font-medium border transition-all"
          :class="toggleClass(selectedClip?.noiseReduction)"
          @click="setClip('noiseReduction', !selectedClip?.noiseReduction)"
        >
          <Sparkles :size="12" /> Noise Red.
        </button>
        <button
          class="flex items-center justify-center gap-1 px-2 py-1.5 rounded text-[11px] font-medium border transition-all"
          :class="toggleClass(selectedClip?.loop)"
          @click="setClip('loop', !selectedClip?.loop)"
        >
          <Repeat :size="12" /> BGM Loop
        </button>
        <button
          class="flex items-center justify-center gap-1 px-2 py-1.5 rounded text-[11px] font-medium border transition-all"
          :class="toggleClass(selectedClip?.duck)"
          @click="setClip('duck', !selectedClip?.duck)"
        >
          <ArrowDownToLine :size="12" /> Duck
        </button>
      </div>

      <!-- Fade In -->
      <div>
        <label class="text-[10px] text-text-secondary block mb-1">Fade In (seconds)</label>
        <input type="number" min="0" max="60" step="0.1" v-model.number="fadeInSeconds"
          class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
        <div class="text-[9px] text-text-secondary mt-1">
          0 = none. Sets clip.Transition = { type:'fade', duration }.
        </div>
      </div>
    </template>

    <div v-else-if="selectedClip" class="text-[10px] text-text-secondary text-center py-2">
      Select an audio clip to edit its sound.
    </div>
  </div>
</template>
