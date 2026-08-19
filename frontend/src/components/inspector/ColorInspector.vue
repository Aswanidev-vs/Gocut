<script setup>
import { ref, reactive, onMounted, watch, computed } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { usePlayerStore } from '../../stores/playerStore'
import { useUiStore } from '../../stores/uiStore'
import { OpenFilePicker, ImportLut } from '../../lib/wails'
import { Minus, Plus, RotateCcw, BarChart2, FileText, Upload } from 'lucide-vue-next'
import { computeSpline } from '../../lib/curves'

const timelineStore = useTimelineStore()
const playerStore = usePlayerStore()
const uiStore = useUiStore()
const activeTab = ref('basic')

const clip = computed(() => timelineStore.selectedClips[0])
const chromaKeyColor = computed(() => clip.value?.color?.chromaKeyColor || '')
const chromaKeySimilarity = computed(() => clip.value?.color?.chromaKeySimilarity || 0.01)
const chromaKeyBlend = computed(() => clip.value?.color?.chromaKeyBlend || 0.0)
const lutPath = computed(() => clip.value?.color?.lutPath || '')
const lutFileName = computed(() => {
  if (!lutPath.value) return ''
  return lutPath.value.split(/[\\/]/).pop() || lutPath.value
})
const lutTitle = computed(() => {
  if (!lutFileName.value) return ''
  return lutFileName.value.replace(/\.cube$/i, '')
})

async function importLut() {
  if (!clip.value) return
  try {
    const files = await OpenFilePicker([
      { name: '3D LUT Files (*.cube)', extensions: ['cube'] }
    ])
    if (!files || files.length === 0) return
    try {
      const res = await ImportLut(files[0])
      timelineStore.updateClipColor(clip.value.id, { lutPath: res.path })
      uiStore.addToast(`LUT imported: ${res.title || lutTitleFromPath(res.path)} (${res.size}³)`, 'success', 2000)
    } catch (err) {
      console.error('Failed to import LUT:', err)
      uiStore.addToast(`Failed to import LUT: ${typeof err === 'string' ? err : (err?.message || err)}`, 'error', 4000)
    }
  } catch (err) {
    console.error('Failed to import LUT:', err)
  }
}

function lutTitleFromPath(path) {
  if (!path) return 'LUT'
  const base = path.split(/[\\/]/).pop() || path
  return base.replace(/\.cube$/i, '') || 'LUT'
}

function clearLut() {
  if (!clip.value) return
  timelineStore.updateClipColor(clip.value.id, { lutPath: '' })
}

function updateChromaKey(key, value) {
  if (!clip.value) return
  timelineStore.updateClip(clip.value.id, {
    color: {
      ...clip.value.color,
      [key]: value,
    },
  })
}

const channelColors = {
  rgb: '#ffffff',
  r: '#ef4444',
  g: '#22c55e',
  b: '#3b82f6',
}

const channelLabels = ['rgb', 'r', 'g', 'b']
const channels = reactive({
  rgb: reactive({ points: [{ x: 0, y: 0 }, { x: 100, y: 100 }] }),
  r: reactive({ points: [{ x: 0, y: 0 }, { x: 100, y: 100 }] }),
  g: reactive({ points: [{ x: 0, y: 0 }, { x: 100, y: 100 }] }),
  b: reactive({ points: [{ x: 0, y: 0 }, { x: 100, y: 100 }] }),
})
const selectedChannel = ref('rgb')
const canvasRef = ref(null)
const dragging = ref(null)

const canvasSize = { w: 256, h: 256 }
const margin = 24

function nearestPoint(channel, mx, my) {
  const pts = channels[channel].points
  let best = 0
  let bestDist = Infinity
  pts.forEach((p, i) => {
    const d = Math.hypot(p.x - mx, p.y - my)
    if (d < bestDist) {
      bestDist = d
      best = i
    }
  })
  return bestDist < 18 ? best : -1
}

function clampPoint(p) {
  return {
    x: Math.max(0, Math.min(100, p.x)),
    y: Math.max(0, Math.min(100, p.y)),
  }
}

