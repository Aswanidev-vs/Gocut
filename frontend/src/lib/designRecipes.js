// Fusion-style macros for Gocut's Design workspace.
//
// A recipe is one human-language move ("Fade Up", "Lower Third") that builds a
// pre-wired node cluster with its keyframes already set, so making a small
// animation takes one click instead of twelve. Everything here composes the
// existing node types — no new engine concepts.

const COL = 240
const ROW = 110

function at(col, row = 0) {
  return { x: 80 + col * COL, y: 90 + row * ROW }
}

// addNode auto-wires the new node to whatever is selected, which makes cluster
// building order dependent. Recipes wire their own graph, so the selection is
// cleared first; params are applied after creation so the type's other defaults
// survive.
function node(store, type, opts = {}) {
  const { params, ...rest } = opts
  store.clearSelection()
  const created = store.addNode(type, rest)
  if (!created) return null
  if (params) {
    for (const [id, value] of Object.entries(params)) store.updateNodeParam(created.id, id, value)
  }
  return created
}

function link(store, from, to, toPort = 'in', fromPort = 'out') {
  store.addConnection(from.id, fromPort, to.id, toPort)
}

function keys(store, target, param, frames) {
  for (const [time, value, easing] of frames) {
    store.addKeyframe(target.id, param, time, value, easing || 'easeOut')
  }
}

// Every title recipe shares the same spine: a backdrop, an Over (Merge) and an
// Output. The move itself plugs into the Merge's foreground.
function titleSpine(store, backdrop) {
  const bg = node(store, 'background', {
    ...at(0, 0),
    label: 'Backdrop',
    params: backdrop || { type: 'vertical', color: '#0B0B0F', color2: '#16161C' },
  })
  const merge = node(store, 'merge', { ...at(2, 0), label: 'Over' })
  const out = node(store, 'output', { ...at(3, 0), label: 'Output' })
  link(store, bg, merge, 'bg')
  link(store, merge, out)
  return { bg, merge, out }
}

function headline(store, text, size = 110, color = '#EDEDF0') {
  return node(store, 'text', {
    ...at(0, 1),
    label: 'Headline',
    params: { text, fontSize: size, color },
  })
}

