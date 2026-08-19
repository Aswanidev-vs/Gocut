<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Sparkles, Wand2, Box, Play, ArrowRight, Lightbulb, Zap, Layers, Video, Palette } from 'lucide-vue-next'
import { useDesignStore } from '../../stores/designStore'

const emit = defineEmits(['start-template', 'start-blank', 'start-effect'])
const designStore = useDesignStore()

const canvasDemoRef = ref(null)
let demoAnimId = null
let demoTime = 0

// Built-in starter Fusion animation examples
const DEMO_PRESETS = [
  {
    id: 'cyberpunk_title',
    title: 'Neon Kinetic Title',
    category: 'Kinetic Text+',
    desc: 'Text+ with write-on animation, chromatic glow, and particle drift over background.',
    badge: 'Popular',
    icon: Sparkles,
    color: '#00D4FF',
    accent: 'from-cyan-500/20 to-blue-500/5 border-cyan-500/30 hover:border-cyan-400',
    setup: () => {
      designStore.nodes = []
      designStore.connections = []
      const bg = designStore.addNode('background', { x: 80, y: 140, label: 'Background', params: { type: 'radial', color: '#090d16', color2: '#020408' } })
      const fn = designStore.addNode('fastNoise', { x: 80, y: 280, label: 'DustNoise', params: { detail: 5, contrast: 2.2, scale: 35, seethe: 0.8, color1: '#001122', color2: '#00ffee' } })
      const mrg1 = designStore.addNode('merge', { x: 300, y: 200, label: 'Merge_Noise', params: { mode: 'screen', blend: 0.35 } })
      const txt = designStore.addNode('textPlus', {
        x: 300, y: 340, label: 'KineticTitle',
        params: { text: 'CYBERPUNK 2088', fontSize: 68, color: '#00D4FF', fontFamily: 'monospace', tracking: 6, bold: true },
        keyframes: {
          writeOnStart: [{ id: 'k1', time: 0, value: 0, easing: 'smooth' }, { id: 'k2', time: 1.5, value: 0, easing: 'smooth' }],
          writeOnEnd: [{ id: 'k3', time: 0, value: 0, easing: 'smooth' }, { id: 'k4', time: 1.8, value: 1, easing: 'smooth' }]
        }
      })
      const glow = designStore.addNode('glow', { x: 500, y: 340, label: 'CyanGlow', params: { intensity: 2.4, color: '#00D4FF', radius: 24 } })
      const mrg2 = designStore.addNode('merge', { x: 680, y: 240, label: 'Merge_Title', params: { mode: 'normal' } })
      const out = designStore.addNode('mediaOut', { x: 880, y: 240, label: 'MediaOut' })

      if (bg && fn && mrg1 && txt && glow && mrg2 && out) {
        designStore.addConnection(bg.id, 'out', mrg1.id, 'bg')
        designStore.addConnection(fn.id, 'out', mrg1.id, 'fg')
        designStore.addConnection(mrg1.id, 'out', mrg2.id, 'bg')
        designStore.addConnection(txt.id, 'out', glow.id, 'in')
        designStore.addConnection(glow.id, 'out', mrg2.id, 'fg')
        designStore.addConnection(mrg2.id, 'out', out.id, 'in')
        designStore.setViewer1(txt.id)
        designStore.setViewer2(mrg2.id)
      }
    }
  },
  {
    id: 'mask_transition',
    title: 'Dynamic Mask Reveal',
    category: 'Motion Graphics',
    desc: 'Soft edge polygon mask slicing across foreground video into dual-toned background.',
    badge: 'Trending',
    icon: Wand2,
    color: '#EC4899',
    accent: 'from-pink-500/20 to-purple-500/5 border-pink-500/30 hover:border-pink-400',
    setup: () => {
      designStore.nodes = []
      designStore.connections = []
      const bg = designStore.addNode('background', { x: 80, y: 120, label: 'DarkBg', params: { color: '#0d0d18' } })
      const fg = designStore.addNode('background', { x: 80, y: 260, label: 'MagentaBanner', params: { color: '#EC4899' } })
      const xf = designStore.addNode('transform', {
        x: 280, y: 260, label: 'BannerXf',
        params: { scaleX: 1, scaleY: 0.35, rotation: -12, opacity: 0.9 },
        keyframes: {
          x: [{ id: 'k1', time: 0, value: -600, easing: 'smooth' }, { id: 'k2', time: 1.2, value: 0, easing: 'smooth' }],
          opacity: [{ id: 'k3', time: 0, value: 0, easing: 'easeOut' }, { id: 'k4', time: 0.8, value: 1, easing: 'linear' }]
        }
      })
      const txt = designStore.addNode('textPlus', { x: 280, y: 400, label: 'SlantedText', params: { text: 'REVEAL TRANSITION', fontSize: 52, color: '#FFFFFF', bold: true } })
      const mrg1 = designStore.addNode('merge', { x: 480, y: 200, label: 'Merge_Banner', params: { mode: 'normal' } })
      const mrg2 = designStore.addNode('merge', { x: 680, y: 240, label: 'Merge_Final', params: { mode: 'normal' } })
      const out = designStore.addNode('mediaOut', { x: 880, y: 240, label: 'MediaOut' })

      if (bg && fg && xf && txt && mrg1 && mrg2 && out) {
        designStore.addConnection(bg.id, 'out', mrg1.id, 'bg')
        designStore.addConnection(fg.id, 'out', xf.id, 'in')
        designStore.addConnection(xf.id, 'out', mrg1.id, 'fg')
        designStore.addConnection(mrg1.id, 'out', mrg2.id, 'bg')
        designStore.addConnection(txt.id, 'out', mrg2.id, 'fg')
        designStore.addConnection(mrg2.id, 'out', out.id, 'in')
        designStore.setViewer1(xf.id)
        designStore.setViewer2(mrg2.id)
      }
    }
  },
  {
    id: 'film_grade',
    title: 'Cinematic Color Grade',
    category: 'Color & Look',
    desc: 'Full Lift/Gamma/Gain color balance with cinematic soft vignette and contrast curves.',
    badge: 'Pro Look',
    icon: Palette,
    color: '#10B981',
    accent: 'from-emerald-500/20 to-teal-500/5 border-emerald-500/30 hover:border-emerald-400',
    setup: () => {
      designStore.nodes = []
      designStore.connections = []
      const bg = designStore.addNode('background', { x: 80, y: 140, label: 'Atmosphere', params: { type: 'horizontal', color: '#0f172a', color2: '#1e1b4b' } })
      const txt = designStore.addNode('textPlus', { x: 80, y: 300, label: 'MovieTitle', params: { text: 'ORIGIN', fontSize: 90, color: '#F1F5F9', tracking: 12, bold: true } })
      const cc = designStore.addNode('colorCorrector', { x: 300, y: 300, label: 'TealOrangeGrade', params: { lift: -0.05, gamma: 1.2, gain: 1.15, saturation: 1.35, tint: 18 } })
      const mrg = designStore.addNode('merge', { x: 520, y: 200, label: 'Composite' })
      const blur = designStore.addNode('blur', { x: 520, y: 360, label: 'AmbientBlur', params: { radius: 12 } })
      const out = designStore.addNode('mediaOut', { x: 740, y: 200, label: 'MediaOut' })

      if (bg && txt && cc && mrg && out) {
        designStore.addConnection(bg.id, 'out', mrg.id, 'bg')
        designStore.addConnection(txt.id, 'out', cc.id, 'in')
        designStore.addConnection(cc.id, 'out', mrg.id, 'fg')
        designStore.addConnection(mrg.id, 'out', out.id, 'in')
        designStore.setViewer1(txt.id)
        designStore.setViewer2(mrg.id)
      }
    }
  }
]

