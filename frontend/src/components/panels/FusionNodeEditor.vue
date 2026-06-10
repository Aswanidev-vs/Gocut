<script setup>
import { ref, computed, onMounted } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { Plus, Trash2, Sliders, Play, Settings, Move, Compass, Database } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const selectedClip = computed(() => {
  if (timelineStore.selectedClips.length > 0) {
    return timelineStore.selectedClips[0]
  }
  return null
})

// Node template definitions
const nodes = ref([
  { id: 'media_in', label: 'MediaIn1', type: 'source', x: 80, y: 80, color: 'bg-amber-500/10 border-amber-500/50 text-amber-400' },
  { id: 'color_correct', label: 'ColorCorrect1', type: 'color', x: 240, y: 80, color: 'bg-emerald-500/10 border-emerald-500/50 text-emerald-400' },
  { id: 'transform', label: 'Transform1', type: 'transform', x: 400, y: 80, color: 'bg-blue-500/10 border-blue-500/50 text-blue-400' },
  { id: 'merge', label: 'Merge1', type: 'merge', x: 560, y: 80, color: 'bg-purple-500/10 border-purple-500/50 text-purple-400' },
  { id: 'media_out', label: 'MediaOut1', type: 'output', x: 720, y: 80, color: 'bg-rose-500/10 border-rose-500/50 text-rose-400' }
])

const connections = ref([
  { from: 'media_in', to: 'color_correct' },
  { from: 'color_correct', to: 'transform' },
  { from: 'transform', to: 'merge' },
  { from: 'merge', to: 'media_out' }
])

const selectedNodeId = ref('transform')
const draggingNodeId = ref(null)
const dragOffset = { x: 0, y: 0 }

function selectNode(id, type) {
  selectedNodeId.value = id
  
  // Map selected node to uiStore / inspector tabs for immediate feedback
  if (type === 'color') {
    uiStore.setActiveInspectorTab('color')
  } else if (type === 'transform') {
    uiStore.setActiveInspectorTab('edit')
  } else if (type === 'source') {
    uiStore.setActiveInspectorTab('edit')
  }
  
  uiStore.addToast(`Selected node: ${id}`, 'info', 1000)
}

// Node dragging logic
function startDrag(e, nodeId) {
  draggingNodeId.value = nodeId
  const node = nodes.value.find(n => n.id === nodeId)
  if (node) {
    dragOffset.x = e.clientX - node.x
    dragOffset.y = e.clientY - node.y
  }
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  e.preventDefault()
}

function onDrag(e) {
  if (draggingNodeId.value !== null) {
    const node = nodes.value.find(n => n.id === draggingNodeId.value)
    if (node) {
      node.x = e.clientX - dragOffset.x
      node.y = Math.max(20, Math.min(e.clientY - dragOffset.y, 220)) // constrain inside viewport
    }
  }
}

