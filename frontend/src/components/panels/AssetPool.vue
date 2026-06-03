<script setup>
import { computed, ref } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { OpenFilePicker } from '../../lib/wails'
import { Plus, Upload, Video, Music, Image as ImageIcon, FileText, X } from 'lucide-vue-next'

const projectStore = useProjectStore()
const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const isImporting = ref(false)
const filter = ref('all')

const extensions = {
  video: { name: 'Video', extensions: ['mp4', 'mov', 'avi', 'mkv', 'webm'] },
  audio: { name: 'Audio', extensions: ['mp3', 'wav', 'aac', 'flac'] },
  image: { name: 'Image', extensions: ['png', 'jpg', 'jpeg', 'gif', 'webp'] },
}

const assets = computed(() => projectStore.project?.assets || [])

const filteredAssets = computed(() => {
  if (filter.value === 'all') return assets.value
  return assets.value.filter(a => a.type === filter.value)
})

const filters = [
  { id: 'all',   label: 'All' },
  { id: 'video', label: 'Video' },
  { id: 'audio', label: 'Audio' },
  { id: 'image', label: 'Image' },
]

const draggingAssetId = ref(null)

function startDragAsset(e, asset) {
  draggingAssetId.value = asset.id
  e.dataTransfer.effectAllowed = 'copy'
  e.dataTransfer.setData('application/x-gocut-asset', asset.id)
  e.dataTransfer.setData('text/plain', asset.path)
}

function endDragAsset() {
  draggingAssetId.value = null
}

/**
 * Run the actual import + add-to-timeline pipeline once we have a list of
 * absolute file paths. Used by both the picker and drag-and-drop handlers.
 */
async function importPaths(paths) {
  if (!paths || paths.length === 0) return
  isImporting.value = true
  try {
    const imported = await projectStore.importMedia(paths)
    if (imported.length === 0) {
      uiStore.addToast('No files could be imported. Check that ffmpeg can read them.', 'error', 4000)
      return
    }
    for (const asset of imported) {
      let trackType = 'video'
      if (asset.type === 'audio') trackType = 'audio'
      else if (asset.type === 'image') trackType = 'video'
      timelineStore.addClip({
        assetId: asset.id,
        trackType,
        startTime: timelineStore.currentTime,
        duration: asset.duration || 3,
      })
      // If it's a video, automatically add an associated audio clip on the audio track
      if (asset.type === 'video') {
        timelineStore.addClip({
          assetId: asset.id,
          trackType: 'audio',
          startTime: timelineStore.currentTime,
          duration: asset.duration || 3,
        })
      }
    }
    uiStore.addToast(`Imported ${imported.length} file${imported.length > 1 ? 's' : ''}`, 'success', 2000)
  } catch (err) {
    uiStore.addToast(err?.message || 'Failed to import', 'error', 4000)
  } finally {
    isImporting.value = false
  }
}

async function handleFileDrop(e) {
  e.preventDefault()
  const files = Array.from(e.dataTransfer.files || [])
  // Wails v2 WebView does not always expose absolute paths via
  // dataTransfer.files[i].path. If we got only names, fall back to the
  // picker so the user can still select the files.
  const paths = files
    .map(f => f.path)
    .filter(p => p && (p.includes('\\') || p.includes('/')))
  if (paths.length === 0) {
    uiStore.addToast('Drag-and-drop cannot resolve file paths in this build. Use the Import button.', 'warn', 4000)
    await pickAndImport()
    return
  }
  await importPaths(paths)
}

function onDragOver(e) { e.preventDefault() }
function onDragLeave() {}

async function pickAndImport() {
  isImporting.value = true
  try {
    const result = await OpenFilePicker([
      extensions.video,
      extensions.audio,
      extensions.image,
    ])
    // Normalize: the Wails binding returns []string. Be defensive in case
    // a future build returns a single string.
    let paths = []
    if (Array.isArray(result)) paths = result
    else if (typeof result === 'string' && result) paths = [result]
    if (paths.length === 0) return
    await importPaths(paths)
  } catch (err) {
    uiStore.addToast(err?.message || 'Failed to open file picker', 'error')
  } finally {
    isImporting.value = false
  }
}

function addToTimeline(asset) {
  let trackType = 'video'
  if (asset.type === 'audio') trackType = 'audio'
  else if (asset.type === 'image') trackType = 'video'
  timelineStore.addClip({
    assetId: asset.id,
    trackType,
    startTime: timelineStore.currentTime,
    duration: asset.duration || 3,
  })
  // If it's a video, automatically add an associated audio clip on the audio track
  if (asset.type === 'video') {
    timelineStore.addClip({
      assetId: asset.id,
      trackType: 'audio',
      startTime: timelineStore.currentTime,
      duration: asset.duration || 3,
    })
  }
  uiStore.addToast('Added to timeline', 'success', 1200)
}