function getCanvasPos(e) {
  const canvas = canvasRef.value
  if (!canvas) return { x: 0, y: 0 }
  const rect = canvas.getBoundingClientRect()
  // Convert mouse position to 0-100 normalized coords
  // CSS may scale the canvas, so use rect dimensions
  const plotW = canvasSize.w - margin * 2
  const plotH = canvasSize.h - margin * 2
  // Mouse position relative to canvas element in CSS pixels
  const cssX = e.clientX - rect.left
  const cssY = e.clientY - rect.top
  // Convert CSS pixels to canvas pixels
  const canvasX = (cssX / rect.width) * canvasSize.w
  const canvasY = (cssY / rect.height) * canvasSize.h
  // Convert canvas pixels to 0-100 plot coords (Y inverted: canvas top = curve top = 100)
  const normX = ((canvasX - margin) / plotW) * 100
  const normY = (1 - (canvasY - margin) / plotH) * 100
  return { x: normX, y: normY }
}

function onMouseDown(e) {
  const pos = getCanvasPos(e)
  const clamped = clampPoint(pos)
  const idx = nearestPoint(selectedChannel.value, clamped.x, clamped.y)
  if (idx >= 0) {
    dragging.value = { idx, channel: selectedChannel.value }
  } else {
    // Click on empty area: add a new point
    const pts = channels[selectedChannel.value].points
    pts.push(clampPoint(pos))
    pts.sort((a, b) => a.x - b.x)
    const newIdx = pts.findIndex(p => p.x === clamped.x && p.y === clamped.y)
    dragging.value = { idx: newIdx >= 0 ? newIdx : pts.length - 1, channel: selectedChannel.value }
    draw()
    emitCurves()
  }
}

function onMouseMove(e) {
  if (!dragging.value) return
  const pos = getCanvasPos(e)
  const ch = dragging.value.channel
  const pts = channels[ch].points
  const p = clampPoint(pos)
  pts[dragging.value.idx] = p
  draw()
  emitCurves()
}

function onMouseUp() {
  if (dragging.value) {
    emitCurves()
  }
  dragging.value = null
}

function addPoint() {
  const pts = channels[selectedChannel.value].points
  const x = 50
  const before = pts.filter((p) => p.x <= x).sort((a, b) => a.x - b.x)[0]
  const after = pts.filter((p) => p.x >= x).sort((a, b) => a.x - b.x)[pts.length - 1]
  // linear blend between surrounding points
  const t = after && before && after.x !== before.x ? (x - before.x) / (after.x - before.x) : 0.5
  const y = before.y + (after.y - before.y) * t
  pts.push(clampPoint({ x, y: Math.max(5, Math.min(95, y)) }))
  pts.sort((a, b) => a.x - b.x)
  draw()
  emitCurves()
}

function removePoint() {
  const ch = selectedChannel.value
  const pts = channels[ch].points
  if (pts.length <= 2) return
  // remove the point closest to 50 unless min/max
  let remove = -1
  if (pts.length > 2) {
    let best = Infinity
    pts.forEach((p, i) => {
      if (i === 0 || i === pts.length - 1) return
      const d = Math.abs(p.x - 50) + Math.abs(p.y - 50)
      if (d < best) {
        best = d
        remove = i
      }
    })
  }
  if (remove >= 0) {
    pts.splice(remove, 1)
  }
  draw()
  emitCurves()
}

function resetCurves() {
  Object.keys(channels).forEach((k) => {
    channels[k].points = [{ x: 0, y: 0 }, { x: 100, y: 100 }]
  })
  draw()
  emitCurves()
}

