<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { 
  Plus, Trash2, Sliders, Play, Settings, Move, Compass, Database, 
  Image, Type, CircleDot, Spline, Shield, RefreshCw, Info, Search 
} from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const nodes = ref([
  { id: 'media_in', label: 'MediaIn1', type: 'source', x: 60, y: 110, activeViewer: 1, color: 'bg-amber-500/10 border-amber-500/40 text-amber-400' },
  { id: 'color_correct', label: 'ColorCorrect1', type: 'color', x: 220, y: 110, activeViewer: null, color: 'bg-emerald-500/10 border-emerald-500/40 text-emerald-400' },
  { id: 'transform', label: 'Transform1', type: 'transform', x: 380, y: 110, activeViewer: 2, color: 'bg-blue-500/10 border-blue-500/40 text-blue-400' },
  { id: 'merge', label: 'Merge1', type: 'merge', x: 540, y: 110, activeViewer: null, color: 'bg-purple-500/10 border-purple-500/40 text-purple-400' },
  { id: 'media_out', label: 'MediaOut1', type: 'output', x: 700, y: 110, activeViewer: null, color: 'bg-rose-500/10 border-rose-500/40 text-rose-400' }
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

// Shift+Space / Add Tool Search Dialog
const isSearchOpen = ref(false)
const searchQuery = ref('')
const searchInputRef = ref(null)

const allTools = [
  { type: 'transform', label: 'Transform (XF)', desc: 'Translate, rotate, scale' },
  { type: 'color', label: 'Color Corrector (CC)', desc: 'Lift, gamma, gain, saturation' },
  { type: 'text', label: 'Text+ (TXT)', desc: '3D title and styled text overlays' },
  { type: 'mask', label: 'Rectangle Mask (MR)', desc: 'Crop or isolate region' },
  { type: 'ellipse', label: 'Ellipse Mask (ME)', desc: 'Circular compositing mask' },
  { type: 'blur', label: 'Gaussian Blur (BL)', desc: 'Smooth image filters' },
  { type: 'merge', label: 'Merge (MRG)', desc: 'Combine foreground and background' }
]

const filteredTools = computed(() => {
  return allTools.filter(t => t.label.toLowerCase().includes(searchQuery.value.toLowerCase()))
})

function selectNode(id, type) {
  selectedNodeId.value = id
  uiStore.setActiveInspectorTab(type === 'color' ? 'color' : 'edit')
  window.dispatchEvent(new CustomEvent('fusion:select-node', { detail: { id, type } }))
}

function startDrag(e, nodeId) {
  if (e.shiftKey) {
    const prevConn = connections.value.find(c => c.to === nodeId)
    const nextConn = connections.value.find(c => c.from === nodeId)
    connections.value = connections.value.filter(c => c.from !== nodeId && c.to !== nodeId)
    if (prevConn && nextConn) {
      connections.value.push({ from: prevConn.from, to: nextConn.to })
    }
    uiStore.addToast(`Detached node ${nodeId} using Shift-Drag`, 'info', 1200)
  }

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
      node.y = Math.max(10, Math.min(e.clientY - dragOffset.y, 240))
    }
  }
}