const TITLE_RECIPES = [
  {
    id: 'title-fade-up',
    name: 'Fade Up',
    group: 'Titles',
    blurb: 'Text fades in and lifts into place.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'Your title')
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Rise' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'opacity', [[0, 0, 'easeOut'], [0.7, 1, 'easeOut']])
      keys(store, xf, 'y', [[0, 40, 'easeOut'], [0.7, 0, 'easeOut']])
      return xf
    },
  },
  {
    id: 'title-slide-in',
    name: 'Slide In',
    group: 'Titles',
    blurb: 'Text slides in from the left edge.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'Breaking news')
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Slide' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'x', [[0, -900, 'easeOut'], [1, 0, 'easeOut']])
      keys(store, xf, 'opacity', [[0, 0, 'easeOut'], [0.35, 1, 'easeOut']])
      return xf
    },
  },
  {
    id: 'title-pop-in',
    name: 'Pop In',
    group: 'Titles',
    blurb: 'Text snaps in with a small overshoot.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'Pop!')
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Pop' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'scaleX', [[0, 0.7, 'easeOut'], [0.45, 1.12, 'easeOut'], [0.75, 1, 'easeInOut']])
      keys(store, xf, 'scaleY', [[0, 0.7, 'easeOut'], [0.45, 1.12, 'easeOut'], [0.75, 1, 'easeInOut']])
      keys(store, xf, 'opacity', [[0, 0, 'easeOut'], [0.2, 1, 'linear']])
      return xf
    },
  },
  {
    id: 'title-typewriter',
    name: 'Typewriter',
    group: 'Titles',
    blurb: 'Characters reveal one after another.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = node(store, 'textPlus', {
        ...at(0, 1),
        label: 'Write-On',
        params: { text: 'Type it out…', fontSize: 96, color: '#FFB020' },
      })
      link(store, src, merge, 'fg')
      // Text+ carries the reveal natively: Write On Start/End is the fraction of
      // the string that is visible, so animating End is a real character reveal.
      keys(store, src, 'writeOnEnd', [[0.15, 0, 'linear'], [1.7, 1, 'linear']])
      return src
    },
  },
  {
    id: 'title-bounce-in',
    name: 'Bounce In',
    group: 'Titles',
    blurb: 'Text drops in and settles.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'Ta-da')
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Drop' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'y', [[0, -420, 'easeIn'], [1.1, 0, 'bounce']])
      keys(store, xf, 'opacity', [[0, 0, 'linear'], [0.3, 1, 'linear']])
      return xf
    },
  },
  {
    id: 'title-spin-in',
    name: 'Spin In',
    group: 'Titles',
    blurb: 'Text rotates and grows into place.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'Wheee')
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Spin' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'rotation', [[0, -140, 'easeOut'], [0.9, 0, 'easeOut']])
      keys(store, xf, 'scaleX', [[0, 0.6, 'easeOut'], [0.7, 1, 'easeOut']])
      keys(store, xf, 'scaleY', [[0, 0.6, 'easeOut'], [0.7, 1, 'easeOut']])
      keys(store, xf, 'opacity', [[0, 0, 'linear'], [0.3, 1, 'linear']])
      return xf
    },
  },
  {
    id: 'title-lower-third',
    name: 'Lower Third',
    group: 'Titles',
    blurb: 'Name bar with an accent rule that slides in as one unit.',
    build(store) {
      const bg = node(store, 'background', {
        ...at(0, 0),
        label: 'Backdrop',
        params: { type: 'vertical', color: '#0B0B0F', color2: '#16161C' },
      })
      // Merge chain: backdrop -> bar -> accent rule -> name -> output.
      const bar = node(store, 'rectangle', {
        ...at(0, 2),
        label: 'Bar',
        params: { width: 900, height: 150, fill: '#101015', cornerRadius: 10 },
      })
      const barXf = node(store, 'transform', { ...at(1, 2), label: 'Bar Move', params: { y: 330 } })
      link(store, bar, barXf)
      const barOver = node(store, 'merge', { ...at(2, 1), label: 'Bar Over' })
      link(store, bg, barOver, 'bg')
      link(store, barXf, barOver, 'fg')

      const accent = node(store, 'rectangle', {
        ...at(0, 3),
        label: 'Accent Rule',
        params: { width: 8, height: 96, fill: '#FFB020', cornerRadius: 4 },
      })
      const accentXf = node(store, 'transform', { ...at(1, 3), label: 'Accent Move', params: { y: 330 } })
      link(store, accent, accentXf)
      const accentOver = node(store, 'merge', { ...at(2, 2), label: 'Accent Over' })
      link(store, barOver, accentOver, 'bg')
      link(store, accentXf, accentOver, 'fg')

      const label = node(store, 'text', {
        ...at(0, 4),
        label: 'Name',
        params: { text: 'Jane Doe · Director', fontSize: 54, color: '#EDEDF0' },
      })
      const labelXf = node(store, 'transform', { ...at(1, 4), label: 'Name Move', params: { y: 330 } })
      link(store, label, labelXf)
      const nameOver = node(store, 'merge', { ...at(2, 3), label: 'Name Over' })
      link(store, accentOver, nameOver, 'bg')
      link(store, labelXf, nameOver, 'fg')

      const out = node(store, 'output', { ...at(3, 3), label: 'Output' })
      link(store, nameOver, out)

      // Bar, rule and name slide in from off-frame and settle in their own spots.
      keys(store, barXf, 'x', [[0, -1100, 'easeOut'], [0.9, 0, 'easeOut']])
      keys(store, accentXf, 'x', [[0, -1100, 'easeOut'], [0.9, -410, 'easeOut']])
      keys(store, labelXf, 'x', [[0, -1100, 'easeOut'], [0.9, -320, 'easeOut']])
      return labelXf
    },
  },
]