function generateCurvesFilter() {
  // Check if any channel has been modified from default (0,0)-(100,100)
  let hasChanges = false
  for (const ch of channelLabels) {
    const pts = channels[ch].points
    if (pts.length !== 2 || pts[0].x !== 0 || pts[0].y !== 0 || pts[1].x !== 100 || pts[1].y !== 100) {
      hasChanges = true
      break
    }
  }
  if (!hasChanges) return ''

  // FFmpeg curves filter format: master='x/y x/y':red='x/y':green='x/y':blue='x/y'
  const channelMap = { rgb: 'master', r: 'red', g: 'green', b: 'blue' }
  const parts = channelLabels
    .map((ch) => {
      const pts = [...channels[ch].points].sort((a, b) => a.x - b.x)
      const path = pts
        .map((p) => `${(p.x / 100).toFixed(4)}/${(p.y / 100).toFixed(4)}`)
        .join(' ')
      return `${channelMap[ch]}='${path}'`
    })
  return parts.join(':')
}

function emitCurves() {
  if (!timelineStore.selectedClips.length) return
  const clip = timelineStore.selectedClips[0]
  timelineStore.updateClip(clip.id, {
    color: {
      ...clip.color,
      curves: generateCurvesFilter(),
    },
  })
}

function draw() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  ctx.clearRect(0, 0, canvasSize.w, canvasSize.h)

  // background grid
  ctx.strokeStyle = '#2A2A2A'
  ctx.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const x = margin + ((canvasSize.w - margin * 2) * i) / 4
    const y = margin + ((canvasSize.h - margin * 2) * i) / 4
    ctx.beginPath()
    ctx.moveTo(x, margin)
    ctx.lineTo(x, canvasSize.h - margin)
    ctx.stroke()
    ctx.beginPath()
    ctx.moveTo(margin, y)
    ctx.lineTo(canvasSize.w - margin, y)
    ctx.stroke()
  }

  // diagonal
  ctx.strokeStyle = '#444444'
  ctx.beginPath()
  ctx.moveTo(margin, margin)
  ctx.lineTo(canvasSize.w - margin, canvasSize.h - margin)
  ctx.stroke()

  function map(p) {
    return {
      x: margin + (p.x / 100) * (canvasSize.w - margin * 2),
      y: canvasSize.h - margin - (p.y / 100) * (canvasSize.h - margin * 2),
    }
  }

  channelLabels.forEach((ch) => {
    const pts = channels[ch].points
    if (pts.length === 0) return

    // Draw the smooth curve
    ctx.strokeStyle = channelColors[ch]
    ctx.lineWidth = ch === 'rgb' ? 1.5 : 1
    ctx.beginPath()
    
    const spline = computeSpline(pts)
    // Draw in 1px increments across the canvas
    const steps = canvasSize.w - margin * 2
    for (let i = 0; i <= steps; i++) {
      const x = (i / steps) * 100 // 0 to 100
      const y = spline(x)
      const mapped = map({ x, y: Math.max(0, Math.min(100, y)) })
      if (i === 0) ctx.moveTo(mapped.x, mapped.y)
      else ctx.lineTo(mapped.x, mapped.y)
    }
    ctx.stroke()

    // Draw the control points
    const mappedPts = pts.map(map)
    mappedPts.forEach((p) => {
      ctx.fillStyle = channelColors[ch]
      ctx.beginPath()
      ctx.arc(p.x, p.y, ch === 'rgb' ? 4 : 3, 0, Math.PI * 2)
      ctx.fill()
    })
  })
}

watch(selectedChannel, () => {
  draw()
})

const histogramCanvasRef = ref(null)
const showHistogram = ref(true)