function stopDrag() {
  draggingNodeId.value = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

// Calculate SVG Bezier Spline paths
function getBezierPath(fromNodeId, toNodeId) {
  const fromNode = nodes.value.find(n => n.id === fromNodeId)
  const toNode = nodes.value.find(n => n.id === toNodeId)
  if (!fromNode || !toNode) return ''

  const startX = fromNode.x + 100 // Right side of node box
  const startY = fromNode.y + 24
  const endX = toNode.x           // Left side of node box
  const endY = toNode.y + 24

  const controlX1 = startX + 60
  const controlY1 = startY
  const controlX2 = endX - 60
  const controlY2 = endY

  return `M ${startX} ${startY} C ${controlX1} ${controlY1}, ${controlX2} ${controlY2}, ${endX} ${endY}`
}

function addFusionNode(type) {
  const lastNode = nodes.value[nodes.value.length - 2] // node before media_out
  const outputNode = nodes.value[nodes.value.length - 1]
  
  const id = `${type}_${Date.now().toString().slice(-4)}`
  const label = type.charAt(0).toUpperCase() + type.slice(1) + 'Node'
  
  let color = 'bg-sky-500/10 border-sky-500/50 text-sky-400'
  if (type === 'mask') color = 'bg-indigo-500/10 border-indigo-500/50 text-indigo-400'
  if (type === 'effect') color = 'bg-pink-500/10 border-pink-500/50 text-pink-400'

  // Position between last node and output
  const newX = (lastNode.x + outputNode.x) / 2
  const newY = lastNode.y + 30

  // Insert node
  nodes.value.splice(nodes.value.length - 1, 0, {
    id, label, type, x: newX, y: newY, color
  })

  // Re-route connections
  connections.value = connections.value.filter(c => c.to !== 'media_out')
  connections.value.push(
    { from: lastNode.id, to: id },
    { from: id, to: 'media_out' }
  )

  uiStore.addToast(`Inserted node: ${label}`, 'success', 1200)
}
</script>

<template>
  <div class="h-full bg-black/95 flex flex-col select-none overflow-hidden relative border-t border-border">
    <!-- Top toolbar of Node Graph -->
    <div class="h-9 px-3 border-b border-border bg-panel/60 flex items-center justify-between text-xs">
      <div class="flex items-center gap-4">
        <span class="text-[10px] font-bold uppercase tracking-wider text-text-secondary">Fusion Node Graph</span>
        <div class="flex items-center gap-1.5 border-l border-border pl-4">
          <button
            @click="addFusionNode('mask')"
            class="px-2 py-0.5 rounded bg-border/40 hover:bg-border text-text-primary text-[10px] flex items-center gap-1"
          >
            <Plus :size="10" /> Add Mask
          </button>
          <button
            @click="addFusionNode('effect')"
            class="px-2 py-0.5 rounded bg-border/40 hover:bg-border text-text-primary text-[10px] flex items-center gap-1"
          >
            <Plus :size="10" /> Add Effect
          </button>
        </div>
      </div>
      
      <div class="text-[10px] text-text-secondary flex items-center gap-1">
        <Move :size="10" /> Drag nodes to structure your composition pipeline
      </div>
    </div>

    <!-- Canvas Node Workspace -->
    <div class="flex-1 relative overflow-auto bg-[radial-gradient(#202020_1px,transparent_1px)] [background-size:16px_16px]">
      <!-- SVG Connectors Layer -->
      <svg class="absolute inset-0 w-full h-full pointer-events-none z-0">
        <g v-for="(conn, idx) in connections" :key="idx">
          <!-- Shadow path for glowing line effect -->
          <path
            :d="getBezierPath(conn.from, conn.to)"
            fill="none"
            stroke="rgba(0, 212, 255, 0.15)"
            stroke-width="5"
          />
          <path
            :d="getBezierPath(conn.from, conn.to)"
            fill="none"
            :stroke="selectedNodeId === conn.from || selectedNodeId === conn.to ? '#00D4FF' : '#4b5563'"
            stroke-width="2"
            class="transition-all duration-300"
          />
        </g>
      </svg>

      <!-- Nodes Layer -->
      <div
        v-for="node in nodes"
        :key="node.id"
        :style="{ left: node.x + 'px', top: node.y + 'px' }"
        class="absolute w-[120px] rounded border px-3 py-2 flex flex-col justify-between cursor-grab active:cursor-grabbing z-10 transition-shadow duration-200"
        :class="[
          node.color,
          selectedNodeId === node.id ? 'ring-2 ring-accent shadow-lg shadow-accent/20' : 'shadow shadow-black/50'
        ]"
        @mousedown="startDrag($event, node.id)"
        @click.stop="selectNode(node.id, node.type)"
      >
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-bold truncate">{{ node.label }}</span>
          <component
            :is="node.type === 'source' ? Database : node.type === 'color' ? Sliders : node.type === 'transform' ? Compass : Settings"
            :size="10"
            class="opacity-60"
          />
        </div>
        
        <!-- Connector Terminals -->
        <div class="flex justify-between items-center mt-2.5">
          <!-- Input terminal -->
          <div v-if="node.type !== 'source'" class="w-1.5 h-1.5 rounded-full bg-border border border-white/40 -ml-4" />
          <div v-else />
          
          <!-- Output terminal -->
          <div v-if="node.type !== 'output'" class="w-1.5 h-1.5 rounded-full bg-accent border border-white/40 -mr-4" />
          <div v-else />
        </div>
      </div>
    </div>
  </div>
</template>