function renderLiveDemo() {
  const canvas = canvasDemoRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  demoTime += 0.02
  const W = canvas.width
  const H = canvas.height

  ctx.fillStyle = '#06060c'
  ctx.fillRect(0, 0, W, H)

  // Animated background grid lines
  ctx.strokeStyle = 'rgba(0, 212, 255, 0.08)'
  ctx.lineWidth = 1
  const gridSize = 30
  for (let x = 0; x < W; x += gridSize) {
    ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, H); ctx.stroke()
  }
  for (let y = 0; y < H; y += gridSize) {
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(W, y); ctx.stroke()
  }

  // Floating particle nodes in preview
  const cx = W / 2
  const cy = H / 2
  const pulse = Math.sin(demoTime * 2) * 8

  // Draw simulated Fusion node links in background
  const p1 = { x: cx - 140, y: cy - 30 }
  const p2 = { x: cx - 140, y: cy + 40 }
  const p3 = { x: cx, y: cy }
  const p4 = { x: cx + 130, y: cy }

  // Connections
  ctx.strokeStyle = 'rgba(0, 212, 255, 0.5)'
  ctx.lineWidth = 2
  ctx.setLineDash([4, 4])
  ctx.lineDashOffset = -demoTime * 20
  
  ctx.beginPath()
  ctx.moveTo(p1.x, p1.y); ctx.bezierCurveTo(p1.x + 60, p1.y, p3.x - 60, p3.y - 10, p3.x, p3.y - 10); ctx.stroke()
  ctx.beginPath()
  ctx.moveTo(p2.x, p2.y); ctx.bezierCurveTo(p2.x + 60, p2.y, p3.x - 60, p3.y + 10, p3.x, p3.y + 10); ctx.stroke()
  ctx.beginPath()
  ctx.moveTo(p3.x, p3.y); ctx.bezierCurveTo(p3.x + 50, p3.y, p4.x - 50, p4.y, p4.x, p4.y); ctx.stroke()
  ctx.setLineDash([])

  // Nodes Pills
  const nodes = [
    { p: p1, label: 'Background', col: '#EAB308', v: '1' },
    { p: p2, label: 'Text+', col: '#EC4899', v: '' },
    { p: p3, label: 'Merge1', col: '#10B981', v: '2' },
    { p: p4, label: 'MediaOut', col: '#F43F5E', v: '' }
  ]

  for (const n of nodes) {
    ctx.fillStyle = '#141422'
    ctx.strokeStyle = n.col
    ctx.lineWidth = 1.5
    ctx.shadowColor = n.col + '80'
    ctx.shadowBlur = 10
    ctx.beginPath()
    ctx.roundRect(n.p.x - 45, n.p.y - 16, 90, 32, 6)
    ctx.fill()
    ctx.stroke()
    ctx.shadowBlur = 0

    ctx.fillStyle = '#FFFFFF'
    ctx.font = 'bold 9px sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText(n.label, n.p.x - (n.v ? 6 : 0), n.p.y)

    if (n.v) {
      ctx.fillStyle = n.v === '1' ? '#00D4FF' : '#EC4899'
      ctx.beginPath()
      ctx.arc(n.p.x + 32, n.p.y, 5, 0, Math.PI * 2)
      ctx.fill()
      ctx.fillStyle = '#000000'
      ctx.font = 'bold 7px sans-serif'
      ctx.fillText(n.v, n.p.x + 32, n.p.y)
    }
  }

  demoAnimId = requestAnimationFrame(renderLiveDemo)
}

