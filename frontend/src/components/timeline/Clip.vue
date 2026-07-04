<script setup>
import { computed, ref } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useTimelineStore } from '../../stores/timelineStore'
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
  timelineStore.splitClipAt(props.clip.id, props.clip.startTime + props.clip.duration / 2)
  closeMenu()
}
function duplicateClip() {
  const dup = { ...props.clip, id: undefined, startTime: props.clip.startTime + props.clip.duration + 0.1 }
  timelineStore.addClipToTrack(props.clip._trackIndex, dup)
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

function onMouseDown(e, mode) {
  e.stopPropagation()
  if (e.button !== 0) return
  timelineStore.selectClip(props.clip.id, e.shiftKey)

  const startX = e.clientX
  const startStart = props.clip.startTime
  const startDuration = props.clip.duration
  const startTrimStart = props.clip.trimStart

  const onMove = (mv) => {
    const dx = mv.clientX - startX
    const dt = dx / props.zoom
    if (mode === 'move') {
      const newStart = timelineStore.snapToClips(Math.max(0, startStart + dt), props.clip.id)
      timelineStore.updateClip(props.clip.id, { startTime: newStart })
    } else if (mode === 'left') {
      const newStart = timelineStore.snapToClips(Math.max(0, startStart + dt), props.clip.id)
      const delta = newStart - startStart
      const newTrimStart = Math.max(0, startTrimStart + delta)
      const newDuration = Math.max(0.1, startDuration - delta)
      timelineStore.updateClip(props.clip.id, {
        startTime: newStart,
        trimStart: newTrimStart,
        duration: newDuration
      })
    } else if (mode === 'right') {
      const newEnd = Math.max(props.clip.startTime + 0.1, startStart + startDuration + dt)
      const newDuration = newEnd - props.clip.startTime
      timelineStore.updateClip(props.clip.id, { duration: newDuration })
    }
  }
  const onUp = () => {
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
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
</script>

<template>
  <div
    class="absolute top-1 bottom-1 rounded overflow-hidden cursor-grab active:cursor-grabbing select-none group"
    :style="{
      left: leftPx + 'px',
      width: widthPx + 'px',
      background: trackColor + '30',
      borderLeft: '2px solid ' + trackColor,
      borderRight: '2px solid ' + trackColor,
    }"
    :class="{ 'ring-2 ring-accent z-10': selected }"
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
        class="fixed z-50 min-w-36 bg-bg border border-border rounded-lg shadow-xl py-1 text-xs"
        :style="{ left: menuX + 'px', top: menuY + 'px' }"
        @click.stop
      >
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary" @click="splitClip">Split at Playhead</button>
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary" @click="duplicateClip">Duplicate</button>
        <div class="h-px bg-border my-1" />
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary" @click="setSpeed(0.5)">Half Speed</button>
        <button class="w-full text-left px-3 py-1.5 hover:bg-accent/10 text-text-primary" @click="setSpeed(2)">Double Speed</button>
        <div class="h-px bg-border my-1" />
        <button class="w-full text-left px-3 py-1.5 hover:bg-red-500/10 text-red-400" @click="deleteClip">Delete</button>
      </div>
    </Teleport>
  </div>
</template>