function drawHistogram() {
  const base64 = playerStore.previewImage
  if (!base64) return

  const img = new Image()
  img.onload = () => {
    const canvas = histogramCanvasRef.value
    if (!canvas) return
    const hctx = canvas.getContext('2d')
    if (!hctx) return

    const offscreen = document.createElement('canvas')
    offscreen.width = 128
    offscreen.height = 72
    const octx = offscreen.getContext('2d')
    if (!octx) return
    octx.drawImage(img, 0, 0, 128, 72)
    
    const imgData = octx.getImageData(0, 0, 128, 72)
    const data = imgData.data

    const rHist = new Array(256).fill(0)
    const gHist = new Array(256).fill(0)
    const bHist = new Array(256).fill(0)
    const lHist = new Array(256).fill(0)

    for (let i = 0; i < data.length; i += 4) {
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      const l = Math.round(0.299 * r + 0.587 * g + 0.114 * b)

      rHist[r]++
      gHist[g]++
      bHist[b]++
      lHist[l]++
    }

    const W = canvas.width
    const H = canvas.height
    hctx.clearRect(0, 0, W, H)

    const maxVal = Math.max(...rHist, ...gHist, ...bHist, ...lHist, 1)

    // grid lines
    hctx.strokeStyle = '#2A2A2A'
    hctx.lineWidth = 1
    for (let i = 1; i < 4; i++) {
      const x = (W / 4) * i
      hctx.beginPath()
      hctx.moveTo(x, 0)
      hctx.lineTo(x, H)
      hctx.stroke()
    }

    const drawChannel = (hist, color) => {
      hctx.strokeStyle = color
      hctx.lineWidth = 1.5
      hctx.beginPath()
      for (let i = 0; i < 256; i++) {
        const x = (i / 255) * W
        const y = H - (hist[i] / maxVal) * H * 0.95
        if (i === 0) hctx.moveTo(x, y)
        else hctx.lineTo(x, y)
      }
      hctx.stroke()
    }

    hctx.globalCompositeOperation = 'screen'
    drawChannel(rHist, '#ef4444')
    drawChannel(gHist, '#22c55e')
    drawChannel(bHist, '#3b82f6')
    drawChannel(lHist, 'rgba(232, 232, 232, 0.4)')
    hctx.globalCompositeOperation = 'source-over'
  }
  img.src = 'data:image/jpeg;base64,' + base64
}

watch(() => playerStore.previewImage, () => {
  if (showHistogram.value) {
    drawHistogram()
  }
})

onMounted(() => {
  draw()
  drawHistogram()
})
</script>

