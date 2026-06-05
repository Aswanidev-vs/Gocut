<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { usePlayerStore } from '../../stores/playerStore'
import { Minus, Plus, RotateCcw, BarChart2 } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const playerStore = usePlayerStore()
const activeTab = ref('basic')

type Preset = { name: string; color: any }

const channelColors: Record<string, string> = {
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
const dragging = ref<{ idx: number; channel: string } | null>(null)

const canvasSize = { w: 256, h: 256 }
const margin = 24

type Point = { x: number; y: number }

function nearestPoint(channel: string, mx: number, my: number): number {
  const pts = (channels as any)[channel].points as Point[]
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

function clampPoint(p: Point): Point {
  return {
    x: Math.max(0, Math.min(100, p.x)),
    y: Math.max(0, Math.min(100, p.y)),
  }
}

function getCanvasPos(e: MouseEvent): Point {
  const canvas = canvasRef.value
  if (!canvas) return { x: 0, y: 0 }
  const rect = canvas.getBoundingClientRect()
  const scaleX = (canvasSize.w - margin * 2) / rect.width
  const scaleY = (canvasSize.h - margin * 2) / rect.height
  return {
    x: (e.clientX - rect.left - margin) * scaleX,
    y: (e.clientY - rect.top - margin) * scaleY,
  }
}

function onMouseDown(e: MouseEvent) {
  const pos = getCanvasPos(e)
  const idx = nearestPoint(selectedChannel.value, pos.x, pos.y)
  if (idx >= 0) {
    dragging.value = { idx, channel: selectedChannel.value }
  }
}

function onMouseMove(e: MouseEvent) {
  if (!dragging.value) return
  const pos = getCanvasPos(e)
  const ch = dragging.value.channel
  const pts = (channels as any)[ch].points as Point[]
  const p = clampPoint(pos)
  pts[dragging.value.idx] = p
  draw()
}

function onMouseUp() {
  dragging.value = null
}

function addPoint() {
  const pts = (channels as any)[selectedChannel.value].points as Point[]
  const x = 50
  const before = pts.filter((p: Point) => p.x <= x).sort((a: Point, b: Point) => a.x - b.x)[0]
  const after = pts.filter((p: Point) => p.x >= x).sort((a: Point, b: Point) => a.x - b.x)[pts.length - 1]
  // linear blend between surrounding points
  const t = after && before && after.x !== before.x ? (x - before.x) / (after.x - before.x) : 0.5
  const y = before.y + (after.y - before.y) * t
  pts.push(clampPoint({ x, y: Math.max(5, Math.min(95, y)) }))
  pts.sort((a: Point, b: Point) => a.x - b.x)
  draw()
  emitCurves()
}

function removePoint() {
  const ch = selectedChannel.value
  const pts = (channels as any)[ch].points as Point[]
  if (pts.length <= 2) return
  // remove the point closest to 50 unless min/max
  let remove = -1
  if (pts.length > 2) {
    let best = Infinity
    pts.forEach((p: Point, i: number) => {
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
    ;(channels as any)[k].points = [{ x: 0, y: 0 }, { x: 100, y: 100 }]
  })
  draw()
  emitCurves()
}

function generateCurvesFilter() {
  const spec = channelLabels
    .map((ch) => {
      const pts = (channels as any)[ch].points as Point[]
      const path = pts
        .sort((a: Point, b: Point) => a.x - b.x)
        .map((p: Point) => `${(p.x / 100).toFixed(4)}/${(p.y / 100).toFixed(4)}`)
        .join(' ')
      return `${ch}='${path}'`
    })
    .join(' ')
  return spec
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

  function map(p: Point) {
    return {
      x: margin + (p.x / 100) * (canvasSize.w - margin * 2),
      y: canvasSize.h - margin - (p.y / 100) * (canvasSize.h - margin * 2),
    }
  }

  channelLabels.forEach((ch) => {
    const pts = (channels as any)[ch].points as Point[]
    const mapped = pts.map(map)
    ctx.strokeStyle = channelColors[ch]
    ctx.lineWidth = ch === 'rgb' ? 1.5 : 1
    ctx.beginPath()
    mapped.forEach((p: Point, i: number) => {
      if (i === 0) ctx.moveTo(p.x, p.y)
      else ctx.lineTo(p.x, p.y)
    })
    ctx.stroke()

    mapped.forEach((p: Point) => {
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
