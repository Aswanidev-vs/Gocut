<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { StartRender, GetRenderProgress, CancelRender, SaveFilePicker } from '../../lib/wails'
import { onWailsEvent, offWailsEvent } from '../../lib/wailsEvents'
import { X, FileDown, Loader2, CheckCircle2, AlertCircle, FolderOpen, XCircle, Minimize2 } from 'lucide-vue-next'

const props = defineProps({ isOpen: Boolean })
const emit = defineEmits(['close'])

const projectStore = useProjectStore()
const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const format = ref('mp4')
const codec = ref('h264')
const resolution = ref('1080p')
const fps = ref(30)
const useProjectFps = ref(true)
const crf = ref(23)
const preset = ref('ultrafast')
const audioBitrate = ref('192k')
// GIF is palette based, so these four knobs are what "compression" means for
// it: fewer frames, fewer colors, ordered dithering, smaller frame.
const gifFps = ref(15)
const gifColors = ref(128)
const gifDither = ref('bayer')
const gifMaxWidth = ref(640)
const includeIn = ref(0)
const includeOut = ref(0)
const outputPath = ref('')

const currentJobId = ref(null)
const isRendering = ref(false)
const progress = ref(0)
const status = ref('idle') // idle | queued | rendering | done | error | cancelled
const errorMessage = ref('')
const finalOutputPath = ref('')
let progressPollTimer = null

function stopProgressPolling() {
  if (progressPollTimer !== null) {
    clearInterval(progressPollTimer)
    progressPollTimer = null
  }
}

const formats = [
  { id: 'mp4',  label: 'MP4 (H.264)', codec: 'h264' },
  { id: 'webm', label: 'WebM (VP9)',  codec: 'vp9' },
  { id: 'gif',  label: 'GIF',          codec: 'gif' },
  { id: 'mp3',  label: 'MP3 (Audio)',  codec: 'mp3' },
  { id: 'aac',  label: 'AAC (Audio)',  codec: 'aac' },
]
const resolutions = [
  { id: '480p',  label: '480p',  w: 854,  h: 480  },
  { id: '720p',  label: '720p',  w: 1280, h: 720  },
  { id: '1080p', label: '1080p', w: 1920, h: 1080 },
  { id: '4K',    label: '4K',    w: 3840, h: 2160 },
]
const presets = ['ultrafast', 'superfast', 'veryfast', 'faster', 'fast', 'medium', 'slow', 'slower', 'veryslow']

const gifPresets = [
  { id: 'high',     label: 'High',     hint: 'Best looking, largest file', fps: 24, colors: 256, dither: 'sierra2_4a', maxWidth: 0 },
  { id: 'balanced', label: 'Balanced', hint: 'Recommended everyday GIF',   fps: 15, colors: 128, dither: 'bayer',      maxWidth: 640 },
  { id: 'tiny',     label: 'Smallest', hint: 'Chat and mobile friendly',   fps: 10, colors: 64,  dither: 'bayer',      maxWidth: 480 },
]
const gifColorsOptions = [256, 128, 64, 32]
const gifMaxWidthOptions = [
  { value: 0,   label: 'Export size' },
  { value: 640, label: '640 px' },
  { value: 480, label: '480 px' },
  { value: 320, label: '320 px' },
]

const activeGifPreset = computed(() => gifPresets.find(p =>
  p.fps === gifFps.value && p.colors === gifColors.value &&
  p.dither === gifDither.value && p.maxWidth === gifMaxWidth.value)?.id)

// Mirrors the backend's downscale rule so the dialog can show the real output
// size rather than an estimate.
const gifOutputSize = computed(() => {
  const res = getResolution()
  if (!gifMaxWidth.value || res.w <= gifMaxWidth.value) return `${res.w}×${res.h}`
  return `${gifMaxWidth.value}×${Math.round(res.h * gifMaxWidth.value / res.w)}`
})

function applyGifPreset(p) {
  gifFps.value = p.fps
  gifColors.value = p.colors
  gifDither.value = p.dither
  gifMaxWidth.value = p.maxWidth
}

