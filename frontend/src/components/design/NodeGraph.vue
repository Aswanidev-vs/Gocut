<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useDesignStore, getNodeType, NODE_TYPES } from '../../stores/designStore'
import { X, Eye, EyeOff, Lock, Unlock } from 'lucide-vue-next'

const props = defineProps({ playheadTime: { type: Number, default: 0 }, isPlaying: { type: Boolean, default: false } })
const designStore = useDesignStore()

const canvasRef = ref(null)
const ctx = ref(null)
const isDragging = ref(false)
const dragNode = ref(null)
const dragOffset = { x: 0, y: 0 }
const isPanning = ref(false)
const panStart = { x: 0, y: 0 }

function worldToScreen(wx, wy) {
  return { x: (wx + designStore.panX) * designStore.zoom, y: (wy + designStore.panY) * designStore.zoom }
}
function screenToWorld(sx, sy) {
  return { x: sx / designStore.zoom - designStore.panX, y: sy / designStore.zoom - designStore.panY }
}

function draw() {
  const c = canvasRef.value
  if (!c) return
  const c2d = c.getContext('2d')
  if (!c2d) return
  const W = c.width, H = c.height
  c2d.clearRect(0, 0, W, H)

  // Background grid
  c2d.strokeStyle = '#1a1a1a'
  c2d.lineWidth = 1
  const gridSize = 30 * designStore.zoom
  const ox = designStore.panX * designStore.zoom % gridSize
  const oy = designStore.panY * designStore.zoom % gridSize
  for (let x = ox; x < W; x += gridSize) { c2d.beginPath(); c2d.moveTo(x, 0); c2d.lineTo(x, H); c2d.stroke() }
  for (let y = oy; y < H; y += gridSize) { c2d.beginPath(); c2d.moveTo(0, y); c2d.lineTo(W, y); c2d.stroke() }

  // Draw connections
  c2d.strokeStyle = '#4a4a4a'
  c2d.lineWidth = 2
  for (const conn of designStore.connections) {
    const from = designStore.nodes.find(n => n.id === conn.fromNode)
    const to = designStore.nodes.find(n => n.id === conn.toNode)
    if (!from || !to) continue
    const f = worldToScreen(from.x + 160, from.y + 30)
    const t = worldToScreen(to.x, to.y + 30)
    c2d.beginPath(); c2d.moveTo(f.x, f.y)
    const cp1x = f.x + (t.x - f.x) * 0.5
    c2d.bezierCurveTo(cp1x, f.y, cp1x, t.y, t.x, t.y)
    c2d.stroke()
  }

  // Draw nodes
  for (const node of designStore.nodes) {
    if (!node.visible) continue
    const type = getNodeType(node.type)
    if (!type) continue
    const pos = worldToScreen(node.x, node.y)
    const nw = 160, nh = 60
    const isSelected = designStore.selectedNodeId === node.id

    // Shadow
    c2d.shadowColor = isSelected ? type.color + '40' : 'rgba(0,0,0,0.3)'
    c2d.shadowBlur = isSelected ? 16 : 6

    // Node bg
    c2d.fillStyle = '#1E1E2E'
    c2d.strokeStyle = isSelected ? type.color : '#3a3a4a'
    c2d.lineWidth = isSelected ? 2 : 1
    c2d.beginPath()
    c2d.roundRect(pos.x, pos.y, nw, nh, 6)
    c2d.fill()
    c2d.stroke()
    c2d.shadowBlur = 0

    // Header bar
    c2d.fillStyle = type.color + '30'
    c2d.beginPath()
    c2d.roundRect(pos.x, pos.y, nw, 20, { upperLeft: 6, upperRight: 6 })
    c2d.fill()

    // Title
    c2d.fillStyle = '#E8E8E8'
    c2d.font = '10px DM Sans, sans-serif'
    c2d.fillText(node.label || type.label, pos.x + 8, pos.y + 14)

    // Input port dot
    if (type.inputs.length > 0) {
      c2d.fillStyle = type.color
      c2d.beginPath()
      c2d.arc(pos.x - 5, pos.y + 30, 6, 0, Math.PI * 2)
      c2d.fill()
    }

    // Output port dot
    if (type.outputs.length > 0) {
      c2d.fillStyle = type.color
      c2d.beginPath()
      c2d.arc(pos.x + nw + 5, pos.y + 30, 6, 0, Math.PI * 2)
      c2d.fill()
    }

    // Keyframe indicator
    const hasKeyframes = Object.keys(node.keyframes).length > 0 && Object.values(node.keyframes).some(k => k.length > 0)
    if (hasKeyframes) {
      c2d.fillStyle = '#F59E0B'
      c2d.font = '9px sans-serif'
      c2d.fillText('◆', pos.x + nw - 22, pos.y + 15)
    }
  }
}

