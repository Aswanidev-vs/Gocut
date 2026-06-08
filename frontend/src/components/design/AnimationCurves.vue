<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useDesignStore, getNodeType, EASING_TYPES } from '../../stores/designStore'

const props = defineProps({ playheadTime: { type: Number, default: 0 } })
const emit = defineEmits(['seek'])
const designStore = useDesignStore()

const canvasRef = ref(null)

const node = computed(() => designStore.selectedNode)
const paramList = computed(() => {
  if (!node.value || !node.value.keyframes) return []
  return Object.entries(node.value.keyframes).filter(([, kfs]) => kfs.length > 0)
})
const activeParam = ref(null)

watch(paramList, (list) => {
  if (list.length && !activeParam.value) activeParam.value = list[0][0]
})

function draw() {
  const c = canvasRef.value
  if (!c) return
  const c2d = c.getContext('2d')
  if (!c2d) return
  const W = c.width, H = c.height
  c2d.clearRect(0, 0, W, H)

  // Background
  c2d.fillStyle = '#111118'
  c2d.fillRect(0, 0, W, H)

  if (!node.value || !activeParam.value || !node.value.keyframes[activeParam.value]) return
  const kfs = node.value.keyframes[activeParam.value]
  if (kfs.length < 2) {
    c2d.fillStyle = '#555'
    c2d.font = '11px sans-serif'
    c2d.textAlign = 'center'
    c2d.fillText('Add at least 2 keyframes to see curve', W / 2, H / 2)
    return
  }

  const compDur = designStore.composition.duration || 5
  const kfsSorted = [...kfs].sort((a, b) => a.time - b.time)
  const minT = kfsSorted[0].time, maxT = kfsSorted[kfsSorted.length - 1].time
  const tRange = Math.max(0.1, maxT - minT)
  // Find value range
  const allVals = kfsSorted.map(k => k.value)
  const minV = Math.min(...allVals), maxV = Math.max(...allVals)
  const vRange = Math.max(0.1, maxV - minV)
  const pad = 16
  const plotW = W - pad * 2, plotH = H - pad * 2

  function toScreen(t, v) {
    return {
      x: pad + ((t - minT) / tRange) * plotW,
      y: pad + (1 - (v - minV) / vRange) * plotH,
    }
  }

  // Grid lines
  c2d.strokeStyle = '#222'
  c2d.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const y = pad + (i / 4) * plotH
    c2d.beginPath(); c2d.moveTo(pad, y); c2d.lineTo(W - pad, y); c2d.stroke()
  }
  for (let i = 0; i <= 8; i++) {
    const x = pad + (i / 8) * plotW
    c2d.beginPath(); c2d.moveTo(x, pad); c2d.lineTo(x, H - pad); c2d.stroke()
  }

  // Playhead line
  if (props.playheadTime >= minT && props.playheadTime <= maxT) {
    const ph = toScreen(props.playheadTime, minV)
    c2d.strokeStyle = '#F59E0B'
    c2d.lineWidth = 1
    c2d.beginPath(); c2d.moveTo(ph.x, pad); c2d.lineTo(ph.x, H - pad); c2d.stroke()
  }

  // Draw curve
  c2d.strokeStyle = '#00D4FF'
  c2d.lineWidth = 2
  c2d.beginPath()
  for (let i = 0; i <= 100; i++) {
    const t = minT + (i / 100) * tRange
    // Linear interpolation
    let vi = minV
    for (let j = 0; j < kfsSorted.length - 1; j++) {
      const a = kfsSorted[j], b = kfsSorted[j + 1]
      if (t >= a.time && t <= b.time) {
        const pct = a.time === b.time ? 0 : (t - a.time) / (b.time - a.time)
        vi = a.value + (b.value - a.value) * pct
        break
      }
    }
    const p = toScreen(t, vi)
    if (i === 0) c2d.moveTo(p.x, p.y)
    else c2d.lineTo(p.x, p.y)
  }
  c2d.stroke()

  // Draw keyframe points
  for (const kf of kfsSorted) {
    const p = toScreen(kf.time, kf.value)
    c2d.fillStyle = node.value?.id === designStore.selectedNodeId ? '#F59E0B' : '#888'
    c2d.beginPath()
    c2d.arc(p.x, p.y, 4, 0, Math.PI * 2)
    c2d.fill()
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
  const pct = Math.max(0, Math.min(1, (mx - 16) / (c.width - 32)))
  if (!node.value || !activeParam.value || !node.value.keyframes[activeParam.value]) return
  const kfs = node.value.keyframes[activeParam.value].sort((a, b) => a.time - b.time)
  const minT = kfs[0].time, maxT = kfs[kfs.length - 1].time
  const tRange = Math.max(0.1, maxT - minT)
  const t = minT + pct * tRange
  emit('seek', t)
}
</script>

<template>
  <div class="flex flex-col h-full bg-[#0D0D15]">
    <div class="h-6 bg-panel/80 border-b border-border flex items-center px-2 gap-2 text-[10px]">
      <span class="text-text-secondary uppercase tracking-wider">Curves</span>
      <span v-if="node" class="text-text-primary">· {{ node.label }}</span>
      <select
        v-if="paramList.length"
        v-model="activeParam"
        class="ml-auto bg-bg border border-border rounded px-1.5 py-0.5 text-[10px] text-text-primary outline-none"
      >
        <option v-for="[pid] in paramList" :key="pid" :value="pid">{{ pid }}</option>
      </select>
    </div>
    <div class="flex-1 relative">
      <canvas
        ref="canvasRef"
        class="w-full h-full cursor-pointer"
        @click="canvasClick"
      />
      <div v-if="!node" class="absolute inset-0 flex items-center justify-center text-[10px] text-text-secondary/50">
        Select a node with keyframes
      </div>
    </div>
  </div>
</template>