function stopDrag() {
  draggingNodeId.value = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

function getBezierPath(fromNodeId, toNodeId) {
  const fromNode = nodes.value.find(n => n.id === fromNodeId)
  const toNode = nodes.value.find(n => n.id === toNodeId)
  if (!fromNode || !toNode) return ''

  const startX = fromNode.x + 120 
  const startY = fromNode.y + 20
  const endX = toNode.x           
  const endY = toNode.y + 20

  const controlX1 = startX + 60
  const controlY1 = startY
  const controlX2 = endX - 60
  const controlY2 = endY

  return `M ${startX} ${startY} C ${controlX1} ${controlY1}, ${controlX2} ${controlY2}, ${endX} ${endY}`
}

function addNode(type) {
  const count = nodes.value.filter(n => n.type === type).length + 1
  const labelMap = {
    transform: `Transform${count}`,
    color: `ColorCorrect${count}`,
    text: `Text${count}`,
    mask: `RectangleMask${count}`,
    ellipse: `EllipseMask${count}`,
    blur: `Blur${count}`,
    merge: `Merge${count}`
  }

  const colorMap = {
    transform: 'bg-blue-500/10 border-blue-500/40 text-blue-400',
    color: 'bg-emerald-500/10 border-emerald-500/40 text-emerald-400',
    text: 'bg-amber-500/10 border-amber-500/40 text-amber-400',
    mask: 'bg-indigo-500/10 border-indigo-500/40 text-indigo-400',
    ellipse: 'bg-indigo-500/10 border-indigo-500/40 text-indigo-400',
    blur: 'bg-pink-500/10 border-pink-500/40 text-pink-400',
    merge: 'bg-purple-500/10 border-purple-500/40 text-purple-400'
  }

  const id = `${type}_${Date.now().toString().slice(-4)}`
  const label = labelMap[type] || `${type}${count}`
  const color = colorMap[type] || 'bg-gray-500/10 border-gray-500/40 text-gray-400'

  const insertX = 350 + (count * 20)
  const insertY = 160

  nodes.value.push({ id, label, type, x: insertX, y: insertY, activeViewer: null, color })

  if (selectedNodeId.value && selectedNodeId.value !== 'media_out') {
    const oldTarget = connections.value.find(c => c.from === selectedNodeId.value)
    if (oldTarget) {
      connections.value = connections.value.filter(c => !(c.from === selectedNodeId.value))
      connections.value.push(
        { from: selectedNodeId.value, to: id },
        { from: id, to: oldTarget.to }
      )
    } else {
      connections.value.push({ from: selectedNodeId.value, to: id })
    }
  }

  selectNode(id, type)
  isSearchOpen.value = false
  searchQuery.value = ''
  uiStore.addToast(`Added Fusion node: ${label}`, 'success', 1200)
}

function deleteSelectedNode() {
  if (!selectedNodeId.value || ['media_in', 'media_out'].includes(selectedNodeId.value)) {
    uiStore.addToast('Cannot delete system source or output node', 'warning', 1500)
    return
  }

  const prevConn = connections.value.find(c => c.to === selectedNodeId.value)
  const nextConn = connections.value.find(c => c.from === selectedNodeId.value)

  nodes.value = nodes.value.filter(n => n.id !== selectedNodeId.value)
  connections.value = connections.value.filter(c => c.from !== selectedNodeId.value && c.to !== selectedNodeId.value)
  if (prevConn && nextConn) {
    connections.value.push({ from: prevConn.from, to: nextConn.to })
  }

  uiStore.addToast('Removed Node', 'info', 1200)
  selectedNodeId.value = 'media_in'
}

function setViewer(node, viewerNum) {
  nodes.value.forEach(n => {
    if (n.activeViewer === viewerNum) n.activeViewer = null
  })
  node.activeViewer = node.activeViewer === viewerNum ? null : viewerNum
  uiStore.addToast(`Routing output of ${node.label} to Viewer ${viewerNum}`, 'info', 1200)
}

const pendingConnectionFrom = ref(null)

function startConnectionFrom(nodeId) {
  pendingConnectionFrom.value = nodeId
  uiStore.addToast(`Selected output from ${nodeId}. Now click another node's input terminal to connect.`, 'info', 2000)
}

function connectToInput(toNodeId) {
  if (!pendingConnectionFrom.value) {
    uiStore.addToast('Select an output terminal first to connect', 'warning', 1800)
    return
  }
  
  if (pendingConnectionFrom.value === toNodeId) {
    uiStore.addToast('Cannot connect a node to itself', 'warning', 1500)
    pendingConnectionFrom.value = null
    return
  }

  // Remove any existing connection going into toNodeId (only 1 input per standard filter node)
  connections.value = connections.value.filter(c => c.to !== toNodeId)

  // Add new connection
  connections.value.push({
    from: pendingConnectionFrom.value,
    to: toNodeId
  })

  uiStore.addToast(`Connected ${pendingConnectionFrom.value} -> ${toNodeId}`, 'success', 1200)
  pendingConnectionFrom.value = null
}

// Global Keyboard Shortcuts for Fusion Nodes (Shift+Space & Delete)
function handleGlobalKeys(e) {
  // Ignore if user typing in input fields
  if (['INPUT', 'TEXTAREA', 'SELECT'].includes(document.activeElement?.tagName)) return

  // Shift + Space: Open Tool Finder
  if (e.shiftKey && e.code === 'Space') {
    e.preventDefault()
    isSearchOpen.value = true
    nextTick(() => {
      if (searchInputRef.value) searchInputRef.value.focus()
    })
  }

  // Delete / Backspace: Remove Node
  if (e.code === 'Delete' || e.code === 'Backspace') {
    e.preventDefault()
    deleteSelectedNode()
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleGlobalKeys)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleGlobalKeys)
})
</script>