const MOTION_RECIPES = [
  {
    id: 'motion-drift',
    name: 'Slow Drift',
    group: 'Motion',
    blurb: 'Content floats slowly across the frame.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'Drifting', 120)
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Drift' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'x', [[0, -90, 'smooth'], [3, 90, 'smooth']])
      keys(store, xf, 'y', [[0, 30, 'smooth'], [3, -30, 'smooth']])
      return xf
    },
  },
  {
    id: 'motion-pulse',
    name: 'Pulse Loop',
    group: 'Motion',
    blurb: 'A steady heartbeat scale that loops cleanly.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'PULSE', 120, '#FFB020')
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Pulse' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      const beat = [[0, 1, 'easeInOut'], [0.5, 1.08, 'easeInOut'], [1, 1, 'easeInOut'], [1.5, 1.08, 'easeInOut'], [2, 1, 'easeInOut']]
      keys(store, xf, 'scaleX', beat)
      keys(store, xf, 'scaleY', beat)
      return xf
    },
  },
  {
    id: 'motion-track-across',
    name: 'Track Across',
    group: 'Motion',
    blurb: 'A constant linear move, good for tickers.',
    build(store) {
      const { merge } = titleSpine(store)
      const src = headline(store, 'TICKER  ·  TICKER  ·  TICKER', 90)
      const xf = node(store, 'transform', { ...at(1, 1), label: 'Track' })
      link(store, src, xf)
      link(store, xf, merge, 'fg')
      keys(store, xf, 'x', [[0, -1000, 'linear'], [2.5, 1000, 'linear']])
      return xf
    },
  },
]

const SHAPE_RECIPES = [
  {
    id: 'shape-badge',
    name: 'Round Badge',
    group: 'Shapes',
    blurb: 'A circle with a label centred inside it.',
    build(store) {
      const bg = node(store, 'background', {
        ...at(0, 0),
        label: 'Backdrop',
        params: { type: 'vertical', color: '#0B0B0F', color2: '#16161C' },
      })
      const disc = node(store, 'ellipse', {
        ...at(0, 1),
        label: 'Disc',
        params: { width: 340, height: 340, fill: '#00D4FF' },
      })
      const discXf = node(store, 'transform', { ...at(1, 1), label: 'Disc Move', params: { y: -40 } })
      link(store, disc, discXf)
      const discOver = node(store, 'merge', { ...at(2, 1), label: 'Disc Over' })
      link(store, bg, discOver, 'bg')
      link(store, discXf, discOver, 'fg')

      const label = node(store, 'text', {
        ...at(0, 2),
        label: 'Label',
        params: { text: 'NEW', fontSize: 64, color: '#08080A', bold: true },
      })
      const labelXf = node(store, 'transform', { ...at(1, 2), label: 'Label Move', params: { y: -40 } })
      link(store, label, labelXf)
      const labelOver = node(store, 'merge', { ...at(2, 2), label: 'Label Over' })
      link(store, discOver, labelOver, 'bg')
      link(store, labelXf, labelOver, 'fg')
      const out = node(store, 'output', { ...at(3, 2), label: 'Output' })
      link(store, labelOver, out)

      keys(store, discXf, 'scaleX', [[0, 0.2, 'easeOut'], [0.6, 1.15, 'easeOut'], [0.85, 1, 'easeInOut']])
      keys(store, discXf, 'scaleY', [[0, 0.2, 'easeOut'], [0.6, 1.15, 'easeOut'], [0.85, 1, 'easeInOut']])
      keys(store, labelXf, 'scaleX', [[0, 0.2, 'easeOut'], [0.6, 1.15, 'easeOut'], [0.85, 1, 'easeInOut']])
      keys(store, labelXf, 'scaleY', [[0, 0.2, 'easeOut'], [0.6, 1.15, 'easeOut'], [0.85, 1, 'easeInOut']])
      keys(store, labelXf, 'opacity', [[0, 0, 'linear'], [0.35, 1, 'linear']])
      return discXf
    },
  },
  {
    id: 'shape-gradient-backdrop',
    name: 'Gradient Backdrop',
    group: 'Shapes',
    blurb: 'A soft two-colour wash to build on.',
    build(store) {
      const bg = node(store, 'background', {
        ...at(0, 0),
        label: 'Wash',
        params: { type: 'radial', color: '#16203A', color2: '#08080A' },
      })
      const out = node(store, 'output', { ...at(1, 0), label: 'Output' })
      link(store, bg, out)
      return bg
    },
  },
  {
    id: 'shape-noise-card',
    name: 'Noise Plate',
    group: 'Shapes',
    blurb: 'Fractal noise for textures and grain passes.',
    build(store) {
      const noise = node(store, 'fastNoise', {
        ...at(0, 0),
        label: 'Noise',
        params: { scale: 24, detail: 5, contrast: 1.8, color1: '#0B0B0F', color2: '#FFB020' },
      })
      const out = node(store, 'output', { ...at(1, 0), label: 'Output' })
      link(store, noise, out)
      keys(store, noise, 'contrast', [[0, 1.2, 'easeInOut'], [2, 2.4, 'easeInOut']])
      return noise
    },
  },
]