function removeAsset(assetId) {
  const ts = useTimelineStore()
  const orphans = ts.clips.filter(c => c.assetId === assetId)
  for (const c of orphans) ts.removeClip(c.id)
  projectStore.removeAsset(assetId)
  uiStore.addToast('Asset removed', 'info', 1200)
}

function assetIcon(type) {
  return { video: Video, audio: Music, image: ImageIcon }[type] || FileText
}

function fileName(path) {
  if (!path) return 'Untitled'
  const sep = path.includes('\\') ? '\\' : '/'
  const last = path.split(sep).pop() || path
  return last
}

function fmtDur(seconds) {
  if (!seconds || isNaN(seconds)) return ''
  const m = Math.floor(seconds / 60)
  const s = Math.floor(seconds % 60)
  return `${m}:${String(s).padStart(2, '0')}`
}
</script>

<template>
  <div
    class="h-full flex flex-col"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="handleFileDrop"
  >
    <div class="flex items-center gap-1 px-2 pt-2 pb-2 flex-wrap">
      <button
        v-for="f in filters"
        :key="f.id"
        class="px-2 py-0.5 rounded-full text-[10px] border transition-colors"
        :class="filter === f.id
          ? 'border-accent bg-accent/10 text-accent'
          : 'border-border text-text-secondary hover:text-text-primary'"
        @click="filter = f.id"
      >
        {{ f.label }}
      </button>
      <div class="flex-1" />
      <button
        class="flex items-center gap-1 px-2 py-1 rounded text-[10px] bg-accent/10 text-accent hover:bg-accent/20 transition-colors"
        @click="pickAndImport"
        :disabled="isImporting"
        title="Import media"
      >
        <Plus :size="10" /> Import
      </button>
    </div>

    <div
      v-if="filteredAssets.length === 0"
      class="flex-1 flex flex-col items-center justify-center px-4 py-6 text-center"
    >
      <div class="w-12 h-12 rounded-full bg-accent/10 flex items-center justify-center mb-3">
        <Upload :size="20" class="text-accent" />
      </div>
      <div class="text-sm text-text-primary mb-1">Import media to get started</div>
      <div class="text-[11px] text-text-secondary mb-3">MP4, MOV, MKV, WebM, MP3, WAV, PNG, JPG</div>
      <button
        class="px-3 py-1.5 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors"
        @click="pickAndImport"
        :disabled="isImporting"
      >
        {{ isImporting ? 'Importing…' : 'Choose Files' }}
      </button>
    </div>

    <div v-else class="flex-1 overflow-y-auto px-2 pb-2">
      <div class="grid grid-cols-2 gap-1.5">
        <div
          v-for="asset in filteredAssets"
          :key="asset.id"
          class="group rounded bg-bg/60 border border-border/60 overflow-hidden hover:border-accent/50 transition-colors cursor-grab active:cursor-grabbing"
          :class="{ 'opacity-50': draggingAssetId === asset.id }"
          draggable="true"
          @dragstart="(e) => startDragAsset(e, asset)"
          @dragend="endDragAsset"
          @dblclick="addToTimeline(asset)"
        >
          <div class="aspect-video bg-bg relative flex items-center justify-center overflow-hidden">
            <img
              v-if="asset.thumbnail"
              :src="`data:image/jpeg;base64,${asset.thumbnail}`"
              class="w-full h-full object-cover"
              :alt="fileName(asset.path)"
            />
            <component v-else :is="assetIcon(asset.type)" :size="22" class="text-text-secondary" />
            <div v-if="asset.duration" class="absolute bottom-1 right-1 text-[9px] bg-bg/85 px-1 py-0.5 rounded text-text-secondary font-mono">
              {{ fmtDur(asset.duration) }}
            </div>
            <div class="absolute top-1 left-1 text-[9px] bg-bg/85 px-1 py-0.5 rounded uppercase text-text-secondary">
              {{ asset.type }}
            </div>
            <div class="absolute inset-0 bg-accent/0 group-hover:bg-accent/10 transition-colors flex items-center justify-center pointer-events-none">
              <Plus :size="20" class="text-accent opacity-0 group-hover:opacity-100 transition-opacity" />
            </div>
          </div>
          <div class="px-1.5 py-1 flex items-center gap-1">
            <span class="text-[10px] text-text-primary truncate flex-1" :title="fileName(asset.path)">{{ fileName(asset.path) }}</span>
            <button
              class="p-0.5 rounded text-text-secondary hover:text-red-400 hover:bg-border transition-colors opacity-0 group-hover:opacity-100"
              @click.stop="removeAsset(asset.id)"
              title="Remove"
            >
              <X :size="9" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
