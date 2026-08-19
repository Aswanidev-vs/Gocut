<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useDesignStore, getNodeType, EASING_TYPES } from '../../stores/designStore'
import { Zap, CornerDownRight, Maximize2, Sliders } from 'lucide-vue-next'

const props = defineProps({ playheadTime: { type: Number, default: 0 } })
const emit = defineEmits(['seek'])
const designStore = useDesignStore()

const canvasRef = ref(null)

const node = computed(() => designStore.selectedNode)
const paramList = computed(() => {
  if (!node.value || !node.value.keyframes) return []
  return Object.entries(node.value.keyframes).filter(([, kfs]) => kfs && kfs.length > 0)
})
const activeParam = ref(null)

watch(paramList, (list) => {
  if (list.length && !activeParam.value) activeParam.value = list[0][0]
})

// Palette for multi-curve tracks
const PARAM_COLORS = ['#00D4FF', '#EC4899', '#10B981', '#F59E0B', '#8B5CF6', '#F43F5E']

function setEasingForSelected(easing) {
  if (!node.value || !activeParam.value) return
  const kfs = node.value.keyframes[activeParam.value]
  if (!kfs) return
  for (const kf of kfs) {
    kf.easing = easing
  }
  draw()
}

function draw() {
  const c = canvasRef.value
  if (!c) return
  const c2d = c.getContext('2d')
  if (!c2d) return
  const W = c.width, H = c.height
  c2d.clearRect(0, 0, W, H)

  // Background
  c2d.fillStyle = '#0B0B12'
  c2d.fillRect(0, 0, W, H)

  if (!node.value || !activeParam.value || !node.value.keyframes[activeParam.value]) return
  const kfs = node.value.keyframes[activeParam.value]
  if (kfs.length < 2) {
    c2d.fillStyle = '#666'
    c2d.font = '10px sans-serif'
    c2d.textAlign = 'center'
    c2d.fillText('Add at least 2 keyframes in Inspector to view Spline curves', W / 2, H / 2)
    return
  }

  const kfsSorted = [...kfs].sort((a, b) => a.time - b.time)
  const minT = Math.max(0, kfsSorted[0].time - 0.2)
  const maxT = Math.max(minT + 0.5, kfsSorted[kfsSorted.length - 1].time + 0.2)
  const tRange = Math.max(0.1, maxT - minT)

  // Find value range with padding
  const allVals = kfsSorted.map(k => k.value)
  const rawMin = Math.min(...allVals), rawMax = Math.max(...allVals)
  const diff = rawMax - rawMin || 1
  const minV = rawMin - diff * 0.1
  const maxV = rawMax + diff * 0.1
  const vRange = Math.max(0.01, maxV - minV)

  const padLeft = 36, padRight = 16, padTop = 16, padBottom = 20
  const plotW = W - padLeft - padRight
  const plotH = H - padTop - padBottom

  function toScreen(t, v) {
    return {
      x: padLeft + ((t - minT) / tRange) * plotW,
      y: padTop + (1 - (v - minV) / vRange) * plotH,
    }
  }

  // Grid lines & Time numbers
  c2d.strokeStyle = '#181824'
  c2d.lineWidth = 1
  c2d.fillStyle = '#555566'
  c2d.font = '8px monospace'

  for (let i = 0; i <= 4; i++) {
    const y = padTop + (i / 4) * plotH
    c2d.beginPath(); c2d.moveTo(padLeft, y); c2d.lineTo(W - padRight, y); c2d.stroke()
    const val = (maxV - (i / 4) * vRange).toFixed(1)
    c2d.textAlign = 'right'
    c2d.fillText(val, padLeft - 4, y + 3)
  }

  for (let i = 0; i <= 8; i++) {
    const x = padLeft + (i / 8) * plotW
    c2d.beginPath(); c2d.moveTo(x, padTop); c2d.lineTo(x, H - padBottom); c2d.stroke()
    const tVal = (minT + (i / 8) * tRange).toFixed(2) + 's'
    c2d.textAlign = 'center'
    c2d.fillText(tVal, x, H - 6)
  }

  // Playhead line
  if (props.playheadTime >= minT && props.playheadTime <= maxT) {
    const ph = toScreen(props.playheadTime, minV)
    c2d.strokeStyle = '#F43F5E'
    c2d.lineWidth = 1.5
    c2d.beginPath(); c2d.moveTo(ph.x, padTop); c2d.lineTo(ph.x, H - padBottom); c2d.stroke()
  }

  // Smooth Bezier Curve Drawing (Fusion style)
  c2d.strokeStyle = '#00D4FF'
  c2d.lineWidth = 2.5
  c2d.beginPath()

  const stepCount = 120
  for (let i = 0; i <= stepCount; i++) {
    const t = minT + (i / stepCount) * tRange
    let vi = minV

    if (t <= kfsSorted[0].time) {
      vi = kfsSorted[0].value
    } else if (t >= kfsSorted[kfsSorted.length - 1].time) {
      vi = kfsSorted[kfsSorted.length - 1].value
    } else {
      for (let j = 0; j < kfsSorted.length - 1; j++) {
        const a = kfsSorted[j], b = kfsSorted[j + 1]
        if (t >= a.time && t <= b.time) {
          const ratio = (t - a.time) / (b.time - a.time)
          let eased = ratio
          if (a.easing === 'smooth' || a.easing === 'easeInOut') {
            eased = ratio < 0.5 ? 2 * ratio * ratio : 1 - Math.pow(-2 * ratio + 2, 2) / 2
          } else if (a.easing === 'easeIn') {
            eased = ratio * ratio
          } else if (a.easing === 'easeOut') {
            eased = 1 - (1 - ratio) * (1 - ratio)
          }
          vi = a.value + (b.value - a.value) * eased
          break
        }
      }
    }

    const p = toScreen(t, vi)
    if (i === 0) c2d.moveTo(p.x, p.y)
    else c2d.lineTo(p.x, p.y)
  }
  c2d.stroke()

  // Keyframe handles & nodes
  for (const kf of kfsSorted) {
    const p = toScreen(kf.time, kf.value)
    
    // Draw Fusion Diamond Keyframe point
    c2d.fillStyle = '#00D4FF'
    c2d.strokeStyle = '#FFFFFF'
    c2d.lineWidth = 1.5
    c2d.beginPath()
    c2d.moveTo(p.x, p.y - 5)
    c2d.lineTo(p.x + 5, p.y)
    c2d.lineTo(p.x, p.y + 5)
    c2d.lineTo(p.x - 5, p.y)
    c2d.closePath()
    c2d.fill()
    c2d.stroke()
  }
}

