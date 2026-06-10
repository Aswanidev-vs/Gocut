<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useDesignStore, getNodeType, NODE_TYPES } from '../../stores/designStore'
import { X, Eye, EyeOff, Lock, Unlock } from 'lucide-vue-next'

const props = defineProps({ playheadTime: { type: Number, default: 0 }, isPlaying: { type: Boolean, default: false } })
const designStore = useDesignStore()

const canvasRef = ref(null)
const isDragging = ref(false)
const dragNode = ref(null)
const dragOffset = { x: 0, y: 0 }
const isPanning = ref(false)
const panStart = { x: 0, y: 0 }

// Connection drag state
const isDraggingConnection = ref(false)
const activeConnectionSource = ref(null) // { nodeId, portType, portName, x, y }
const dragConnectionTarget = ref(null) // hovered socket
const hoveredSocket = ref(null)
const mouseWorldPos = ref({ x: 0, y: 0 })
const dashOffset = ref(0)
const hoveredConnection = ref(null) // connection we are hovering over for insertion

function worldToScreen(wx, wy) {
  return { x: (wx + designStore.panX) * designStore.zoom, y: (wy + designStore.panY) * designStore.zoom }
}
function screenToWorld(sx, sy) {
  return { x: sx / designStore.zoom - designStore.panX, y: sy / designStore.zoom - designStore.panY }
}

function getInputSocketPos(node, idx, total) {
  const nw = 160, nh = 60
  let y = node.y + nh / 2
  if (total > 1) {
    y = node.y + 18 + (idx * (nh - 24) / (total - 1))
  }
  return { x: node.x, y }
}

function getOutputSocketPos(node, idx, total) {
  const nw = 160, nh = 60
  let y = node.y + nh / 2
  if (total > 1) {
    y = node.y + 18 + (idx * (nh - 24) / (total - 1))
  }
  return { x: node.x + nw, y }
}

function getAllSockets() {
  const list = []
  for (const node of designStore.nodes) {
    if (node.visible === false) continue
    const type = getNodeType(node.type)
    if (!type) continue
    
    // Inputs
    for (let i = 0; i < type.inputs.length; i++) {
      const pos = getInputSocketPos(node, i, type.inputs.length)
      list.push({
        nodeId: node.id,
        portType: 'in',
        portName: type.inputs[i],
        x: pos.x,
        y: pos.y
      })
    }
    
    // Outputs
    for (let i = 0; i < type.outputs.length; i++) {
      const pos = getOutputSocketPos(node, i, type.outputs.length)
      list.push({
        nodeId: node.id,
        portType: 'out',
        portName: type.outputs[i],
        x: pos.x,
        y: pos.y
      })
    }
  }
  return list
}