<template>
  <div class="flex flex-col h-full gap-4">
    <!-- Curves section -->
    <div class="flex flex-col">
      <div class="flex items-center justify-between mb-2">
        <div class="text-xs font-semibold text-text-secondary">Curves</div>
        <div class="flex items-center gap-1">
          <button class="p-1 rounded hover:bg-border text-text-secondary hover:text-text-primary" @click="addPoint" title="Add point">
            <Plus :size="14" />
          </button>
          <button class="p-1 rounded hover:bg-border text-text-secondary hover:text-text-primary" @click="removePoint" title="Remove point">
            <Minus :size="14" />
          </button>
          <button class="p-1 rounded hover:bg-border text-text-secondary hover:text-text-primary" @click="resetCurves" title="Reset">
            <RotateCcw :size="14" />
          </button>
        </div>
      </div>
      <div class="flex items-center gap-2 mb-2">
        <button
          v-for="ch in channelLabels"
          :key="ch"
          class="text-[10px] px-2 py-0.5 rounded border transition-colors"
          :class="selectedChannel === ch ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary hover:text-text-primary'"
          @click="selectedChannel = ch"
        >
          {{ ch.toUpperCase() }}
        </button>
      </div>
      <div class="flex items-center justify-center bg-bg border border-border rounded p-1">
        <canvas
          ref="canvasRef"
          :width="canvasSize.w"
          :height="canvasSize.h"
          class="w-full h-auto cursor-crosshair select-none"
          style="image-rendering: auto"
          @mousedown="onMouseDown"
          @mousemove="onMouseMove"
          @mouseup="onMouseUp"
          @mouseleave="onMouseUp"
        />
      </div>
    </div>

    <!-- 3D LUT Section -->
    <div class="flex flex-col">
      <div class="flex items-center justify-between mb-2">
        <div class="text-xs font-semibold text-text-secondary flex items-center gap-1">
          <FileText :size="14" class="text-accent" />
          3D LUT Profile (.cube)
        </div>
      </div>
      <div v-if="!lutPath" class="flex items-center">
        <button
          class="flex items-center justify-center gap-1.5 text-xs px-3 py-1.5 rounded border border-dashed border-border hover:border-accent text-text-secondary hover:text-text-primary hover:bg-border/40 transition-colors w-full"
          @click="importLut"
        >
          <Upload :size="12" /> Import 3D LUT (.cube)
        </button>
      </div>
      <div v-else class="flex items-center justify-between bg-bg/80 border border-border rounded px-2.5 py-1.5">
        <div class="flex items-center gap-1.5 min-w-0">
          <FileText :size="12" class="text-accent shrink-0" />
          <div class="flex flex-col min-w-0">
            <span v-if="lutTitle" class="text-xs font-semibold text-text-primary truncate leading-tight">{{ lutTitle }}</span>
            <span class="text-[10px] font-mono text-text-secondary truncate leading-tight">{{ lutFileName }}</span>
          </div>
        </div>
        <button
          class="text-[10px] px-2 py-0.5 rounded border border-border text-text-secondary hover:text-red-400 hover:border-red-400 transition-colors shrink-0"
          @click="clearLut"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Chroma Key Section -->
    <div class="flex flex-col">
      <div class="flex items-center justify-between mb-2">
        <div class="text-xs font-semibold text-text-secondary">Chroma Key</div>
      </div>
      <div class="flex items-center gap-2 mb-2">
        <button
          v-if="!chromaKeyColor"
          class="text-xs px-3 py-1.5 rounded border border-border text-text-secondary hover:text-text-primary hover:bg-border transition-colors w-full"
          @click="updateChromaKey('chromaKeyColor', '#00ff00')"
        >
          Enable Chroma Key
        </button>
        <template v-else>
          <input
            type="color"
            :value="chromaKeyColor"
            @input="updateChromaKey('chromaKeyColor', $event.target.value)"
            class="w-8 h-8 rounded cursor-pointer border border-border p-0"
          />
          <button
            class="text-[10px] px-2 py-1 rounded border border-border text-text-secondary hover:text-text-primary"
            @click="updateChromaKey('chromaKeyColor', '')"
          >
            Clear
          </button>
        </template>
      </div>
      <div v-if="chromaKeyColor" class="flex flex-col gap-2">
        <div>
          <div class="flex justify-between text-[10px] text-text-secondary mb-1">
            <span>Similarity</span>
            <span class="font-mono">{{ chromaKeySimilarity.toFixed(2) }}</span>
          </div>
          <input
            type="range"
            min="0.01" max="1.0" step="0.01"
            :value="chromaKeySimilarity"
            @input="updateChromaKey('chromaKeySimilarity', parseFloat($event.target.value))"
            class="w-full accent-accent"
          />
        </div>
        <div>
          <div class="flex justify-between text-[10px] text-text-secondary mb-1">
            <span>Blend</span>
            <span class="font-mono">{{ chromaKeyBlend.toFixed(2) }}</span>
          </div>
          <input
            type="range"
            min="0.0" max="1.0" step="0.01"
            :value="chromaKeyBlend"
            @input="updateChromaKey('chromaKeyBlend', parseFloat($event.target.value))"
            class="w-full accent-accent"
          />
        </div>
      </div>
    </div>

    <!-- Histogram Section -->
    <div class="flex flex-col">
      <div class="flex items-center justify-between mb-2">
        <div class="text-xs font-semibold text-text-secondary flex items-center gap-1">
          <BarChart2 :size="14" class="text-accent" />
          Real-time Histogram
        </div>
        <button
          class="text-[10px] px-1.5 py-0.5 rounded border border-border text-text-secondary hover:text-text-primary"
          @click="showHistogram = !showHistogram"
        >
          {{ showHistogram ? 'Hide' : 'Show' }}
        </button>
      </div>
      
      <div
        v-show="showHistogram"
        class="flex items-center justify-center bg-bg/80 border border-border rounded p-2 relative h-32"
      >
        <canvas
          ref="histogramCanvasRef"
          width="256"
          height="128"
          class="w-full h-full select-none"
        />
        <div
          v-if="!playerStore.previewImage"
          class="absolute inset-0 flex items-center justify-center text-[10px] text-text-secondary bg-bg/90"
        >
          Import media / play timeline to analyze colors
        </div>
      </div>
    </div>
  </div>
</template>