onMounted(() => {
  const resize = () => {
    const c = canvasRef.value
    if (!c || !c.parentElement) return
    c.width = c.parentElement.clientWidth
    c.height = c.parentElement.clientHeight
    draw()
  }
  resize()
  window.addEventListener('resize', resize)
})

watch(() => [designStore.selectedNodeId, activeParam.value, designStore.nodes], draw, { deep: true })
watch(() => props.playheadTime, draw)

function canvasClick(e) {
  const c = canvasRef.value
  if (!c) return
  const rect = c.getBoundingClientRect()
  const mx = e.clientX - rect.left
  const padLeft = 36, padRight = 16
  const pct = Math.max(0, Math.min(1, (mx - padLeft) / (c.width - padLeft - padRight)))
  if (!node.value || !activeParam.value || !node.value.keyframes[activeParam.value]) return
  const kfs = node.value.keyframes[activeParam.value].sort((a, b) => a.time - b.time)
  const minT = Math.max(0, kfs[0].time - 0.2)
  const maxT = Math.max(minT + 0.5, kfs[kfs.length - 1].time + 0.2)
  const tRange = Math.max(0.1, maxT - minT)
  const t = minT + pct * tRange
  emit('seek', t)
}
</script>

<template>
  <div class="flex flex-col h-full bg-[#090910] border-t border-border/80 select-none">
    <!-- Fusion Spline Toolbar -->
    <div class="h-7 bg-[#12121A] border-b border-border/70 flex items-center px-2 gap-2 text-[10px] text-text-secondary">
      <div class="flex items-center gap-1 font-semibold text-text-primary">
        <Zap :size="12" class="text-accent" />
        <span class="uppercase tracking-wider text-[9px]">Spline Editor</span>
      </div>

      <span v-if="node" class="text-text-primary font-medium">· {{ node.label }}</span>

      <!-- Parameter Selector -->
      <select
        v-if="paramList.length"
        v-model="activeParam"
        class="bg-[#1C1C28] border border-border/80 rounded px-1.5 py-0.5 text-[10px] text-text-primary outline-none font-mono"
      >
        <option v-for="[pid] in paramList" :key="pid" :value="pid">{{ pid }}</option>
      </select>

      <div class="flex-1" />

      <!-- Fusion Spline Quick Actions: Smooth (S), Linear (L), Ease -->
      <div v-if="paramList.length" class="flex items-center gap-1">
        <button
          class="px-2 py-0.5 rounded bg-[#1C1C28] hover:bg-accent/20 border border-border/60 text-[9px] font-mono text-text-primary transition-colors"
          title="Smooth Curves (S)"
          @click="setEasingForSelected('smooth')"
        >
          Smooth [S]
        </button>
        <button
          class="px-2 py-0.5 rounded bg-[#1C1C28] hover:bg-accent/20 border border-border/60 text-[9px] font-mono text-text-primary transition-colors"
          title="Linear Tangents (L)"
          @click="setEasingForSelected('linear')"
        >
          Linear [L]
        </button>
        <button
          class="px-2 py-0.5 rounded bg-[#1C1C28] hover:bg-accent/20 border border-border/60 text-[9px] font-mono text-text-primary transition-colors"
          title="Ease In/Out"
          @click="setEasingForSelected('easeInOut')"
        >
          Ease In/Out
        </button>
      </div>
    </div>

    <div class="flex-1 relative overflow-hidden">
      <canvas
        ref="canvasRef"
        class="w-full h-full cursor-crosshair"
        @click="canvasClick"
      />
      <div v-if="!node || paramList.length === 0" class="absolute inset-0 flex items-center justify-center text-[10px] text-text-secondary/50">
        Select a node with keyframed parameters to edit animation splines
      </div>
    </div>
  </div>
</template>