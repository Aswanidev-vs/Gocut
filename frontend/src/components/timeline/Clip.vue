<script setup>
import { computed, ref } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore, getInterpolatedProperty } from '../../stores/timelineStore'
import { ArrowRightLeft, Palette } from 'lucide-vue-next'

const props = defineProps({
  clip: { type: Object, required: true },
  zoom: { type: Number, default: 50 },
  trackType: { type: String, default: 'video' },
  trackColor: { type: String, default: '#00D4FF' },
  selected: { type: Boolean, default: false },
})

const timelineStore = useTimelineStore()
const projectStore = useProjectStore()

const leftPx = computed(() => props.clip.startTime * props.zoom)
const widthPx = computed(() => Math.max(8, props.clip.duration * props.zoom))
const asset = computed(() => projectStore.getAsset(props.clip.assetId))
const fileName = computed(() => {
  if (props.clip.textProps) return props.clip.textProps.text || 'Text'
  if (props.clip.stickerProps) return props.clip.stickerProps.name || 'Sticker'
  if (!asset.value?.path) return 'Clip'
  const sep = asset.value.path.includes('\\') ? '\\' : '/'
  return asset.value.path.split(sep).pop() || 'Clip'
})

const hasTransition = computed(() => {
  return props.clip.transition && props.clip.transition.type !== 'none' && props.clip.transition.duration > 0
})
const hasEffects = computed(() => {
  const c = props.clip.color
  return c && (c.brightness || c.contrast || c.saturation || c.hue || c.sharpness ||
               c.vignette || c.grain || c.blur || c.chromaKeyColor)
})

const showMenu = ref(false)
const menuX = ref(0)
const menuY = ref(0)

function onContextMenu(e) {
  e.preventDefault()
  e.stopPropagation()
  timelineStore.selectClip(props.clip.id, false)
  menuX.value = e.clientX
  menuY.value = e.clientY
  showMenu.value = true
  window.addEventListener('click', closeMenu, { once: true })
}
function closeMenu() {
  showMenu.value = false
}
function splitClip() {
  timelineStore.splitClipAt(props.clip.id, timelineStore.currentTime)
  closeMenu()
}
function duplicateClip() {
  timelineStore.duplicateClip(props.clip.id)
  closeMenu()
}
function deleteClip() {
  timelineStore.removeClip(props.clip.id)
  closeMenu()
}
function setSpeed(s) {
  timelineStore.updateClip(props.clip.id, { speed: s })
  closeMenu()
}

// ---- drag to move ----
const dragState = ref(null)

// Convert a clientX coordinate to a timeline time (seconds), accounting for
// the horizontal scroll of the .timeline-content container (same math as the
// asset drop handlers).
function clientXToTime(clientX) {
  const scrollContainer = document.querySelector('.timeline-content')
  const rect = scrollContainer ? scrollContainer.getBoundingClientRect() : null
  if (!rect) return null
  const scrollLeft = scrollContainer.scrollLeft || 0
  return (clientX - rect.left + scrollLeft) / timelineStore.zoom
}