onMounted(() => {
  renderLiveDemo()
})

onUnmounted(() => {
  if (demoAnimId) cancelAnimationFrame(demoAnimId)
})

function loadPreset(preset) {
  preset.setup()
}
</script>

<template>
  <div class="flex-1 flex flex-col items-center justify-start p-4 sm:p-8 overflow-y-auto bg-[#07070C] select-none">
    <div class="max-w-4xl w-full">
      <!-- Fusion Hero Banner -->
      <div class="relative rounded-2xl bg-gradient-to-b from-[#161626] to-[#0D0D18] border border-border/70 p-5 sm:p-8 mb-6 shadow-2xl overflow-hidden">
        <!-- Background Canvas animation -->
        <canvas
          ref="canvasDemoRef"
          width="760"
          height="180"
          class="absolute inset-0 w-full h-full object-cover opacity-60 pointer-events-none"
        />

        <div class="relative z-10 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
          <div class="max-w-lg">
            <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-accent/15 border border-accent/30 text-accent text-[11px] font-semibold mb-3 tracking-wide">
              <Zap :size="12" class="animate-pulse" />
              DaVinci Resolve Fusion Workspace
            </div>
            <h1 class="text-xl sm:text-3xl font-extrabold text-white tracking-tight mb-2">
              Node-Based Compositing & Motion
            </h1>
            <p class="text-xs sm:text-sm text-text-secondary leading-relaxed">
              Connect generators, text, transforms, and color grades into visual graph pipelines. Use dual viewers <span class="font-mono text-cyan-400">[1]</span> & <span class="font-mono text-pink-400">[2]</span> for instant real-time feedback.
            </p>
          </div>

          <div class="flex flex-col gap-2 w-full sm:w-auto flex-shrink-0">
            <button
              class="flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-accent text-white font-bold text-xs shadow-lg shadow-accent/25 hover:bg-accent/90 transition-all hover:scale-[1.02] active:scale-[0.98]"
              @click="emit('start-blank')"
            >
              <Box :size="14" />
              <span>Blank Node Graph</span>
            </button>
            <button
              class="flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-[#1F1F30] border border-border/80 text-text-primary font-medium text-xs hover:bg-[#28283E] transition-all"
              @click="emit('start-template')"
            >
              <Sparkles :size="14" class="text-pink-400" />
              <span>Browse All Templates</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Quick Interactive Starter Presets -->
      <div class="mb-6">
        <div class="flex items-center justify-between mb-3 px-1">
          <div class="flex items-center gap-2">
            <Layers :size="14" class="text-accent" />
            <span class="text-xs font-bold uppercase tracking-wider text-text-primary">1-Click Fusion Compositions</span>
          </div>
          <span class="text-[11px] text-text-secondary font-medium">Click to load and animate immediately</span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div
            v-for="preset in DEMO_PRESETS"
            :key="preset.id"
            class="group relative flex flex-col p-4 rounded-xl border bg-gradient-to-br transition-all cursor-pointer hover:scale-[1.02] hover:shadow-xl"
            :class="preset.accent"
            @click="loadPreset(preset)"
          >
            <div class="flex items-center justify-between mb-2.5">
              <div class="w-8 h-8 rounded-lg bg-bg/80 border border-border/60 flex items-center justify-center">
                <component :is="preset.icon" :size="16" :style="{ color: preset.color }" />
              </div>
              <span class="text-[9px] font-bold uppercase px-2 py-0.5 rounded-full bg-black/40 text-text-secondary border border-border/50">
                {{ preset.badge }}
              </span>
            </div>

            <div class="text-[10px] font-bold tracking-wider uppercase mb-0.5" :style="{ color: preset.color }">
              {{ preset.category }}
            </div>
            <div class="text-sm font-bold text-white mb-1.5">{{ preset.title }}</div>
            <p class="text-[11px] text-text-secondary/90 leading-relaxed mb-4 flex-1">
              {{ preset.desc }}
            </p>

            <div class="flex items-center gap-1.5 text-xs font-semibold" :style="{ color: preset.color }">
              <span>Open in Flow Graph</span>
              <ArrowRight :size="13" class="group-hover:translate-x-1 transition-transform" />
            </div>
          </div>
        </div>
      </div>

      <!-- Fusion Quick Guide Bar -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5 text-[11px] text-text-secondary">
        <div class="flex items-center gap-2.5 p-3 rounded-xl bg-[#11111B] border border-border/50">
          <div class="w-6 h-6 rounded bg-yellow-500/15 border border-yellow-500/30 flex items-center justify-center text-yellow-400 font-bold text-[10px]">
            Bg
          </div>
          <div>
            <span class="text-text-primary font-semibold block">Yellow Socket</span>
            <span>Background stream</span>
          </div>
        </div>

        <div class="flex items-center gap-2.5 p-3 rounded-xl bg-[#11111B] border border-border/50">
          <div class="w-6 h-6 rounded bg-green-500/15 border border-green-500/30 flex items-center justify-center text-green-400 font-bold text-[10px]">
            Fg
          </div>
          <div>
            <span class="text-text-primary font-semibold block">Green Socket</span>
            <span>Foreground layer</span>
          </div>
        </div>

        <div class="flex items-center gap-2.5 p-3 rounded-xl bg-[#11111B] border border-border/50">
          <div class="w-6 h-6 rounded bg-blue-500/15 border border-blue-500/30 flex items-center justify-center text-blue-400 font-bold text-[10px]">
            Msk
          </div>
          <div>
            <span class="text-text-primary font-semibold block">Blue Socket</span>
            <span>Effect mask input</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>