function getSocketAtScreen(mx, my) {
  const sockets = getAllSockets()
  for (const s of sockets) {
    const sPos = worldToScreen(s.x, s.y)
    const dist = Math.hypot(sPos.x - mx, sPos.y - my)
    if (dist <= 20) {
      return s
    }
  }
  return null
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
  for (const conn of designStore.connections) {
    const from = designStore.nodes.find(n => n.id === conn.fromNode)
    const to = designStore.nodes.find(n => n.id === conn.toNode)
    if (!from || !to) continue

    const fromType = getNodeType(from.type)
    const toType = getNodeType(to.type)
    if (!fromType || !toType) continue
    const fromIdx = fromType.outputs.indexOf(conn.fromPort)
    const toIdx = toType.inputs.indexOf(conn.toPort)

    const fPos = getOutputSocketPos(from, fromIdx >= 0 ? fromIdx : 0, fromType.outputs.length)
    const tPos = getInputSocketPos(to, toIdx >= 0 ? toIdx : 0, toType.inputs.length)

    const f = worldToScreen(fPos.x, fPos.y)
    const t = worldToScreen(tPos.x, tPos.y)

    const isInsertionHovered = hoveredConnection.value && hoveredConnection.value.id === conn.id

    c2d.strokeStyle = isInsertionHovered ? '#00D4FF' : '#4a4a4a'
    c2d.lineWidth = isInsertionHovered ? 4 : 2
    
    // Marching ants animation
    c2d.setLineDash([6, 4])
    c2d.lineDashOffset = -dashOffset.value * (isInsertionHovered ? 2 : 1)
    
    c2d.beginPath(); c2d.moveTo(f.x, f.y)
    const cp1x = f.x + (t.x - f.x) * 0.5
    c2d.bezierCurveTo(cp1x, f.y, cp1x, t.y, t.x, t.y)
    c2d.stroke()
    c2d.setLineDash([])
  }

  // Draw temporary connection line preview
  if (isDraggingConnection.value && activeConnectionSource.value) {
    const src = activeConnectionSource.value
    const fromPos = worldToScreen(src.x, src.y)
    const toPos = worldToScreen(mouseWorldPos.value.x, mouseWorldPos.value.y)
    
    c2d.strokeStyle = '#00D4FF'
    c2d.lineWidth = 2
    c2d.setLineDash([4, 4])
    c2d.beginPath()
    c2d.moveTo(fromPos.x, fromPos.y)
    const cp1x = fromPos.x + (toPos.x - fromPos.x) * 0.5
    c2d.bezierCurveTo(cp1x, fromPos.y, cp1x, toPos.y, toPos.x, toPos.y)
    c2d.stroke()
    c2d.setLineDash([])
  }

  // Draw nodes
  for (const node of designStore.nodes) {
    if (node.visible === false) continue
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

    // Input port dots
    for (let i = 0; i < type.inputs.length; i++) {
      const socketWorld = getInputSocketPos(node, i, type.inputs.length)
      const socketScreen = worldToScreen(socketWorld.x, socketWorld.y)
      
      const isHovered = dragConnectionTarget.value && 
                        dragConnectionTarget.value.nodeId === node.id && 
                        dragConnectionTarget.value.portType === 'in' && 
                        dragConnectionTarget.value.portName === type.inputs[i]
      
      const isCompatible = isDraggingConnection.value && 
                           activeConnectionSource.value && 
                           activeConnectionSource.value.nodeId !== node.id && 
                           activeConnectionSource.value.portType === 'out'

      // Glow compatible sockets
      if (isCompatible) {
        c2d.strokeStyle = 'rgba(0, 212, 255, 0.4)'
        c2d.lineWidth = 1.5
        c2d.beginPath()
        c2d.arc(socketScreen.x, socketScreen.y, (8 + Math.sin(dashOffset.value * 0.8) * 2) * designStore.zoom, 0, Math.PI * 2)
        c2d.stroke()
      }

      c2d.fillStyle = isHovered ? '#00D4FF' : '#9CA3AF'
      c2d.strokeStyle = '#1E1E2E'
      c2d.lineWidth = 1.5
      c2d.beginPath()
      c2d.arc(socketScreen.x, socketScreen.y, 5 * designStore.zoom, 0, Math.PI * 2)
      c2d.fill()
      c2d.stroke()

      // Show socket label on hover
      const isCurrentlyHovered = (hoveredSocket.value && 
                                  hoveredSocket.value.nodeId === node.id && 
                                  hoveredSocket.value.portType === 'in' && 
                                  hoveredSocket.value.portName === type.inputs[i]) || isHovered
      if (isCurrentlyHovered) {
        c2d.fillStyle = 'rgba(15, 15, 25, 0.9)'
        c2d.strokeStyle = '#00D4FF'
        c2d.lineWidth = 1
        const text = type.inputs[i].toUpperCase()
        c2d.font = '9px monospace'
        const tw = c2d.measureText(text).width
        c2d.beginPath()
        c2d.roundRect(socketScreen.x - tw - 16, socketScreen.y - 8, tw + 8, 16, 3)
        c2d.fill()
        c2d.stroke()
        c2d.fillStyle = '#E2E8F0'
        c2d.fillText(text, socketScreen.x - tw - 12, socketScreen.y + 3)
      }
    }

    // Output port dots
    for (let i = 0; i < type.outputs.length; i++) {
      const socketWorld = getOutputSocketPos(node, i, type.outputs.length)
      const socketScreen = worldToScreen(socketWorld.x, socketWorld.y)
      
      const isHovered = dragConnectionTarget.value && 
                        dragConnectionTarget.value.nodeId === node.id && 
                        dragConnectionTarget.value.portType === 'out' && 
                        dragConnectionTarget.value.portName === type.outputs[i]

      const isCompatible = isDraggingConnection.value && 
                           activeConnectionSource.value && 
                           activeConnectionSource.value.nodeId !== node.id && 
                           activeConnectionSource.value.portType === 'in'

      // Glow compatible sockets
      if (isCompatible) {
        c2d.strokeStyle = 'rgba(0, 212, 255, 0.4)'
        c2d.lineWidth = 1.5
        c2d.beginPath()
        c2d.arc(socketScreen.x, socketScreen.y, (8 + Math.sin(dashOffset.value * 0.8) * 2) * designStore.zoom, 0, Math.PI * 2)
        c2d.stroke()
      }
      
      c2d.fillStyle = isHovered ? '#FFFFFF' : '#00D4FF'
      c2d.strokeStyle = '#1E1E2E'
      c2d.lineWidth = 1.5
      c2d.beginPath()
      c2d.arc(socketScreen.x, socketScreen.y, 5 * designStore.zoom, 0, Math.PI * 2)
      c2d.fill()
      c2d.stroke()

      // Show socket label on hover
      const isCurrentlyHovered = (hoveredSocket.value && 
                                  hoveredSocket.value.nodeId === node.id && 
                                  hoveredSocket.value.portType === 'out' && 
                                  hoveredSocket.value.portName === type.outputs[i]) || isHovered
      if (isCurrentlyHovered) {
        c2d.fillStyle = 'rgba(15, 15, 25, 0.9)'
        c2d.strokeStyle = '#00D4FF'
        c2d.lineWidth = 1
        const text = type.outputs[i].toUpperCase()
        c2d.font = '9px monospace'
        const tw = c2d.measureText(text).width
        c2d.beginPath()
        c2d.roundRect(socketScreen.x + 8, socketScreen.y - 8, tw + 8, 16, 3)
        c2d.fill()
        c2d.stroke()
        c2d.fillStyle = '#E2E8F0'
        c2d.fillText(text, socketScreen.x + 12, socketScreen.y + 3)
      }
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

let animId = null
function animate() {
  dashOffset.value = (dashOffset.value + 0.3) % 10
  draw()
  animId = requestAnimationFrame(animate)
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

  animate()

  return () => {
    window.removeEventListener('resize', resize)
    obs?.disconnect()
    if (animId) cancelAnimationFrame(animId)
  }
})

watch(() => [designStore.nodes, designStore.connections, designStore.selectedNodeId, designStore.zoom, designStore.panX, designStore.panY], () => draw(), { deep: true })
watch(() => props.playheadTime, () => { if (props.isPlaying) draw() })

function onMouseDown(e) {
  const rect = canvasRef.value.getBoundingClientRect()
  const mx = e.clientX - rect.left, my = e.clientY - rect.top

  // Check socket hit first
  const clickedSocket = getSocketAtScreen(mx, my)
  if (clickedSocket) {
    if (activeConnectionSource.value) {
      // Connect click-to-connect style
      const src = activeConnectionSource.value
      const dst = clickedSocket
      if (src.nodeId !== dst.nodeId && src.portType !== dst.portType) {
        if (src.portType === 'out' && dst.portType === 'in') {
          designStore.addConnection(src.nodeId, src.portName, dst.nodeId, dst.portName)
        } else if (src.portType === 'in' && dst.portType === 'out') {
          designStore.addConnection(dst.nodeId, dst.portName, src.nodeId, src.portName)
        }
      }
      activeConnectionSource.value = null
      isDraggingConnection.value = false
    } else {
      isDraggingConnection.value = true
      activeConnectionSource.value = clickedSocket
      mouseWorldPos.value = screenToWorld(mx, my)
    }
    return
  }

  // Hit test nodes
  for (let i = designStore.nodes.length - 1; i >= 0; i--) {
    const n = designStore.nodes[i]
    if (n.visible === false) continue
    const pos = worldToScreen(n.x, n.y)
    if (mx >= pos.x && mx <= pos.x + 160 && my >= pos.y && my <= pos.y + 60) {
      designStore.selectedNodeId = n.id
      isDragging.value = true
      dragNode.value = n
      
      const w = screenToWorld(mx, my)
      dragOffset.x = w.x - n.x
      dragOffset.y = w.y - n.y

      // Shift key detach & heal logic
      if (e.shiftKey) {
        const incoming = designStore.connections.filter(c => c.toNode === n.id)
        const outgoing = designStore.connections.filter(c => c.fromNode === n.id)
        
        if (incoming.length > 0 && outgoing.length > 0) {
          const firstIn = incoming[0]
          for (const out of outgoing) {
            designStore.addConnection(firstIn.fromNode, firstIn.fromPort, out.toNode, out.toPort)
          }
        }
        
        designStore.connections = designStore.connections.filter(c => c.fromNode !== n.id && c.toNode !== n.id)
      }

      activeConnectionSource.value = null
      return
    }
  }

  // Click on empty space
  activeConnectionSource.value = null
  designStore.selectedNodeId = null
  isPanning.value = true
  panStart.x = e.clientX - designStore.panX * designStore.zoom
  panStart.y = e.clientY - designStore.panY * designStore.zoom
}

function onMouseMove(e) {
  const rect = canvasRef.value.getBoundingClientRect()
  const mx = e.clientX - rect.left
  const my = e.clientY - rect.top

  if (isDraggingConnection.value && activeConnectionSource.value) {
    mouseWorldPos.value = screenToWorld(mx, my)
    
    // Check if hovering over another compatible socket
    const hovered = getSocketAtScreen(mx, my)
    if (hovered && hovered.nodeId !== activeConnectionSource.value.nodeId && hovered.portType !== activeConnectionSource.value.portType) {
      dragConnectionTarget.value = hovered
    } else {
      dragConnectionTarget.value = null
    }
  } else if (isDragging.value && dragNode.value) {
    const w = screenToWorld(mx, my)
    dragNode.value.x = Math.max(0, w.x - dragOffset.x)
    dragNode.value.y = Math.max(0, w.y - dragOffset.y)

    // Check hover insertion logic
    const type = getNodeType(dragNode.value.type)
    if (type && type.inputs.length > 0 && type.outputs.length > 0) {
      const center = { x: dragNode.value.x + 80, y: dragNode.value.y + 30 }
      let closestConn = null
      let foundOverlap = false

      for (const conn of designStore.connections) {
        if (conn.fromNode === dragNode.value.id || conn.toNode === dragNode.value.id) continue
        
        const from = designStore.nodes.find(n => n.id === conn.fromNode)
        const to = designStore.nodes.find(n => n.id === conn.toNode)
        if (!from || !to) continue
        
        const fromType = getNodeType(from.type)
        const toType = getNodeType(to.type)
        if (!fromType || !toType) continue
        const fromIdx = fromType.outputs.indexOf(conn.fromPort)
        const toIdx = toType.inputs.indexOf(conn.toPort)
        
        const f = getOutputSocketPos(from, fromIdx >= 0 ? fromIdx : 0, fromType.outputs.length)
        const t = getInputSocketPos(to, toIdx >= 0 ? toIdx : 0, toType.inputs.length)
        
        // Sample 5 points on the Bezier curve
        const samples = [0.2, 0.35, 0.5, 0.65, 0.8]
        const cp1x = f.x + (t.x - f.x) * 0.5
        const cp1y = f.y
        const cp2x = f.x + (t.x - f.x) * 0.5
        const cp2y = t.y

        for (const val of samples) {
          const mt = 1 - val
          const mt2 = mt * mt
          const mt3 = mt2 * mt
          const val2 = val * val
          const val3 = val2 * val
          
          const px = mt3 * f.x + 3 * mt2 * val * cp1x + 3 * mt * val2 * cp2x + val3 * t.x
          const py = mt3 * f.y + 3 * mt2 * val * cp1y + 3 * mt * val2 * cp2y + val3 * t.y

          // Check if point is inside node bounds (width 160, height 60) with a 15px hover padding
          if (px >= dragNode.value.x - 15 && px <= dragNode.value.x + 175 &&
              py >= dragNode.value.y - 15 && py <= dragNode.value.y + 75) {
            closestConn = conn
            foundOverlap = true
            break
          }
        }
        if (foundOverlap) break
      }
      hoveredConnection.value = closestConn
    } else {
      hoveredConnection.value = null
    }
  } else if (isPanning.value) {
    designStore.panX = (e.clientX - panStart.x) / designStore.zoom
    designStore.panY = (e.clientY - panStart.y) / designStore.zoom
  }

  // Update hovered socket when not dragging anything
  if (!isDraggingConnection.value && !isDragging.value && !isPanning.value) {
    hoveredSocket.value = getSocketAtScreen(mx, my)
  } else {
    hoveredSocket.value = null
  }
}

function onMouseUp() {
  if (isDraggingConnection.value) {
    if (dragConnectionTarget.value && activeConnectionSource.value) {
      const src = activeConnectionSource.value
      const dst = dragConnectionTarget.value
      if (src.portType === 'out' && dst.portType === 'in') {
        designStore.addConnection(src.nodeId, src.portName, dst.nodeId, dst.portName)
      } else if (src.portType === 'in' && dst.portType === 'out') {
        designStore.addConnection(dst.nodeId, dst.portName, src.nodeId, src.portName)
      }
    }
    isDraggingConnection.value = false
    activeConnectionSource.value = null
    dragConnectionTarget.value = null
  }

  // Handle drop hover insertion
  if (isDragging.value && dragNode.value && hoveredConnection.value) {
    const conn = hoveredConnection.value
    const fromNodeId = conn.fromNode
    const fromPort = conn.fromPort
    const toNodeId = conn.toNode
    const toPort = conn.toPort
    
    const type = getNodeType(dragNode.value.type)
    if (type && type.inputs.length > 0 && type.outputs.length > 0) {
      const dragInputPort = type.inputs[0]
      const dragOutputPort = type.outputs[0]
      
      designStore.removeConnection(conn.id)
      designStore.addConnection(fromNodeId, fromPort, dragNode.value.id, dragInputPort)
      designStore.addConnection(dragNode.value.id, dragOutputPort, toNodeId, toPort)
    }
    hoveredConnection.value = null
  }

  isDragging.value = false
  dragNode.value = null
  isPanning.value = false
}

function onWheel(e) {
  e.preventDefault()
  const delta = e.deltaY > 0 ? 0.9 : 1.1
  designStore.zoom = Math.max(0.1, Math.min(5, designStore.zoom * delta))
}

function onDblClick(e) {
  const rect = canvasRef.value.getBoundingClientRect()
  const mx = e.clientX - rect.left, my = e.clientY - rect.top
  const w = screenToWorld(mx, my)
  const newNode = designStore.addNode('text', { x: w.x - 80, y: w.y - 30 })
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
      · Shift+Drag to detach/heal · Drag lines to connect · Hover lines to insert
    </div>
  </div>
</template>