<template>
  <div class="h-full bg-black/95 flex flex-col select-none overflow-hidden relative border-t border-border">
    <!-- Fusion Quick Tool Shelf -->
    <div class="h-10 px-3 border-b border-border bg-panel/60 flex items-center justify-between">
      <div class="flex items-center gap-1">
        <button @click="addNode('text')" class="p-1.5 rounded hover:bg-border/60 text-text-secondary hover:text-amber-400 transition-colors" title="Text+ (Txt)"><Type :size="14" /></button>
        <button @click="addNode('color')" class="p-1.5 rounded hover:bg-border/60 text-text-secondary hover:text-emerald-400 transition-colors" title="Color Corrector (CC)"><Sliders :size="14" /></button>
        <button @click="addNode('transform')" class="p-1.5 rounded hover:bg-border/60 text-text-secondary hover:text-blue-400 transition-colors" title="Transform (Xf)"><Compass :size="14" /></button>
        <button @click="addNode('merge')" class="p-1.5 rounded hover:bg-border/60 text-text-secondary hover:text-purple-400 transition-colors" title="Merge (Mrg)"><Plus :size="14" /></button>
        <div class="h-4 w-px bg-border mx-1" />
        <button @click="addNode('mask')" class="p-1.5 rounded hover:bg-border/60 text-text-secondary hover:text-indigo-400 transition-colors" title="Rectangle Mask (Msk)"><CircleDot :size="14" /></button>
        <button @click="addNode('blur')" class="p-1.5 rounded hover:bg-border/60 text-text-secondary hover:text-pink-400 transition-colors" title="Blur (Blr)"><Spline :size="14" /></button>
      </div>

      <div class="flex items-center gap-2">
        <button
          @click="isSearchOpen = true; nextTick(() => searchInputRef?.focus())"
          class="px-2.5 py-1 rounded bg-border/40 hover:bg-border text-text-primary text-[10px] flex items-center gap-1 font-bold shadow-sm"
        >
          <Search :size="10" /> Shift+Space Tool Finder
        </button>
        <button
          @click="deleteSelectedNode"
          class="p-1.5 rounded hover:bg-red-950 text-text-secondary hover:text-red-400 transition-colors"
          title="Delete Node (Delete)"
        >
          <Trash2 :size="12" />
        </button>
      </div>
    </div>

    <!-- Canvas Node Workspace -->
    <div class="flex-1 relative overflow-auto bg-[radial-gradient(#202028_1.2px,transparent_1.2px)] [background-size:20px_20px]">
      <!-- SVG Connectors Layer -->
      <svg class="absolute inset-0 w-full h-full pointer-events-none z-0">
        <g v-for="(conn, idx) in connections" :key="idx">
          <!-- Outer glowing flow spline -->
          <path
            :d="getBezierPath(conn.from, conn.to)"
            fill="none"
            stroke="rgba(0, 212, 255, 0.12)"
            stroke-width="6"
            class="glow-path"
          />
          <!-- Core flowing spline (marching ants effect) -->
          <path
            :d="getBezierPath(conn.from, conn.to)"
            fill="none"
            :stroke="selectedNodeId === conn.from || selectedNodeId === conn.to ? '#00D4FF' : '#52525b'"
            stroke-width="2.5"
            stroke-dasharray="6, 6"
            class="flow-line transition-all duration-300"
          />
        </g>
      </svg>

      <!-- Nodes Layer -->
      <div
        v-for="node in nodes"
        :key="node.id"
        :style="{ left: node.x + 'px', top: node.y + 'px' }"
        class="absolute w-[120px] rounded-lg border px-3 py-2 flex flex-col justify-between cursor-grab active:cursor-grabbing z-10 transition-all duration-200"
        :class="[
          node.color,
          selectedNodeId === node.id ? 'ring-2 ring-accent ring-offset-2 ring-offset-black shadow-lg shadow-accent/20 border-transparent bg-accent/20 scale-[1.03]' : 'shadow-md shadow-black/80 hover:scale-[1.01]'
        ]"
        @mousedown="startDrag($event, node.id)"
        @click.stop="selectNode(node.id, node.type)"
      >
        <div class="flex items-center justify-between">
          <span class="text-[10px] font-bold truncate tracking-wide text-text-primary">{{ node.label }}</span>
          <!-- Viewer routing dots (Resolve style 1 & 2 circles) -->
          <div class="flex items-center gap-0.5">
            <button
              @click.stop="setViewer(node, 1)"
              class="w-3.5 h-3.5 rounded-full border flex items-center justify-center text-[7px] font-mono leading-none transition-all"
              :class="node.activeViewer === 1 ? 'bg-accent border-accent text-bg font-bold shadow-sm shadow-accent/50' : 'border-zinc-700 text-zinc-500 hover:text-text-primary'"
            >1</button>
            <button
              @click.stop="setViewer(node, 2)"
              class="w-3.5 h-3.5 rounded-full border flex items-center justify-center text-[7px] font-mono leading-none transition-all"
              :class="node.activeViewer === 2 ? 'bg-amber-500 border-amber-500 text-bg font-bold shadow-sm shadow-amber-500/50' : 'border-zinc-700 text-zinc-500 hover:text-text-primary'"
            >2</button>
          </div>
        </div>
        
        <!-- Connector Terminals (Click-to-Connect) -->
        <div class="flex justify-between items-center mt-2.5">
          <!-- Input terminal -->
          <button 
            v-if="node.type !== 'source'" 
            @click.stop="connectToInput(node.id)" 
            class="w-3 h-3 rounded bg-zinc-800 border border-zinc-500 hover:border-accent hover:bg-accent/40 -ml-[18px] z-20 flex items-center justify-center transition-all"
            :class="pendingConnectionFrom ? 'ring-2 ring-accent animate-pulse' : ''"
            title="Click to connect source here"
          >
            <div class="w-1.5 h-1.5 rounded-full bg-zinc-400" />
          </button>
          <div v-else />
          
          <!-- Output terminal -->
          <button 
            v-if="node.type !== 'output'" 
            @click.stop="startConnectionFrom(node.id)" 
            class="w-3 h-3 rounded bg-accent hover:bg-accent-hover border border-white/40 -mr-[18px] z-20 flex items-center justify-center transition-all"
            :class="pendingConnectionFrom === node.id ? 'ring-2 ring-white scale-110' : ''"
            title="Click output to route"
          >
            <div class="w-1.5 h-1.5 rounded-full bg-white" />
          </button>
          <div v-else />
        </div>
      </div>
    </div>

    <!-- Search Tool Dialog (Shift + Space style) -->
    <div v-if="isSearchOpen" class="absolute inset-0 bg-black/70 flex items-center justify-center z-50 p-4 backdrop-blur-sm">
      <div class="bg-panel border border-border w-full max-w-sm rounded-xl overflow-hidden shadow-2xl animate-in fade-in zoom-in duration-150">
        <div class="p-3 border-b border-border flex items-center gap-2">
          <Search :size="14" class="text-text-secondary" />
          <input
            ref="searchInputRef"
            type="text"
            v-model="searchQuery"
            placeholder="Select Tool / Effect Node…"
            class="flex-1 bg-transparent text-text-primary text-xs outline-none"
            @keydown.esc="isSearchOpen = false"
          />
        </div>
        <div class="max-h-60 overflow-y-auto p-1.5 space-y-0.5">
          <button
            v-for="tool in filteredTools"
            :key="tool.type"
            @click="addNode(tool.type)"
            class="w-full text-left p-2 rounded-lg hover:bg-accent/15 hover:text-accent transition-all flex flex-col"
          >
            <span class="text-[11px] font-bold text-text-primary">{{ tool.label }}</span>
            <span class="text-[9px] text-text-secondary mt-0.5">{{ tool.desc }}</span>
          </button>
        </div>
        <div class="p-2 border-t border-border bg-panel/30 flex justify-end">
          <button @click="isSearchOpen = false" class="px-2.5 py-1 rounded bg-border text-text-primary text-[10px]">Cancel</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@keyframes flow {
  from {
    stroke-dashoffset: 24;
  }
  to {
    stroke-dashoffset: 0;
  }
}

.flow-line {
  animation: flow 1.2s linear infinite;
}
</style>
