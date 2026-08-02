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

// ============== NODE TYPE DEFINITIONS ==============
export const NODE_TYPES = {
  media:    { cat: 'Sources',   label: 'Media In',   col: '#00D4FF', in: [],          out: ['out'], params: [
    { id: 'assetId',   label: 'Source',   type: 'asset' },
    { id: 'startTime', label: 'In',       type: 'number', def: 0, min: 0, step: 0.1, suffix: 's' },
    { id: 'duration',  label: 'Duration', type: 'number', def: 5, min: 0.1, step: 0.1, suffix: 's' },
  ]},
  text:     { cat: 'Sources',   label: 'Text',       col: '#F59E0B', in: [],          out: ['out'], params: [
    { id: 'text',       label: 'Text',  type: 'text' },
    { id: 'fontSize',   label: 'Size',  type: 'number', def: 48, min: 8, max: 300, step: 1, suffix: 'px' },
    { id: 'color',      label: 'Color', type: 'color' },
    { id: 'fontFamily', label: 'Font',  type: 'font' },
    { id: 'bold',       label: 'Bold',  type: 'toggle' },
    { id: 'italic',     label: 'Italic',type: 'toggle' },
  ], defaults: { text: 'Hello Gocut', color: '#FFFFFF' }},
  rectangle:{ cat: 'Sources',   label: 'Rectangle',  col: '#F59E0B', in: [],          out: ['out'], params: [
    { id: 'width',  label: 'Width',  type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'fill',   label: 'Fill',   type: 'color' },
    { id: 'cornerRadius', label: 'Radius', type: 'number', def: 0, min: 0, max: 200, step: 1, suffix: 'px' },
  ], defaults: { fill: '#00D4FF' }},
  ellipse:  { cat: 'Sources',   label: 'Ellipse',    col: '#F59E0B', in: [],          out: ['out'], params: [
    { id: 'width',  label: 'Width',  type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 200, min: 1, step: 1, suffix: 'px' },
    { id: 'fill',   label: 'Fill',   type: 'color' },
  ], defaults: { fill: '#EC4899' }},
  polygon:  { cat: 'Sources',   label: 'Polygon',    col: '#F59E0B', in: [],          out: ['out'], params: [
    { id: 'sides',  label: 'Sides',  type: 'number', def: 5, min: 3, max: 12, step: 1 },
    { id: 'radius', label: 'Radius', type: 'number', def: 100, min: 1, step: 1, suffix: 'px' },
    { id: 'fill',   label: 'Fill',   type: 'color' },
  ], defaults: { fill: '#10B981' }},
  star:     { cat: 'Sources',   label: 'Star',       col: '#F59E0B', in: [],          out: ['out'], params: [
    { id: 'numPoints',   label: 'Points',  type: 'number', def: 5, min: 3, max: 20, step: 1 },
    { id: 'innerRadius', label: 'Inner R', type: 'number', def: 40, min: 0, step: 1, suffix: 'px' },
    { id: 'outerRadius', label: 'Outer R', type: 'number', def: 100, min: 1, step: 1, suffix: 'px' },
    { id: 'fill',        label: 'Fill',    type: 'color' },
  ], defaults: { fill: '#F59E0B' }},
  gradient: { cat: 'Sources',   label: 'Gradient',   col: '#F59E0B', in: [],          out: ['out'], params: [
    { id: 'color1', label: 'Color 1', type: 'color' },
    { id: 'color2', label: 'Color 2', type: 'color' },
    { id: 'angle',  label: 'Angle',   type: 'number', def: 90, min: 0, max: 360, step: 1, suffix: '°' },
  ], defaults: { color1: '#00D4FF', color2: '#EC4899' }},
  transform: { cat: 'Transform', label: 'Transform', col: '#8B5CF6', in: ['in'],       out: ['out'], params: [
    { id: 'x',         label: 'X',         type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'y',         label: 'Y',         type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'scaleX',    label: 'Scale X',   type: 'number', def: 1, min: 0, max: 10, step: 0.05 },
    { id: 'scaleY',    label: 'Scale Y',   type: 'number', def: 1, min: 0, max: 10, step: 0.05 },
    { id: 'rotation',  label: 'Rotation',  type: 'number', def: 0, min: -720, max: 720, step: 1, suffix: '°' },
    { id: 'opacity',   label: 'Opacity',   type: 'number', def: 1, min: 0, max: 1, step: 0.01 },
  ]},
  merge:     { cat: 'Composite', label: 'Merge',     col: '#10B981', in: ['fg', 'bg'], out: ['out'], params: [
    { id: 'mode', label: 'Mode', type: 'select', def: 'normal', options: [
      { value: 'normal', label: 'Normal' }, { value: 'multiply', label: 'Multiply' },
      { value: 'screen', label: 'Screen' }, { value: 'overlay', label: 'Overlay' },
      { value: 'add', label: 'Add' }, { value: 'subtract', label: 'Subtract' },
      { value: 'difference', label: 'Difference' }, { value: 'lighten', label: 'Lighten' },
      { value: 'darken', label: 'Darken' }, { value: 'colordodge', label: 'Color Dodge' },
      { value: 'colorburn', label: 'Color Burn' },
    ]},
  ]},
  blur:        { cat: 'Effects', label: 'Blur',     col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'radius', label: 'Radius', type: 'number', def: 5, min: 0, max: 100, step: 0.5, suffix: 'px' },
  ]},
  glow:        { cat: 'Effects', label: 'Glow',     col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'intensity', label: 'Intensity', type: 'number', def: 1, min: 0, max: 5, step: 0.1 },
    { id: 'color',     label: 'Color',     type: 'color' },
    { id: 'radius',    label: 'Radius',    type: 'number', def: 10, min: 0, max: 100, step: 1, suffix: 'px' },
  ], defaults: { color: '#00D4FF' }},
  shadow:      { cat: 'Effects', label: 'Shadow',   col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'color',   label: 'Color',     type: 'color' },
    { id: 'blur',    label: 'Blur',      type: 'number', def: 8, min: 0, max: 100, step: 1, suffix: 'px' },
    { id: 'offsetX', label: 'Offset X',  type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'offsetY', label: 'Offset Y',  type: 'number', def: 4, step: 1, suffix: 'px' },
  ], defaults: { color: '#000000' }},
  colorCorrect:{ cat: 'Effects', label: 'Color Correct', col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'brightness', label: 'Brightness', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'contrast',   label: 'Contrast',   type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'saturation', label: 'Saturation', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'hue',        label: 'Hue',        type: 'number', def: 0, min: -180, max: 180, step: 1, suffix: '°' },
  ]},
  chromaKey:   { cat: 'Effects', label: 'Chroma Key', col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'keyColor',   label: 'Key Color',  type: 'color' },
    { id: 'similarity', label: 'Similarity', type: 'number', def: 0.4, min: 0, max: 1, step: 0.01 },
    { id: 'smoothness', label: 'Smoothness', type: 'number', def: 0.1, min: 0, max: 1, step: 0.01 },
  ], defaults: { keyColor: '#00FF00' }},
  math: { cat: 'Math', label: 'Math', col: '#A78BFA', in: ['a', 'b'], out: ['out'], params: [
    { id: 'operation', label: 'Operation', type: 'select', def: 'add', options: [
      { value: 'add', label: 'A + B' }, { value: 'subtract', label: 'A - B' },
      { value: 'multiply', label: 'A × B' }, { value: 'divide', label: 'A ÷ B' },
      { value: 'min', label: 'Min(A, B)' }, { value: 'max', label: 'Max(A, B)' },
      { value: 'sin', label: 'sin(A)' }, { value: 'cos', label: 'cos(A)' },
    ]},
  ]},
  output: { cat: 'Output', label: 'Output', col: '#F472B6', in: ['in'], out: [], params: []},

  // ============ EXTENDED NODE TYPES ============
  solidColor: { cat: 'Sources', label: 'Solid Color', col: '#F59E0B', in: [], out: ['out'], params: [
    { id: 'color', label: 'Color', type: 'color' },
    { id: 'width', label: 'Width', type: 'number', def: 1920, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 1080, min: 1, step: 1, suffix: 'px' },
  ], defaults: { color: '#00D4FF' }},

  noise: { cat: 'Sources', label: 'Noise', col: '#F59E0B', in: [], out: ['out'], params: [
    { id: 'noiseType', label: 'Type', type: 'select', def: 'perlin', options: [
      { value: 'perlin', label: 'Perlin' }, { value: 'fractal', label: 'Fractal' }, { value: 'white', label: 'White' },
    ]},
    { id: 'scale', label: 'Scale', type: 'number', def: 50, min: 1, max: 500, step: 1 },
    { id: 'octaves', label: 'Octaves', type: 'number', def: 4, min: 1, max: 8, step: 1 },
    { id: 'seed', label: 'Seed', type: 'number', def: 0, min: 0, max: 9999, step: 1 },
  ]},

  crop: { cat: 'Transform', label: 'Crop', col: '#8B5CF6', in: ['in'], out: ['out'], params: [
    { id: 'x', label: 'X', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'y', label: 'Y', type: 'number', def: 0, step: 1, suffix: 'px' },
    { id: 'width', label: 'Width', type: 'number', def: 1920, min: 1, step: 1, suffix: 'px' },
    { id: 'height', label: 'Height', type: 'number', def: 1080, min: 1, step: 1, suffix: 'px' },
  ]},

  cornerPin: { cat: 'Transform', label: 'Corner Pin', col: '#8B5CF6', in: ['in'], out: ['out'], params: [
    { id: 'tlX', label: 'Top-Left X', type: 'number', def: 0, step: 1 },
    { id: 'tlY', label: 'Top-Left Y', type: 'number', def: 0, step: 1 },
    { id: 'trX', label: 'Top-Right X', type: 'number', def: 1920, step: 1 },
    { id: 'trY', label: 'Top-Right Y', type: 'number', def: 0, step: 1 },
    { id: 'blX', label: 'Bot-Left X', type: 'number', def: 0, step: 1 },
    { id: 'blY', label: 'Bot-Left Y', type: 'number', def: 1080, step: 1 },
    { id: 'brX', label: 'Bot-Right X', type: 'number', def: 1920, step: 1 },
    { id: 'brY', label: 'Bot-Right Y', type: 'number', def: 1080, step: 1 },
  ]},

  channelSplit: { cat: 'Effects', label: 'Channel Split', col: '#EC4899', in: ['in'], out: ['r', 'g', 'b', 'a'], params: []},
  channelMerge: { cat: 'Effects', label: 'Channel Merge', col: '#EC4899', in: ['r', 'g', 'b', 'a'], out: ['out'], params: []},

  levels: { cat: 'Effects', label: 'Levels', col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'inBlack', label: 'In Black', type: 'number', def: 0, min: 0, max: 255, step: 1 },
    { id: 'inWhite', label: 'In White', type: 'number', def: 255, min: 0, max: 255, step: 1 },
    { id: 'gamma', label: 'Gamma', type: 'number', def: 1, min: 0.1, max: 10, step: 0.1 },
    { id: 'outBlack', label: 'Out Black', type: 'number', def: 0, min: 0, max: 255, step: 1 },
    { id: 'outWhite', label: 'Out White', type: 'number', def: 255, min: 0, max: 255, step: 1 },
  ]},

  invert: { cat: 'Effects', label: 'Invert', col: '#EC4899', in: ['in'], out: ['out'], params: []},

  temperature: { cat: 'Effects', label: 'Temperature', col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'temperature', label: 'Temperature', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
    { id: 'tint', label: 'Tint', type: 'number', def: 0, min: -1, max: 1, step: 0.01 },
  ]},

  directionalBlur: { cat: 'Effects', label: 'Directional Blur', col: '#EC4899', in: ['in'], out: ['out'], params: [
    { id: 'angle', label: 'Angle', type: 'number', def: 0, min: 0, max: 360, step: 1, suffix: '°' },
    { id: 'distance', label: 'Distance', type: 'number', def: 10, min: 0, max: 200, step: 1, suffix: 'px' },
  ]},

  mask: { cat: 'Effects', label: 'Mask', col: '#EC4899', in: ['in'], out: ['out'], params: [
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
    serialize, deserialize,
    toggleNodeSelection, selectAllNodes, clearSelection, moveSelectedNodes, deleteSelectedNodes,
    groupNodes, ungroupNode, renameGroup,
    saveBookmark, loadBookmark,
  }
})