function applyFormat() {
  const f = formats.find(f => f.id === format.value)
  if (f) codec.value = f.codec
}

function getResolution() {
  return resolutions.find(r => r.id === resolution.value) || resolutions[1]
}

async function pickOutput() {
  try {
    const name = (projectStore.projectName || 'export') + '.' + format.value
    const paths = await SaveFilePicker(name, [{ name: 'Output', extensions: [format.value] }])
    if (paths) {
      outputPath.value = paths
    }
  } catch (e) {
    // user cancelled; ignore
  }
}

function fmtTime(seconds) {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  const ms = Math.floor((seconds % 1) * 1000)
  const pad = (n, len = 2) => String(n).padStart(len, '0')
  return pad(h) + ':' + pad(m) + ':' + pad(s) + '.' + pad(ms, 3)
}

async function startExport() {
  if (!projectStore.project) {
    uiStore.addToast('No project loaded', 'error')
    return
  }
  if (!outputPath.value) {
    uiStore.addToast('Please pick an output file', 'warn')
    return
  }
  const res = getResolution()
  stopProgressPolling()
  isRendering.value = true
  status.value = 'queued'
  progress.value = 0
  errorMessage.value = ''
  finalOutputPath.value = ''

  const settings = {
    jobId: '',
    outputPath: outputPath.value,
    format: format.value,
    codec: codec.value,
    width: res.w,
    height: res.h,
    fps: useProjectFps.value ? (projectStore.project.fps || 30) : fps.value,
    bitrate: 'auto',
    audioBitrate: audioBitrate.value,
    crf: crf.value,
    preset: preset.value,
    gifFps: gifFps.value,
    gifColors: gifColors.value,
    gifDither: gifDither.value,
    gifMaxWidth: gifMaxWidth.value,
    startTime: includeIn.value,
    endTime: includeOut.value || timelineStore.duration,
  }

  try {
    const jobId = await StartRender(JSON.parse(JSON.stringify(projectStore.project)), settings)
    currentJobId.value = jobId
    status.value = 'rendering'
    startProgressPolling(jobId)
  } catch (e) {
    stopProgressPolling()
    isRendering.value = false
    status.value = 'error'
    errorMessage.value = e?.message || String(e)
    uiStore.addToast('Failed to start render: ' + errorMessage.value, 'error')
  }
}

async function cancelExport() {
  if (!currentJobId.value) return
  try {
    await CancelRender(currentJobId.value)
    stopProgressPolling()
    status.value = 'cancelled'
    isRendering.value = false
    uiStore.addToast('Render cancelled', 'info')
  } catch (e) {
    uiStore.addToast('Cancel failed: ' + (e?.message || e), 'error')
  }
}

function handleProgressEvent(ev) {
  if (!ev || ev.jobId !== currentJobId.value) return
  if (ev.percent != null) progress.value = ev.percent
  if (ev.status) status.value = ev.status
  if (ev.error) errorMessage.value = ev.error
  if (ev.outputPath) finalOutputPath.value = ev.outputPath
  if (ev.status === 'done' || ev.status === 'error' || ev.status === 'cancelled') {
    stopProgressPolling()
    isRendering.value = false
  }
}

async function pollRenderProgress(jobId) {
  if (!jobId || jobId !== currentJobId.value) return

  try {
    const ev = await GetRenderProgress(jobId)
    if (jobId === currentJobId.value) {
      handleProgressEvent(ev)
    }
  } catch (e) {
    // The event stream is still the primary path. Ignore a transient poll
    // failure so it cannot replace a useful FFmpeg error event.
  }
}

function startProgressPolling(jobId) {
  stopProgressPolling()
  // Events can be emitted before StartRender resolves. Polling closes that
  // small race and also handles very fast FFmpeg failures reliably.
  pollRenderProgress(jobId)
  progressPollTimer = setInterval(() => pollRenderProgress(jobId), 300)
}

onWailsEvent('render:progress', handleProgressEvent)
onUnmounted(() => {
  stopProgressPolling()
  offWailsEvent('render:progress', handleProgressEvent)
})

