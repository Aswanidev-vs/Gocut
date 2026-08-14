import { defineStore, getActivePinia } from 'pinia'
import { ref, computed } from 'vue'

const generateId = () => crypto.randomUUID()

function getDesignHistoryStoreSafely() {
  try {
    const pinia = getActivePinia()
    if (!pinia) return null
    return pinia._s && pinia._s.get('designHistory') ? pinia._s.get('designHistory') : null
  } catch (_) {
    return null
  }
}

function pushDesignSnapshot() {
  try {
    const hs = getDesignHistoryStoreSafely()
    if (hs) hs.pushSnapshot()
  } catch (_) { /* ignore */ }
}

// Socket port color mapping in DaVinci Resolve Fusion style
export const FUSION_SOCKET_COLORS = {
  bg: '#EAB308',     // Yellow (Background input)
  fg: '#22C55E',     // Green (Foreground input)
  mask: '#3B82F6',   // Blue (Effect Mask input)
  matte: '#EC4899',  // Magenta / Garbage Matte
  in: '#94A3B8',    // Standard input
  out: '#FFFFFF',   // Standard output (White)
}

// ============== NODE TYPE DEFINITIONS ==============
export const NODE_TYPES = {
  // Fusion Standard Media Nodes
  mediaIn: { cat: 'Sources', label: 'MediaIn', col: '#0284C7', in: ['mask'], out: ['out'], params: [
    { id: 'assetId',   label: 'Source Clip', type: 'asset' },
    { id: 'startTime', label: 'Global In',   type: 'number', def: 0, min: 0, step: 0.1, suffix: 's' },
    { id: 'duration',  label: 'Hold Frames', type: 'number', def: 5, min: 0.1, step: 0.1, suffix: 's' },
  ]},
  mediaOut: { cat: 'Output', label: 'MediaOut', col: '#F43F5E', in: ['in'], out: [], params: []},

  // Fusion Source Generators
  background: { cat: 'Sources', label: 'Background', col: '#F59E0B', in: ['mask'], out: ['out'], params: [
    { id: 'type', label: 'Type', type: 'select', def: 'solid', options: [
      { value: 'solid', label: 'Solid Color' },
      { value: 'horizontal', label: 'Horizontal Gradient' },
      { value: 'vertical', label: 'Vertical Gradient' },
      { value: 'radial', label: 'Radial Gradient' },
    ]},
    { id: 'color', label: 'Primary Color', type: 'color' },
    { id: 'color2', label: 'Secondary Color', type: 'color' },
    { id: 'alpha', label: 'Alpha', type: 'number', def: 1, min: 0, max: 1, step: 0.01 },
    { id: 'width', label: 'Width', type: 'number', def: 1920, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 1080, min: 1, step: 1, suffix: 'px' },
  ], defaults: { color: '#000000', color2: '#1e293b' }},

  fastNoise: { cat: 'Sources', label: 'FastNoise', col: '#F59E0B', in: ['mask'], out: ['out'], params: [
    { id: 'detail', label: 'Detail', type: 'number', def: 4, min: 1, max: 10, step: 1 },
    { id: 'contrast', label: 'Contrast', type: 'number', def: 1.5, min: 0, max: 5, step: 0.1 },
    { id: 'brightness', label: 'Brightness', type: 'number', def: 0, min: -1, max: 1, step: 0.05 },
    { id: 'scale', label: 'Scale', type: 'number', def: 15, min: 1, max: 100, step: 1 },
    { id: 'seethe', label: 'Seethe Rate', type: 'number', def: 0.5, min: 0, max: 5, step: 0.1 },
    { id: 'color1', label: 'Color 1', type: 'color' },
    { id: 'color2', label: 'Color 2', type: 'color' },
  ], defaults: { color1: '#000000', color2: '#FFFFFF' }},

  textPlus: { cat: 'Sources', label: 'Text+', col: '#F59E0B', in: ['mask'], out: ['out'], params: [
    { id: 'text', label: 'Styled Text', type: 'text' },
    { id: 'fontSize', label: 'Size', type: 'number', def: 64, min: 8, max: 400, step: 1, suffix: 'px' },
    { id: 'color', label: 'Color', type: 'color' },
    { id: 'fontFamily', label: 'Font', type: 'font' },
    { id: 'tracking', label: 'Tracking', type: 'number', def: 0, min: -10, max: 50, step: 1 },
    { id: 'writeOnStart', label: 'Write On Start', type: 'number', def: 0, min: 0, max: 1, step: 0.01 },
    { id: 'writeOnEnd', label: 'Write On End', type: 'number', def: 1, min: 0, max: 1, step: 0.01 },
    { id: 'bold', label: 'Bold', type: 'toggle' },
    { id: 'italic', label: 'Italic', type: 'toggle' },
  ], defaults: { text: 'Fusion Text+', color: '#FFFFFF' }},

  // Legacy aliases
  media: { cat: 'Sources', label: 'Media In', col: '#00D4FF', in: ['mask'], out: ['out'], params: [
    { id: 'assetId',   label: 'Source',   type: 'asset' },
    { id: 'startTime', label: 'In',       type: 'number', def: 0, min: 0, step: 0.1, suffix: 's' },
    { id: 'duration',  label: 'Duration', type: 'number', def: 5, min: 0.1, step: 0.1, suffix: 's' },
  ]},
  text: { cat: 'Sources', label: 'Text', col: '#F59E0B', in: ['mask'], out: ['out'], params: [
    { id: 'text',       label: 'Text',  type: 'text' },
    { id: 'fontSize',   label: 'Size',  type: 'number', def: 48, min: 8, max: 300, step: 1, suffix: 'px' },
    { id: 'color',      label: 'Color', type: 'color' },
    { id: 'fontFamily', label: 'Font',  type: 'font' },
    { id: 'bold',       label: 'Bold',  type: 'toggle' },
    { id: 'italic',     label: 'Italic',type: 'toggle' },
  ], defaults: { text: 'Hello Gocut', color: '#FFFFFF' }},

  rectangle: { cat: 'Sources', label: 'Rectangle', col: '#F59E0B', in: ['mask'], out: ['out'], params: [
    { id: 'width', label: 'Width', type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'fill', label: 'Fill', type: 'color' },
    { id: 'cornerRadius', label: 'Radius', type: 'number', def: 0, min: 0, max: 200, step: 1, suffix: 'px' },
  ], defaults: { fill: '#00D4FF' }},

  ellipse: { cat: 'Sources', label: 'Ellipse', col: '#F59E0B', in: ['mask'], out: ['out'], params: [
    { id: 'width', label: 'Width', type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'fill', label: 'Fill', type: 'color' },
  ], defaults: { fill: '#EC4899' }},

  // Fusion Compositing & Mask Nodes
  merge: { cat: 'Composite', label: 'Merge', col: '#10B981', in: ['bg', 'fg', 'mask'], out: ['out'], params: [
    { id: 'mode', label: 'Apply Mode', type: 'select', def: 'normal', options: [
      { value: 'normal', label: 'Normal / Over' },
      { value: 'multiply', label: 'Multiply' },
      { value: 'screen', label: 'Screen' },
      { value: 'overlay', label: 'Overlay' },
      { value: 'add', label: 'Add / Linear Dodge' },
      { value: 'subtract', label: 'Subtract' },
      { value: 'difference', label: 'Difference' },
      { value: 'lighten', label: 'Lighten' },
      { value: 'darken', label: 'Darken' },
      { value: 'colordodge', label: 'Color Dodge' },
      { value: 'colorburn', label: 'Color Burn' },
    ]},
    { id: 'blend', label: 'Blend', type: 'number', def: 1, min: 0, max: 1, step: 0.01 },
    { id: 'burnIn', label: 'Burn In', type: 'number', def: 0, min: 0, max: 1, step: 0.01 },
    { id: 'centerPivot', label: 'Center Pivot', type: 'toggle', def: true },
  ]},

  maskPolygon: { cat: 'Mask', label: 'Polygon Mask', col: '#3B82F6', in: [], out: ['out'], params: [
    { id: 'invert', label: 'Invert Mask', type: 'toggle', def: false },
    { id: 'softEdge', label: 'Soft Edge / Feather', type: 'number', def: 0, min: 0, max: 100, step: 0.5, suffix: 'px' },
    { id: 'borderWidth', label: 'Border Width', type: 'number', def: 0, min: 0, max: 50, step: 0.5, suffix: 'px' },
    { id: 'paintMode', label: 'Paint Mode', type: 'select', def: 'merge', options: [
      { value: 'merge', label: 'Merge / Add' },
      { value: 'subtract', label: 'Subtract' },
      { value: 'intersect', label: 'Intersect' },
    ]},
  ]},

  // Fusion Transform & Spatial Nodes
  transform: { cat: 'Transform', label: 'Transform (Xf)', col: '#8B5CF6', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'x', label: 'Center X', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'y', label: 'Center Y', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'scaleX', label: 'Size X', type: 'number', def: 1, min: 0, max: 10, step: 0.01 },
    { id: 'scaleY', label: 'Size Y', type: 'number', def: 1, min: 0, max: 10, step: 0.01 },
    { id: 'rotation', label: 'Angle', type: 'number', def: 0, min: -720, max: 720, step: 1, suffix: '°' },
    { id: 'opacity', label: 'Blend', type: 'number', def: 1, min: 0, max: 1, step: 0.01 },
    { id: 'edges', label: 'Edges', type: 'select', def: 'black', options: [
      { value: 'black', label: 'Black / Transparent' },
      { value: 'wrap', label: 'Wrap / Repeat' },
      { value: 'mirror', label: 'Mirror' },
    ]},
  ]},

  // Fusion Color Nodes
  colorCorrector: { cat: 'Color', label: 'ColorCorrector (CC)', col: '#EC4899', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'lift', label: 'Lift / Shadows', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'gamma', label: 'Gamma / Midtones', type: 'number', def: 1, min: 0.1, max: 4, step: 0.05 },
    { id: 'gain', label: 'Gain / Highlights', type: 'number', def: 1, min: 0, max: 4, step: 0.05 },
    { id: 'saturation', label: 'Saturation', type: 'number', def: 1, min: 0, max: 4, step: 0.02 },
    { id: 'tint', label: 'Tint Hue', type: 'number', def: 0, min: -180, max: 180, step: 1, suffix: '°' },
    { id: 'brightness', label: 'Brightness Offset', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
  ]},

  colorCorrect: { cat: 'Color', label: 'Color Correct', col: '#EC4899', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'brightness', label: 'Brightness', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'contrast',   label: 'Contrast',   type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'saturation', label: 'Saturation', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'hue',        label: 'Hue',        type: 'number', def: 0, min: -180, max: 180, step: 1, suffix: '°' },
  ]},

  blur: { cat: 'Effects', label: 'Blur', col: '#EC4899', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'radius', label: 'Blur Size', type: 'number', def: 5, min: 0, max: 100, step: 0.5, suffix: 'px' },
    { id: 'blend', label: 'Blend', type: 'number', def: 1, min: 0, max: 1, step: 0.01 },
  ]},

  glow: { cat: 'Effects', label: 'Glow', col: '#EC4899', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'intensity', label: 'Glow / Shine', type: 'number', def: 1, min: 0, max: 5, step: 0.1 },
    { id: 'color', label: 'Glow Color', type: 'color' },
    { id: 'radius', label: 'Glow Size', type: 'number', def: 10, min: 0, max: 100, step: 1, suffix: 'px' },
  ], defaults: { color: '#00D4FF' }},

  shadow: { cat: 'Effects', label: 'Drop Shadow', col: '#EC4899', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'color', label: 'Shadow Color', type: 'color' },
    { id: 'blur', label: 'Softness', type: 'number', def: 8, min: 0, max: 100, step: 1, suffix: 'px' },
    { id: 'offsetX', label: 'Offset X', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'offsetY', label: 'Offset Y', type: 'number', def: 4, step: 1, suffix: 'px' },
  ], defaults: { color: '#000000' }},

  chromaKey: { cat: 'Effects', label: 'Chroma Keyer', col: '#EC4899', in: ['in', 'mask'], out: ['out'], params: [
    { id: 'keyColor', label: 'Key Color', type: 'color' },
    { id: 'similarity', label: 'Acceptance Range', type: 'number', def: 0.4, min: 0, max: 1, step: 0.01 },
    { id: 'smoothness', label: 'Edge Softness', type: 'number', def: 0.1, min: 0, max: 1, step: 0.01 },
  ], defaults: { keyColor: '#00FF00' }},

  output: { cat: 'Output', label: 'Output', col: '#F43F5E', in: ['in'], out: [], params: []},

  // ============ EXTENDED UTILITY NODES ============
  crop: { cat: 'Transform', label: 'Crop', col: '#8B5CF6', in: ['in'], out: ['out'], params: [
    { id: 'x', label: 'X', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'y', label: 'Y', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'width', label: 'Width', type: 'number', def: 1920, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 1080, min: 1, step: 1, suffix: 'px' },
  ]},

  mask: { cat: 'Mask', label: 'Mask', col: '#3B82F6', in: ['in'], out: ['out'], params: [
    { id: 'maskType', label: 'Shape', type: 'select', def: 'rectangle', options: [
      { value: 'rectangle', label: 'Rectangle' }, { value: 'ellipse', label: 'Ellipse' },
    ]},
    { id: 'x', label: 'X', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'y', label: 'Y', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'width', label: 'Width', type: 'number', def: 1920, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 1080, min: 1, step: 1, suffix: 'px' },
    { id: 'feather', label: 'Feather', type: 'number', def: 0, min: 0, max: 200, step: 1, suffix: 'px' },
  ]},

  timeShift: { cat: 'Time', label: 'Time Shift', col: '#F59E0B', in: ['in'], out: ['out'], params: [
    { id: 'offset', label: 'Offset', type: 'number', def: 0, step: 0.1, suffix: 's' },
  ]},

  expression: { cat: 'Utility', label: 'Expression', col: '#A78BFA', in: ['a', 'b', 'c', 'd'], out: ['out'], params: [
    { id: 'expression', label: 'Expression', type: 'text', def: 'a + b' },
  ]},
}