const COMPOSITE_RECIPES = [
  {
    id: 'composite-glow-title',
    name: 'Glowing Title',
    group: 'Composite',
    blurb: 'Text with a coloured halo, over a dark wash.',
    build(store) {
      const { merge } = titleSpine(store, { type: 'radial', color: '#101828', color2: '#08080A' })
      const src = headline(store, 'AFTERGLOW', 100, '#EDEDF0')
      const glow = node(store, 'glow', {
        ...at(1, 1),
        label: 'Glow',
        params: { intensity: 2, radius: 26, color: '#00D4FF' },
      })
      link(store, src, glow)
      link(store, glow, merge, 'fg')
      keys(store, glow, 'intensity', [[0, 0.4, 'easeInOut'], [1.2, 2.6, 'easeInOut'], [2.4, 1.4, 'easeInOut']])
      return glow
    },
  },
  {
    id: 'composite-blend-noise',
    name: 'Texture Blend',
    group: 'Composite',
    blurb: 'Noise laid over a wash with Screen blending.',
    build(store) {
      const bg = node(store, 'background', {
        ...at(0, 0),
        label: 'Wash',
        params: { type: 'horizontal', color: '#0B1020', color2: '#2A1440' },
      })
      const noise = node(store, 'fastNoise', {
        ...at(0, 1),
        label: 'Texture',
        params: { scale: 30, detail: 4, contrast: 1.4, color1: '#000000', color2: '#7C5CFF' },
      })
      const merge = node(store, 'merge', { ...at(1, 0), label: 'Screen' })
      store.updateNodeParam(merge.id, 'mode', 'screen')
      link(store, bg, merge, 'bg')
      link(store, noise, merge, 'fg')
      const out = node(store, 'output', { ...at(2, 0), label: 'Output' })
      link(store, merge, out)
      keys(store, noise, 'scale', [[0, 12, 'linear'], [4, 60, 'linear']])
      return merge
    },
  },
  {
    id: 'composite-key-demo',
    name: 'Key + Backdrop',
    group: 'Composite',
    blurb: 'Keys a green screen away to reveal a new backdrop.',
    build(store) {
      const bg = node(store, 'background', {
        ...at(0, 2),
        label: 'New Backdrop',
        params: { type: 'radial', color: '#1A2A44', color2: '#08080A' },
      })
      // A green plate stands in for the clip: the keyer visibly removes it, so
      // the recipe demonstrates the technique without needing source media.
      const src = node(store, 'rectangle', {
        ...at(0, 0),
        label: 'Green Screen',
        params: { width: 700, height: 700, fill: '#00FF00' },
      })
      const keyer = node(store, 'chromaKey', {
        ...at(1, 0),
        label: 'Chroma Key',
        params: { keyColor: '#00FF00', similarity: 0.42, smoothness: 0.12 },
      })
      link(store, src, keyer)
      const merge = node(store, 'merge', { ...at(2, 1), label: 'Over' })
      link(store, bg, merge, 'bg')
      link(store, keyer, merge, 'fg')
      const out = node(store, 'output', { ...at(3, 1), label: 'Output' })
      link(store, merge, out)
      return keyer
    },
  },
]