watch(() => props.isOpen, (open) => {
  if (open) {
    status.value = 'idle'
    progress.value = 0
    errorMessage.value = ''
    finalOutputPath.value = ''
    currentJobId.value = null
    isRendering.value = false
    includeOut.value = timelineStore.duration
    if (!useProjectFps.value && projectStore.project?.fps) {
      fps.value = projectStore.project.fps
    }
    if (!outputPath.value) {
      const sep = navigator.platform.includes('Win') ? '\\' : '/'
      outputPath.value = '~' + sep + 'Desktop' + sep + (projectStore.projectName || 'export') + '.' + format.value
    }
  } else {
    stopProgressPolling()
  }
})
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-bg/70 backdrop-blur-sm" @click.self="emit('close')">
    <div class="w-[520px] bg-panel border border-border rounded-md shadow-2xl">
      <div class="flex items-center justify-between px-4 py-3 border-b border-border">
        <h3 class="text-sm font-semibold text-text-primary">Export</h3>
        <button class="p-1 rounded hover:bg-border text-text-secondary" @click="emit('close')"><X :size="14" /></button>
      </div>

      <!-- Progress overlay (when rendering) -->
      <div v-if="status === 'queued' || status === 'rendering'" class="px-4 py-6 flex flex-col items-center gap-3">
        <Loader2 :size="32" class="text-accent animate-spin" />
        <div class="text-sm text-text-primary">{{ status === 'queued' ? 'Queued…' : 'Rendering…' }}</div>
        <div class="w-full h-1.5 bg-border rounded-full overflow-hidden">
          <div class="h-full bg-accent transition-all duration-1000 ease-out" :style="{ width: (progress > 0 ? progress : 2) + '%' }" />
        </div>
        <div class="text-[10px] text-text-secondary font-mono">
          <span v-if="progress === 0 && status === 'rendering'">Initializing Engine...</span>
          <span v-else>{{ Math.round(progress) }}%</span>
        </div>
        <button class="px-3 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border border border-border" @click="cancelExport">
          Cancel
        </button>
      </div>

      <div v-else-if="status === 'done'" class="px-4 py-6 flex flex-col items-center gap-3 text-center">
        <CheckCircle2 :size="32" class="text-green-400" />
        <div class="text-sm text-text-primary">Export complete!</div>
        <div class="text-[10px] text-text-secondary font-mono break-all max-w-md">{{ finalOutputPath }}</div>
        <button class="px-3 py-1.5 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors" @click="emit('close')">
          Done
        </button>
      </div>

      <div v-else-if="status === 'error'" class="px-4 py-6 flex flex-col items-center gap-3 text-center">
        <AlertCircle :size="32" class="text-red-400" />
        <div class="text-sm text-text-primary">Export failed</div>
        <div class="text-[10px] text-red-400 max-w-md break-all">{{ errorMessage }}</div>
        <button class="px-3 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border border border-border" @click="status = 'idle'">
          Try Again
        </button>
      </div>

      <div v-else-if="status === 'cancelled'" class="px-4 py-6 flex flex-col items-center gap-3 text-center">
        <XCircle :size="32" class="text-text-secondary" />
        <div class="text-sm text-text-primary">Render cancelled</div>
        <button class="px-3 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border border border-border" @click="status = 'idle'">
          Back
        </button>
      </div>

      <div v-else class="p-4 grid grid-cols-2 gap-3">
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Format</label>
          <select v-model="format" @change="applyFormat" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
            <option v-for="f in formats" :key="f.id" :value="f.id">{{ f.label }}</option>
          </select>
        </div>
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Codec</label>
          <input v-model="codec" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent" />
        </div>
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Resolution</label>
          <select v-model="resolution" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
            <option v-for="r in resolutions" :key="r.id" :value="r.id">{{ r.label }} ({{ r.w }}×{{ r.h }})</option>
          </select>
        </div>
        <!-- GIF is palette based and has no inter-frame compression, so frame
             rate, palette size, dithering and frame width are the only levers
             that actually shrink the file. -->
        <div v-if="format === 'gif'" class="col-span-2 rounded-lg border border-border bg-bg/40 p-3 space-y-3">
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 text-[11px] font-semibold text-text-primary">
              <Minimize2 :size="12" class="text-accent" /> GIF Compression
            </div>
            <div class="text-[10px] font-mono text-text-secondary truncate">
              {{ gifOutputSize }} · {{ gifFps }} fps · {{ gifColors }} colors
            </div>
          </div>

          <div class="flex items-center gap-1">
            <button
              v-for="p in gifPresets"
              :key="p.id"
              :title="p.hint"
              class="flex-1 px-2 py-1 rounded text-[10px] font-medium border transition-colors"
              :class="activeGifPreset === p.id
                ? 'bg-accent/15 text-accent border-accent/30'
                : 'text-text-secondary border-border hover:text-text-primary hover:bg-border/60'"
              @click="applyGifPreset(p)"
            >
              {{ p.label }}
            </button>
          </div>

          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="text-[10px] text-text-secondary block mb-1 uppercase tracking-wider">Frame Rate</label>
              <select v-model.number="gifFps" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
                <option v-for="f in [24, 20, 15, 12, 10]" :key="f" :value="f">{{ f }} fps</option>
              </select>
            </div>
            <div>
              <label class="text-[10px] text-text-secondary block mb-1 uppercase tracking-wider">Colors</label>
              <select v-model.number="gifColors" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
                <option v-for="c in gifColorsOptions" :key="c" :value="c">{{ c }}</option>
              </select>
            </div>
            <div>
              <label class="text-[10px] text-text-secondary block mb-1 uppercase tracking-wider">Dither</label>
              <select v-model="gifDither" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
                <option value="bayer">Bayer (smallest)</option>
                <option value="none">None (banding)</option>
                <option value="sierra2_4a">Sierra (smoothest)</option>
              </select>
            </div>
            <div>
              <label class="text-[10px] text-text-secondary block mb-1 uppercase tracking-wider">Max Width</label>
              <select v-model.number="gifMaxWidth" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
                <option v-for="o in gifMaxWidthOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
              </select>
            </div>
          </div>
        </div>

        <div v-if="format !== 'gif'">
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">FPS</label>
          <div class="flex items-center gap-2">
            <label class="flex items-center gap-1 text-xs text-text-secondary">
              <input type="checkbox" v-model="useProjectFps" /> Project
            </label>
            <select :disabled="useProjectFps" v-model="fps" class="flex-1 bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent disabled:opacity-50">
              <option :value="24">24</option>
              <option :value="25">25</option>
              <option :value="30">30</option>
              <option :value="60">60</option>
            </select>
          </div>
        </div>
        <div v-if="format !== 'gif'">
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Quality (CRF)</label>
          <input type="range" min="0" max="51" v-model.number="crf" class="w-full accent-accent" />
          <div class="text-[10px] text-text-secondary text-right font-mono">{{ crf }}</div>
        </div>
        <div v-if="format !== 'gif'">
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Preset</label>
          <select v-model="preset" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary outline-none focus:border-accent">
            <option v-for="p in presets" :key="p" :value="p">{{ p }}</option>
          </select>
        </div>
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">In</label>
          <input type="number" step="0.1" min="0" v-model.number="includeIn" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary font-mono" />
        </div>
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Out</label>
          <input type="number" step="0.1" min="0" v-model.number="includeOut" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm text-text-primary font-mono" />
        </div>
        <div class="col-span-2">
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Output File</label>
          <div class="flex items-center gap-1">
            <input v-model="outputPath" class="flex-1 bg-bg border border-border rounded px-2 py-1.5 text-xs text-text-primary font-mono" />
            <button class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border" @click="pickOutput" title="Browse">
              <FolderOpen :size="14" />
            </button>
          </div>
        </div>
      </div>

      <div v-if="status === 'idle'" class="flex justify-end gap-2 px-4 py-3 border-t border-border bg-bg/30">
        <button class="px-3 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border" @click="emit('close')">Cancel</button>
        <button class="flex items-center gap-1.5 px-3 py-1.5 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors" @click="startExport">
          <FileDown :size="14" /> Start Export
        </button>
      </div>
    </div>
  </div>
</template>
