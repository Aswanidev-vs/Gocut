import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const generateId = () => crypto.randomUUID()

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
  const selectedConnectionId = ref(null)
  const zoom = ref(1)
  const panX = ref(0)
  const panY = ref(0)
  const snapEnabled = ref(true)
  const presets = ref([])

  const selectedNode = computed(() => nodes.value.find(n => n.id === selectedNodeId.value) || null)
  const outputNode = computed(() => nodes.value.find(n => n.type === 'output') || null)

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
    return node
  }

  function removeNode(nodeId) {
    connections.value = connections.value.filter(c => c.fromNode !== nodeId && c.toNode !== nodeId)
    nodes.value = nodes.value.filter(n => n.id !== nodeId)
    if (selectedNodeId.value === nodeId) selectedNodeId.value = null
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
  }

  function addConnection(fromNode, fromPort, toNode, toPort) {
    connections.value = connections.value.filter(c => !(c.toNode === toNode && c.toPort === toPort))
    connections.value.push({ id: generateId(), fromNode, fromPort, toNode, toPort })
  }

  function removeConnection(id) {
    connections.value = connections.value.filter(c => c.id !== id)
  }

  function addKeyframe(nodeId, paramId, time, value, easing = 'linear') {
    const node = nodes.value.find(n => n.id === nodeId)
    if (!node) return
    if (!node.keyframes[paramId]) node.keyframes[paramId] = []
    node.keyframes[paramId] = node.keyframes[paramId].filter(k => Math.abs(k.time - time) > 0.01)
    node.keyframes[paramId].push({ id: generateId(), time, value, easing })
    node.keyframes[paramId].sort((a, b) => a.time - b.time)
  }

  function removeKeyframe(nodeId, paramId, keyframeId) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (!node || !node.keyframes[paramId]) return
    node.keyframes[paramId] = node.keyframes[paramId].filter(k => k.id !== keyframeId)
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
        return a.value + (b.value - a.value) * t
      }
    }
    return node.params[paramId] ?? 0
  }

  function zoomIn() { zoom.value = Math.min(5, zoom.value * 1.2) }
  function zoomOut() { zoom.value = Math.max(0.1, zoom.value / 1.2) }
  function zoomFit() { zoom.value = 1; panX.value = 0; panY.value = 0 }

  function saveAsPreset() {
    presets.value.push({ id: generateId(), name: 'Preset ' + (presets.value.length + 1), nodes: JSON.parse(JSON.stringify(nodes.value)) })
  }
  function loadPreset(id) {
    const p = presets.value.find(x => x.id === id)
    if (p) nodes.value = JSON.parse(JSON.stringify(p.nodes))
  }
  function insertTemplate(nodesData) {
    for (const n of nodesData) {
      n.id = generateId()
      n.x += 50
      n.y += 50
    }
    nodes.value.push(...nodesData)
  }

  function updateNodeParam(nodeId, paramId, value) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (node) node.params[paramId] = value
  }

  function updateNodePosition(nodeId, x, y) {
    const node = nodes.value.find(n => n.id === nodeId)
    if (node) { node.x = x; node.y = y }
  }

  return {
    composition, nodes, connections, selectedNodeId, selectedConnectionId, zoom, panX, panY, snapEnabled, presets,
    selectedNode, outputNode,
    addNode, removeNode, removeSelectedNode, duplicateSelectedNode, addConnection, removeConnection,
    addKeyframe, removeKeyframe, getParamValue,
    zoomIn, zoomOut, zoomFit, saveAsPreset, loadPreset, insertTemplate,
    updateNodeParam, updateNodePosition,
  }
})