export function getNodeType(type) {
  const def = NODE_TYPES[type]
  if (!def) return null
  return { type, category: def.cat, label: def.label, color: def.col, inputs: def.in, outputs: def.out, params: def.params }
}

export const EASING_TYPES = [
  { id: 'linear',   label: 'Linear' },
  { id: 'easeIn',   label: 'Ease In' },
  { id: 'easeOut',  label: 'Ease Out' },
  { id: 'easeInOut',label: 'Ease In/Out' },
  { id: 'smooth',   label: 'Smooth (S)' },
  { id: 'bounce',   label: 'Bounce' },
  { id: 'elastic',  label: 'Elastic' },
]

function makeDefaultParams(type) {
  const def = NODE_TYPES[type]
  if (!def) return {}
  const params = {}
  for (const p of def.params) {
    params[p.id] = p.def !== undefined ? p.def : (p.type === 'toggle' ? false : '')
  }
  if (def.defaults) Object.assign(params, def.defaults)
  return params
}

export const useDesignStore = defineStore('design', () => {
  const composition = ref({ name: 'Main Comp', width: 1920, height: 1080, fps: 30, duration: 5, background: '#0a0a0a' })
  const nodes = ref([])
  const connections = ref([])
  const selectedNodeId = ref(null)
  const selectedNodeIds = ref(new Set())
  const selectedConnectionId = ref(null)

  // DaVinci Resolve Fusion Dual Viewers (Viewer 1 / Left & Viewer 2 / Right)
  const viewer1NodeId = ref(null)
  const viewer2NodeId = ref(null)
  const activeViewer = ref(1) // 1 or 2
  const zoom = ref(1)
  const panX = ref(0)
  const panY = ref(0)
  const snapEnabled = ref(true)
  const presets = ref([])
  const groups = ref([])
  const bookmarks = ref({})

  const selectedNode = computed(() => nodes.value.find(n => n.id === selectedNodeId.value) || null)
  const outputNode = computed(() => nodes.value.find(n => n.type === 'output') || null)
  const selectedNodes = computed(() => nodes.value.filter(n => selectedNodeIds.value.has(n.id)))

  function addNode(type, opts = {}) {
    if (!NODE_TYPES[type]) return null
    const node = {
      id: generateId(), type,
      x: opts.x ?? 150 + Math.random() * 100,
      y: opts.y ?? 150 + Math.random() * 100,
      params: opts.params ?? makeDefaultParams(type),
      keyframes: opts.keyframes ?? {},
      visible: true, locked: false,
      label: opts.label ?? NODE_TYPES[type].label,
    }
    
    // Auto-connect with selected node
    const prevId = selectedNodeId.value
    nodes.value.push(node)
    
    if (prevId) {
      const prev = nodes.value.find(n => n.id === prevId)
      if (prev) {
        const prevType = getNodeType(prev.type)
        const newType = getNodeType(type)
        if (prevType && newType && prevType.outputs.length > 0 && newType.inputs.length > 0) {
          // Add connection from first output of previous to first input of new
          addConnection(prev.id, prevType.outputs[0], node.id, newType.inputs[0])
        }
      }
    }
    
    selectedNodeId.value = node.id
    pushDesignSnapshot()
    return node
  }

  function removeNode(nodeId) {
    connections.value = connections.value.filter(c => c.fromNode !== nodeId && c.toNode !== nodeId)
    nodes.value = nodes.value.filter(n => n.id !== nodeId)
    if (selectedNodeId.value === nodeId) selectedNodeId.value = null
    pushDesignSnapshot()
  }

  function removeSelectedNode() {
    if (selectedNodeId.value) removeNode(selectedNodeId.value)
  }

  function duplicateSelectedNode() {
    if (!selectedNodeId.value) return
    const node = nodes.value.find(n => n.id === selectedNodeId.value)
    if (!node) return
    const cloned = JSON.parse(JSON.stringify(node))
    cloned.id = generateId()
    cloned.x += 40
    cloned.y += 40
    cloned.label = node.label + ' Copy'
    nodes.value.push(cloned)
    selectedNodeId.value = cloned.id
    pushDesignSnapshot()
  }

  function addConnection(fromNode, fromPort, toNode, toPort) {
    connections.value = connections.value.filter(c => !(c.toNode === toNode && c.toPort === toPort))
    connections.value.push({ id: generateId(), fromNode, fromPort, toNode, toPort })
    pushDesignSnapshot()
  }

  function removeConnection(id) {
    connections.value = connections.value.filter(c => c.id !== id)
    pushDesignSnapshot()
  }

  function addKeyframe(nodeId, paramId, time, value, easing = 'linear') {
    const node = nodes.value.find(n => n.id === nodeId)
    if (!node) return
    if (!node.keyframes[paramId]) node.keyframes[paramId] = []
    node.keyframes[paramId] = node.keyframes[paramId].filter(k => Math.abs(k.time - time) > 0.01)
    node.keyframes[paramId].push({ id: generateId(), time, value, easing })
    node.keyframes[paramId].sort((a, b) => a.time - b.time)
    pushDesignSnapshot()
  }

  function removeKeyframe(nodeId, paramId, keyframeId) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (!node || !node.keyframes[paramId]) return
    node.keyframes[paramId] = node.keyframes[paramId].filter(k => k.id !== keyframeId)
    pushDesignSnapshot()
  }

  function getParamValue(nodeId, paramId, time) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (!node || !node.keyframes[paramId] || node.keyframes[paramId].length === 0) {
      const def = NODE_TYPES[node?.type]
      const pdef = def?.params.find(p => p.id === paramId)
      return pdef?.def ?? node?.params[paramId] ?? 0
    }
    const kfs = node.keyframes[paramId]
    if (time <= kfs[0].time) return kfs[0].value
    if (time >= kfs[kfs.length - 1].time) return kfs[kfs.length - 1].value
    for (let i = 0; i < kfs.length - 1; i++) {
      const a = kfs[i], b = kfs[i + 1]
      if (time >= a.time && time < b.time) {
        const t = a.time === b.time ? 0 : (time - a.time) / (b.time - a.time)
        return a.value + (b.value - a.value) * applyEasing(t, a.easing)
      }
    }
    return node.params[paramId] ?? 0
  }

  function applyEasing(t, easing) {
    switch (easing) {
      case 'easeIn': return t * t
      case 'easeOut': return 1 - (1 - t) * (1 - t)
      case 'easeInOut': return t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2
      case 'bounce': {
        const n1 = 7.5625, d1 = 2.75
        if (t < 1 / d1) return n1 * t * t
        if (t < 2 / d1) return n1 * (t -= 1.5 / d1) * t + 0.75
        if (t < 2.5 / d1) return n1 * (t -= 2.25 / d1) * t + 0.9375
        return n1 * (t -= 2.625 / d1) * t + 0.984375
      }
      case 'elastic': {
        if (t === 0 || t === 1) return t
        return -Math.pow(2, 10 * t - 10) * Math.sin((t * 10 - 10.75) * (2 * Math.PI / 3))
      }
      default: return t // linear
    }
  }

  function zoomIn() { zoom.value = Math.min(5, zoom.value * 1.2) }
  function zoomOut() { zoom.value = Math.max(0.1, zoom.value / 1.2) }
  function zoomFit() { zoom.value = 1; panX.value = 0; panY.value = 0 }

  function saveAsPreset() {
    presets.value.push({ id: generateId(), name: 'Preset ' + (presets.value.length + 1), nodes: JSON.parse(JSON.stringify(nodes.value)) })
  }
  function loadPreset(id) {
    const p = presets.value.find(x => x.id === id)
    if (p) {
      nodes.value = JSON.parse(JSON.stringify(p.nodes))
      pushDesignSnapshot()
    }
  }
  function insertTemplate(nodesData) {
    for (const n of nodesData) {
      n.id = generateId()
      n.x += 50
      n.y += 50
    }
    nodes.value.push(...nodesData)
    pushDesignSnapshot()
  }

  function updateNodeParam(nodeId, paramId, value) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (node) node.params[paramId] = value
    pushDesignSnapshot()
  }

  function updateNodePosition(nodeId, x, y) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (node) { node.x = x; node.y = y }
    pushDesignSnapshot()
  }

  function setViewerNode(viewerNum, nodeId) {
    if (viewerNum === 1) {
      viewer1NodeId.value = viewer1NodeId.value === nodeId ? null : nodeId
    } else if (viewerNum === 2) {
      viewer2NodeId.value = viewer2NodeId.value === nodeId ? null : nodeId
    }
  }

  function setViewer1(nodeId) {
    viewer1NodeId.value = viewer1NodeId.value === nodeId ? null : nodeId
  }

  function setViewer2(nodeId) {
    viewer2NodeId.value = viewer2NodeId.value === nodeId ? null : nodeId
  }

  // ============ MULTI-SELECTION ============
  function toggleNodeSelection(nodeId, additive = false) {
    if (!additive) {
      selectedNodeIds.value = new Set([nodeId])
    } else {
      if (selectedNodeIds.value.has(nodeId)) {
        selectedNodeIds.value.delete(nodeId)
        selectedNodeIds.value = new Set(selectedNodeIds.value)
      } else {
        selectedNodeIds.value = new Set([...selectedNodeIds.value, nodeId])
      }
    }
    selectedNodeId.value = nodeId
  }

  function selectAllNodes() {
    selectedNodeIds.value = new Set(nodes.value.map(n => n.id))
    if (nodes.value.length > 0) selectedNodeId.value = nodes.value[0].id
  }

  function clearSelection() {
    selectedNodeIds.value = new Set()
    selectedNodeId.value = null
    selectedConnectionId.value = null
  }

  // ============ GROUPS ============
  function groupNodes(nodeIds) {
    if (nodeIds.length < 2) return
    const groupId = generateId()
    const groupNodes = nodes.value.filter(n => nodeIds.includes(n.id))
    if (groupNodes.length < 2) return

    // Calculate bounding box
    let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity
    for (const n of groupNodes) {
      minX = Math.min(minX, n.x)
      minY = Math.min(minY, n.y)
      maxX = Math.max(maxX, n.x + 160)
      maxY = Math.max(maxY, n.y + 60)
    }

    groups.value.push({
      id: groupId,
      name: 'Group ' + (groups.value.length + 1),
      nodeIds: [...nodeIds],
      x: minX - 10,
      y: minY - 30,
      width: maxX - minX + 20,
      height: maxY - minY + 50,
      collapsed: false,
    })
    pushDesignSnapshot()
  }

  function ungroupNode(groupId) {
    groups.value = groups.value.filter(g => g.id !== groupId)
    pushDesignSnapshot()
  }

  function renameGroup(groupId, name) {
    const g = groups.value.find(g => g.id === groupId)
    if (g) g.name = name
  }

  // ============ BOOKMARKS ============
  function saveBookmark(slot) {
    bookmarks.value[slot] = {
      panX: panX.value,
      panY: panY.value,
      zoom: zoom.value,
    }
    pushDesignSnapshot()
  }

  function loadBookmark(slot) {
    const b = bookmarks.value[slot]
    if (b) {
      panX.value = b.panX
      panY.value = b.panY
      zoom.value = b.zoom
    }
  }

  function moveSelectedNodes(dx, dy) {
    for (const id of selectedNodeIds.value) {
      const node = nodes.value.find(n => n.id === id)
      if (node && !node.locked) {
        node.x = Math.max(0, node.x + dx)
        node.y = Math.max(0, node.y + dy)
      }
    }
  }

  function deleteSelectedNodes() {
    if (selectedNodeIds.value.size > 0) {
      for (const id of selectedNodeIds.value) {
        removeNode(id)
      }
      selectedNodeIds.value = new Set()
    } else if (selectedNodeId.value) {
      removeNode(selectedNodeId.value)
    }
  }

  function serialize() {
    return JSON.parse(JSON.stringify({
      composition: composition.value,
      nodes: nodes.value,
      connections: connections.value,
      presets: presets.value,
      groups: groups.value,
      bookmarks: bookmarks.value,
    }))
  }

  function deserialize(data) {
    if (!data) return
    if (data.composition) composition.value = data.composition
    if (data.nodes) nodes.value = data.nodes
    if (data.connections) connections.value = data.connections
    if (data.presets) presets.value = data.presets
    if (data.groups) groups.value = data.groups
    if (data.bookmarks) bookmarks.value = data.bookmarks
    selectedNodeId.value = null
    selectedNodeIds.value = new Set()
    selectedConnectionId.value = null
  }

  return {
    composition, nodes, connections, selectedNodeId, selectedNodeIds, selectedConnectionId, zoom, panX, panY, snapEnabled, presets, groups, bookmarks,
    selectedNode, outputNode, selectedNodes,
    addNode, removeNode, removeSelectedNode, duplicateSelectedNode, addConnection, removeConnection,
    addKeyframe, removeKeyframe, getParamValue,
    zoomIn, zoomOut, zoomFit, saveAsPreset, loadPreset, insertTemplate,
    updateNodeParam, updateNodePosition,
    viewer1NodeId, viewer2NodeId, setViewerNode, setViewer1, setViewer2,
    toggleNodeSelection, selectAllNodes, clearSelection, moveSelectedNodes, deleteSelectedNodes,
    groupNodes, ungroupNode, renameGroup,
    saveBookmark, loadBookmark,
  }
})