// Looks are effects rather than standalone moves: with a node selected they are
// spliced into the chain right after it (Fusion's insert-after behaviour);
// otherwise they get a demo source so the result is still visible.
export const LOOK_RECIPES = [
  {
    id: 'look-glow',
    name: 'Glow',
    group: 'Look',
    blurb: 'Coloured halo around the result.',
    type: 'glow',
    params: { intensity: 1.6, radius: 22, color: '#FFB020' },
  },
  {
    id: 'look-defocus',
    name: 'Defocus',
    group: 'Look',
    blurb: 'Soft, out-of-focus background blur.',
    type: 'blur',
    params: { radius: 14 },
  },
  {
    id: 'look-color-pop',
    name: 'Color Pop',
    group: 'Look',
    blurb: 'Punchier saturation and highlights.',
    type: 'colorCorrector',
    params: { saturation: 1.45, gain: 1.12, gamma: 1.05 },
  },
  {
    id: 'look-shadow',
    name: 'Drop Shadow',
    group: 'Look',
    blurb: 'Lift the content off the backdrop.',
    type: 'shadow',
    params: { color: '#000000', blur: 26, offsetX: 0, offsetY: 18 },
  },
  {
    id: 'look-chroma',
    name: 'Chroma Key',
    group: 'Look',
    blurb: 'Remove a green screen from the source.',
    type: 'chromaKey',
    params: { keyColor: '#00FF00', similarity: 0.4, smoothness: 0.1 },
  },
  {
    id: 'look-cool-grade',
    name: 'Cool Grade',
    group: 'Look',
    blurb: 'Cool, low-contrast night grade.',
    type: 'colorCorrector',
    params: { tint: -45, lift: -0.05, gamma: 1.1 },
  },
]

export const RECIPES = [
  ...TITLE_RECIPES,
  ...MOTION_RECIPES,
  ...SHAPE_RECIPES,
  ...COMPOSITE_RECIPES,
]

export const RECIPE_GROUPS = ['Titles', 'Motion', 'Shapes', 'Composite']

export function applyRecipe(store, recipe) {
  if (!recipe) return null
  store.clearSelection()
  return recipe.build(store)
}

export function applyLook(store, look) {
  if (!look) return null
  const target = store.selectedNode
  const usable = target && target.type !== 'output' && target.type !== 'mediaOut'

  const effect = node(store, look.type, {
    x: usable ? target.x + COL : 380,
    y: usable ? target.y : 220,
    label: look.name,
    params: look.params,
  })
  if (!effect) return null

  if (usable) {
    link(store, target, effect)
    // Heal the chain: whatever consumed the target now consumes the effect,
    // otherwise the old wire would bypass it.
    for (const c of store.connections.filter(c => c.fromNode === target.id && c.toNode !== effect.id)) {
      store.removeConnection(c.id)
      store.addConnection(effect.id, 'out', c.toNode, c.toPort)
    }
  } else {
    const demo = node(store, 'text', {
      ...at(0, 1),
      label: 'Demo Source',
      params: { text: look.name, fontSize: 96, color: '#EDEDF0' },
    })
    if (demo) link(store, demo, effect)
    // With nothing selected there may be no Output yet either, and an effect
    // wired to nothing renders nowhere. Feed the existing Output, or add one.
    const out = store.nodes.find(n => n.type === 'output' || n.type === 'mediaOut')
    if (out) {
      link(store, effect, out, 'in')
    } else {
      const created = node(store, 'output', { ...at(2, 1), label: 'Output' })
      if (created) link(store, effect, created)
    }
  }

  store.selectedNodeId = effect.id
  return effect
}
