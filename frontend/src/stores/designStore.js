import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const generateId = () => crypto.randomUUID()

// ============== NODE TYPE DEFINITIONS ==============
// Compact node palette: sources, transforms, composites, effects, math, output
const NODE_DEFS = {
  // ---- Sources ----
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

  // ---- Transform / Composite ----
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

  // ---- Effects ----
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

  // ---- Math ----
  math: { cat: 'Math', label: 'Math', col: '#A78BFA', in: ['a', 'b'], out: ['out'], params: [
    { id: 'operation', label: 'Operation', type: 'select', def: 'add', options: [
      { value: 'add', label: 'A + B' }, { value: 'subtract', label: 'A - B' },
      { value: 'multiply', label: 'A × B' }, { value: 'divide', label: 'A ÷ B' },
      { value: 'min', label: 'Min(A, B)' }, { value: 'max', label: 'Max(A, B)' },
      { value: 'sin', label: 'sin(A)' }, { value: 'cos', label: 'cos(A)' },
    ]},
  ]},

  // ---- Output ----
  output: { cat: 'Output', label: 'Output', col: '#F472B6', in: ['in'], out: [], params: []},
}

export const NODE_TYPES = NODE_DEFS

// Helper to get full type info with computed fields
export function getNodeType(type) {
  const def = NODE_DEFS[type]
  if (!def) return null
  return {
    type,
    category: def.cat,
    label: def.label,
    color: def.col,
    inputs: def.in,
    outputs: def.out,
    params: def.params,
  }
}

// Easing functions for keyframe animation
export const EASING_TYPES = {
  linear:        { label: 'Linear',           fn: (t) => t },
  easeIn:        { label: 'Ease In',          fn: (t) => t * t },
  easeOut:       { label: 'Ease Out',         fn: (t) => 1 - (1 - t) * (1 - t) },
  easeInOut:     { label: 'Ease In/Out',      fn: (t) => t < 0.5 ? 2*t*t : 1 - Math.pow(-2*t+2, 2)/2 },
  easeInCubic:   { label: 'Ease In Cubic',    fn: (t) => t*t*t },
  easeOutCubic:  { label: 'Ease Out Cubic',   fn: (t) => 1 - Math.pow(1-t, 3) },
  easeInOutCubic:{ label: 'Ease In/Out Cubic',fn: (t) => t < 0.5 ? 4*t*t*t : 1 - Math.pow(-2*t+2, 3)/2 },
  bounce:        { label: 'Bounce',           fn: (t) => {
    const n1=7.5625, d1=2.75
    if (t<1/d1) return n1*t*t
    if (t<2/d1) return n1*(t-=1.5/d1)*t+0.75
    if (t<2.5/d1) return n1*(t-=2.25/d1)*t+0.9375
    return n1*(t-=2.625/d1)*t+0.984375
  }},
  elastic:       { label: 'Elastic',          fn: (t) => {
    const c4=(2*Math.PI)/3
    return t===0?0:t===1?1:Math.pow(2,-10*t)*Math.sin((t*10-0.75)*c4)+1
  }},
}

function makeDefaultParams(type) {
  const def = NODE_DEFS[type]
  if (!def) return {}
  const params = {}
  for (const p of def.params) {
    params[p.id] = p.def !== undefined ? p.def : (p.type === 'toggle' ? false : '')
  }
  // Apply node-specific defaults
  if (def.defaults) Object.assign(params, def.defaults)
  return params
}

// Create starter composition with a working example
function createStarterNodes() {
  const text = {
    id: generateId(), type: 'text', x: 80, y: 200,
    params: { ...makeDefaultParams('text'), text: 'Welcome to Gocut Design' },
    keyframes: {
      x: [
        { id: generateId(), time: 0,   value: 0,   easing: 'easeOut' },
        { id: generateId(), time: 1.5, value: 320, easing: 'easeInOut' },
      ],
      opacity: [
        { id: generateId(), time: 0,   value: 0, easing: 'easeOut' },
        { id: generateId(), time: 0.8, value: 1, easing: 'linear' },
      ],
    },
    visible: true, locked: false, label: 'Heading Text',
  }
  const glow = {
    id: generateId(), type: 'glow', x: 420, y: 200,
    params: { ...makeDefaultParams('glow') },
    keyframes: {
      intensity: [
        { id: generateId(), time: 0,   value: 0.2, easing: 'easeInOut' },
        { id: generateId(), time: 1.5, value: 2.0, easing: 'easeInOut' },
        { id: generateId(), time: 3,   value: 1.0, easing: 'easeInOut' },
      ],
    },
    visible: true, locked: false, label: 'Glow',
  }
  const out = {
    id: generateId(), type: 'output', x: 760, y: 200,
    params: {}, keyframes: {},
    visible: true, locked: false, label: 'Output',
  }
  return [text, glow, out]
}

function createStarterConnections(nodes) {
  return [
    { id: generateId(), fromNode: nodes[0].id, fromPort: 'out', toNode: nodes[1].id, toPort: 'in' },
    { id: generateId(), fromNode: nodes[1].id, fromPort: 'out', toNode: nodes[2].id, toPort: 'in' },
  ]
}

export const useDesignStore = defineStore('design', () => {
  // Composition metadata
  const composition = ref({
    name: 'Main Comp',
    width: 1920,
    height: 1080,
    fps: 30,
    duration: 5,
    background: '#0a0a0a',
  })

})