function onMouseDown(e, mode) {
  e.stopPropagation()
  if (e.button !== 0) return

  // Razor tool: a plain left-click on the clip body splits it at the click
  // position. No selection, no drag.
  if (mode === 'move' && timelineStore.activeTool === 'razor') {
    const t = clientXToTime(e.clientX)
    if (t !== null) timelineStore.splitClipAt(props.clip.id, t)
    return
  }

  timelineStore.selectClip(props.clip.id, e.shiftKey)

  // Group drag: grabbing an already-selected clip while others are selected
  // moves the whole selection together (no snapping during group moves).
  const isGroupDrag = mode === 'move' &&
    timelineStore.selectedClipIds.length > 1 &&
    timelineStore.selectedClipIds.includes(props.clip.id)

  const startX = e.clientX
  const startStart = props.clip.startTime
  const startDuration = props.clip.duration
  const startTrimStart = props.clip.trimStart
  let lastDt = 0 // incremental delta for group drags

  const onMove = (mv) => {
    const dx = mv.clientX - startX
    const dt = dx / props.zoom
    if (mode === 'move') {
      if (isGroupDrag) {
        timelineStore.shiftSelection(dt - lastDt)
        lastDt = dt
        return
      }
      const raw = Math.max(0, startStart + dt)
      const snapped = timelineStore.snapToClips(raw, props.clip.id)
      timelineStore.setSnapGuide(snapped !== raw ? snapped : null)
      timelineStore.updateClip(props.clip.id, { startTime: snapped })
    } else if (mode === 'left') {
      const raw = Math.max(0, startStart + dt)
      const snapped = timelineStore.snapToClips(raw, props.clip.id)
      timelineStore.setSnapGuide(snapped !== raw ? snapped : null)
      const delta = snapped - startStart
      const newTrimStart = Math.max(0, startTrimStart + delta)
      const newDuration = Math.max(0.1, startDuration - delta)
      timelineStore.updateClip(props.clip.id, {
        startTime: snapped,
        trimStart: newTrimStart,
        duration: newDuration
      })
    } else if (mode === 'right') {
      const rawEnd = startStart + startDuration + dt
      const snappedEnd = timelineStore.snapToClips(rawEnd, props.clip.id)
      const newEnd = Math.max(startStart + 0.1, snappedEnd)
      const newDuration = newEnd - startStart
      timelineStore.updateClip(props.clip.id, { duration: newDuration })
    }
  }
  const onUp = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    timelineStore.setSnapGuide(null)
    projectStore.markDirty()
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

function onKeyframeMouseDown(e, kf) {
  e.stopPropagation()
  if (e.button !== 0) return
  timelineStore.selectClip(props.clip.id, false)
  const startX = e.clientX
  const startTime = kf.time
  
  const onMove = (mv) => {
    const dt = (mv.clientX - startX) / props.zoom
    let newTime = startTime + dt
    newTime = Math.max(0, Math.min(props.clip.duration, newTime))
    timelineStore.updateKeyframe(props.clip.id, kf.id, { time: newTime })
  }
  const onUp = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
    projectStore.markDirty()
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

// Volume envelope preview: for audio-track clips with volume keyframes,
// sample the shared interpolator 24x across the clip and map volume
// 0..2 -> y 0..100% (inverted) for an SVG polyline overlay.
const volumeEnvPoints = computed(() => {
  if (props.trackType !== 'audio') return ''
  const kfs = props.clip.keyframes
  if (!kfs || !kfs.some(k => k.property === 'volume')) return ''
  const N = 24
  const pts = []
  for (let i = 0; i < N; i++) {
    const t = (i / (N - 1)) * props.clip.duration
    const v = getInterpolatedProperty(props.clip, 'volume', t, props.clip.volume ?? 1)
    const x = (i / (N - 1)) * 100
    const y = Math.min(100, Math.max(0, (1 - (v / 2)) * 100))
    pts.push(`${x.toFixed(2)},${y.toFixed(2)}`)
  }
  return pts.join(' ')
})
</script>

<template>
  <div
    class="absolute top-1 bottom-1 rounded overflow-hidden select-none group"
    :class="[timelineStore.activeTool === 'razor' ? 'cursor-crosshair' : 'cursor-grab active:cursor-grabbing', { 'ring-2 ring-accent z-10': selected }]"
    :style="{
      left: leftPx + 'px',
      width: widthPx + 'px',
      background: trackColor + '30',
      borderLeft: '2px solid ' + trackColor,
      borderRight: '2px solid ' + trackColor,
    }"
    @mousedown="(e) => onMouseDown(e, 'move')"
    @dblclick.stop="timelineStore.selectClip(clip.id)"
    @contextmenu.prevent="onContextMenu"
  >
    <div
      class="px-1.5 py-0.5 truncate text-[10px] font-medium pointer-events-none flex items-center gap-1"
      :style="{ color: trackColor }"
    >
      {{ fileName }}
      <div v-if="hasTransition || hasEffects" class="flex items-center gap-0.5 ml-1 opacity-80">
        <ArrowRightLeft v-if="hasTransition" :size="9" title="Transition applied" />
        <Palette v-if="hasEffects" :size="9" title="Effects applied" />
      </div>
    </div>

    <!-- Waveform (audio/video) -->
    <div
      v-if="asset?.waveform?.length"
      class="absolute left-0 right-0 bottom-0 top-4 flex items-center gap-px px-0.5 overflow-hidden pointer-events-none opacity-70"
    >
      <div
        v-for="(v, i) in asset.waveform.slice(0, Math.max(8, Math.floor(widthPx / 2)))"
        :key="i"
        class="flex-1 rounded-sm"
        :style="{ height: Math.max(2, Math.abs(v) * 100) + '%', background: trackColor }"
      />
    </div>

    <!-- Thumbnail strip (video) -->
    <div
      v-if="trackType === 'video' && asset?.thumbnail"
      class="absolute right-1 top-0.5 bottom-0.5 w-8 rounded-sm overflow-hidden pointer-events-none opacity-60"
    >
      <img :src="`data:image/jpeg;base64,${asset.thumbnail}`" class="w-full h-full object-cover" />
    </div>

    <!-- Volume envelope (audio clips with volume keyframes) -->
    <svg
      v-if="volumeEnvPoints"
      class="absolute inset-0 w-full h-full pointer-events-none opacity-70"
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
    >
      <polyline
        :points="volumeEnvPoints"
        fill="none"
        stroke="#00D4FF"
        stroke-width="1.5"
        vector-effect="non-scaling-stroke"
      />
    </svg>

    <!-- Keyframes -->
    <div v-if="clip.keyframes && clip.keyframes.length" class="absolute left-0 right-0 bottom-0 top-0 pointer-events-none">
      <div v-for="kf in clip.keyframes" :key="kf.id"
           class="absolute top-0 bottom-0 w-3 flex items-end justify-center pointer-events-auto cursor-ew-resize group/kf"
           :style="{ left: `calc(${kf.time * zoom}px - 6px)` }"
           @mousedown.stop="(e) => onKeyframeMouseDown(e, kf)"
           @dblclick.stop="timelineStore.removeKeyframe(clip.id, kf.id)"
           :title="`${kf.property}: ${kf.value}`">
        <div class="w-1.5 h-1.5 bg-accent rotate-45 mb-1 group-hover/kf:scale-150 transition-transform"></div>
      </div>
    </div>

    <!-- Edge handles -->
    <div
      class="absolute top-0 bottom-0 left-0 w-1.5 cursor-ew-resize hover:bg-white/30"
      @mousedown.stop="(e) => onMouseDown(e, 'left')"
    />
    <div
      class="absolute top-0 bottom-0 right-0 w-1.5 cursor-ew-resize hover:bg-white/30"
      @mousedown.stop="(e) => onMouseDown(e, 'right')"
    />

    <!-- Right-click context menu -->
    <Teleport to="body">
      <div
        v-if="showMenu"
        class="fixed z-50 min-w-40 bg-bg border border-border rounded-lg shadow-xl py-1 text-xs"
        :style="{ left: menuX + 'px', top: menuY + 'px' }"
        @click.stop
      >
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary flex items-center justify-between" @click="splitClip">
          <span>Split at Playhead</span>
          <span class="text-text-secondary text-[10px]">S</span>
        </button>
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary flex items-center justify-between" @click="duplicateClip">
          <span>Duplicate</span>
          <span class="text-text-secondary text-[10px]">Ctrl+D</span>
        </button>
        <div class="h-px bg-border my-1" />
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary flex items-center justify-between" @click="setSpeed(0.5)">
          <span>Half Speed</span>
          <span class="text-text-secondary text-[10px]">J</span>
        </button>
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary flex items-center justify-between" @click="setSpeed(2)">
          <span>Double Speed</span>
          <span class="text-text-secondary text-[10px]">L</span>
        </button>
        <div class="h-px bg-border my-1" />
        <button class="w-full text-left px-3 py-1.5 hover:bg-red-500/10 text-red-400 flex items-center justify-between" @click="deleteClip">
          <span>Delete</span>
          <span class="text-text-secondary text-[10px]">Del</span>
        </button>
      </div>
    </Teleport>
  </div>
</template>
