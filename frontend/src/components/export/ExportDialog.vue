<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { StartRender, GetRenderProgress, CancelRender, SaveFilePicker } from '../../lib/wails'
import { onWailsEvent, offWailsEvent } from '../../lib/wailsEvents'
import { X, FileDown, Loader2, CheckCircle2, AlertCircle, FolderOpen, XCircle } from 'lucide-vue-next'

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
const includeIn = ref(0)
const includeOut = ref(0)
const outputPath = ref('')

const currentJobId = ref(null)
const isRendering = ref(false)
const progress = ref(0)
const status = ref('idle') // idle | queued | rendering | done | error | cancelled
const errorMessage = ref('')
const finalOutputPath = ref('')

const formats = [
  { id: 'mp4',  label: 'MP4 (H.264)', codec: 'h264' },
  { id: 'webm', label: 'WebM (VP9)',  codec: 'vp9' },
  { id: 'gif',  label: 'GIF',          codec: 'gif' },
  { id: 'mp3',  label: 'MP3 (Audio)',  codec: 'mp3' },
]
const resolutions = [
  { id: '480p',  label: '480p',  w: 854,  h: 480  },
  { id: '720p',  label: '720p',  w: 1280, h: 720  },
  { id: '1080p', label: '1080p', w: 1920, h: 1080 },
  { id: '4K',    label: '4K',    w: 3840, h: 2160 },
]
const presets = ['ultrafast', 'superfast', 'veryfast', 'faster', 'fast', 'medium', 'slow', 'slower', 'veryslow']

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
    startTime: includeIn.value,
    endTime: includeOut.value || timelineStore.duration,
  }

  try {
    const jobId = await StartRender(JSON.parse(JSON.stringify(projectStore.project)), settings)
    currentJobId.value = jobId
    status.value = 'rendering'
  } catch (e) {
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
    isRendering.value = false
  }
}

onWailsEvent('render:progress', handleProgressEvent)
onUnmounted(() => offWailsEvent('render:progress', handleProgressEvent))

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
        <div>
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
        <div>
          <label class="text-[11px] text-text-secondary block mb-1 uppercase tracking-wider">Quality (CRF)</label>
          <input type="range" min="0" max="51" v-model.number="crf" class="w-full accent-accent" />
          <div class="text-[10px] text-text-secondary text-right font-mono">{{ crf }}</div>
        </div>
        <div>
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