onMounted(() => {
  const c = canvasRef.value
  if (!c) return
  const resize = () => {
    c.width = c.parentElement?.clientWidth || 400
    c.height = c.parentElement?.clientHeight || 300
    draw()
  }
  resize()
  window.addEventListener('resize', resize)
  const obs = c.parentElement ? new ResizeObserver(resize) : null
  if (obs) obs.observe(c.parentElement)

  return () => { window.removeEventListener('resize', resize); obs?.disconnect() }
})

watch(() => [designStore.nodes, designStore.connections, designStore.selectedNodeId, designStore.zoom, designStore.panX, designStore.panY], () => requestAnimationFrame(draw), { deep: true })
watch(() => props.playheadTime, () => { if (props.isPlaying) requestAnimationFrame(draw) })

function onMouseDown(e) {
  const rect = canvasRef.value.getBoundingClientRect()
  const mx = e.clientX - rect.left, my = e.clientY - rect.top
  const w = screenToWorld(mx, my)

  // Hit test nodes (backwards for z-order)
  for (let i = designStore.nodes.length - 1; i >= 0; i--) {
    const n = designStore.nodes[i]
    if (!n.visible) continue
    const pos = worldToScreen(n.x, n.y)
    if (mx >= pos.x && mx <= pos.x + 160 && my >= pos.y && my <= pos.y + 60) {
      designStore.selectedNodeId = n.id
      isDragging.value = true
      dragNode.value = n
      dragOffset.x = w.x - n.x
      dragOffset.y = w.y - n.y
      return
    }
  }

  // Click on empty space
  designStore.selectedNodeId = null
  isPanning.value = true
  panStart.x = e.clientX - designStore.panX * designStore.zoom
  panStart.y = e.clientY - designStore.panY * designStore.zoom
}

function onMouseMove(e) {
  if (isDragging.value && dragNode.value) {
    const rect = canvasRef.value.getBoundingClientRect()
    const w = screenToWorld(e.clientX - rect.left, e.clientY - rect.top)
    dragNode.value.x = Math.max(0, w.x - dragOffset.x)
    dragNode.value.y = Math.max(0, w.y - dragOffset.y)
    requestAnimationFrame(draw)
  } else if (isPanning.value) {
    designStore.panX = (e.clientX - panStart.x) / designStore.zoom
    designStore.panY = (e.clientY - panStart.y) / designStore.zoom
    requestAnimationFrame(draw)
  }
}

function onMouseUp() {
  isDragging.value = false; dragNode.value = null
  isPanning.value = false
}

function onWheel(e) {
  e.preventDefault()
  const delta = e.deltaY > 0 ? 0.9 : 1.1
  designStore.zoom = Math.max(0.1, Math.min(5, designStore.zoom * delta))
  requestAnimationFrame(draw)
}

function onDblClick(e) {
  const rect = canvasRef.value.getBoundingClientRect()
  const mx = e.clientX - rect.left, my = e.clientY - rect.top
  const w = screenToWorld(mx, my)
  const newNode = designStore.addNode('text', { x: w.x - 80, y: w.y - 30 })
  if (newNode) requestAnimationFrame(draw)
}
</script>

<template>
  <div class="relative w-full h-full overflow-hidden bg-[#0F0F0F]">
    <canvas
      ref="canvasRef"
      class="w-full h-full cursor-grab active:cursor-grabbing"
      @mousedown="onMouseDown"
      @mousemove="onMouseMove"
      @mouseup="onMouseUp"
      @mouseleave="onMouseUp"
      @wheel.prevent="onWheel"
      @dblclick="onDblClick"
    />
    <div class="absolute bottom-2 left-2 text-[10px] text-text-secondary/40 font-mono">
      {{ designStore.nodes.length }} nodes · {{ designStore.connections.length }} connections
      · Double-click to add text
    </div>
  </div>
</template>