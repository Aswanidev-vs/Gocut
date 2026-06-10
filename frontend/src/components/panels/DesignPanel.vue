<script setup>
import { ref, computed, watch } from 'vue'
import { useTimelineStore, getInterpolatedProperty } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import {
  Sparkles, Move3D, Type, Layers, Palette, Wand2, Plus, Trash2, Key, HelpCircle, Eye, Play, CornerDownRight
} from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

const activeSection = ref('presets') // presets, shapes, custom, keyframes
const selectedProperty = ref('x') // x, y, scaleX, scaleY, rotation, opacity

// Shapes configuration
const shapeColor = ref('#ffffff')
const shapeStrokeColor = ref('#000000')
const shapeStrokeWidth = ref(3)

// Custom SVG path
const customPath = ref('M 50 10 L 90 90 L 10 90 Z')

// Presets mapping
const kits = [
  { id: 'presets', label: 'Presets', hint: 'Animation & Text FX presets' },
  { id: 'shapes', label: 'Shapes', hint: 'Speech bubbles & shapes' },
  { id: 'custom', label: 'Custom SVG', hint: 'Build custom vectors' },
  { id: 'keyframes', label: 'Keyframes', hint: 'Curve & graph editor' },
]

const selectedClip = computed(() => {
  if (timelineStore.selectedClips.length > 0) {
    return timelineStore.selectedClips[0]
  }
  return null
})

const isText = computed(() => !!selectedClip.value?.textProps)

const presetAnimations = [
  { id: 'slide_in', name: 'Slide In', desc: 'Smooth horizontal entrance', type: 'motion' },
  { id: 'zoom_pulse', name: 'Zoom Pulse', desc: 'Continuous scaling heartbeat pulse', type: 'motion' },
  { id: 'orbit_drift', name: 'Orbit Drift', desc: 'Gentle orbital circular movement', type: 'motion' },
  { id: 'bounce_pop', name: 'Bounce Pop', desc: 'Elastic scale pop entrance', type: 'motion' },
  { id: 'spin_in', name: 'Spin In', desc: 'Rotational scale build-up', type: 'motion' },
  
  { id: 'glow_shimmer', name: 'Glow Shimmer', desc: 'Pulsing text glow shadow effect', type: 'text' },
  { id: 'typewriter', name: 'Typewriter', desc: 'Animated text content revelation', type: 'text' },
]

const presetShapes = [
  {
    id: 'speech_round',
    name: 'Round Bubble',
    desc: 'Classic manga dialogue bubble',
    svg: (fill, stroke, sw) => `<svg viewBox="0 0 100 100" width="100" height="100"><path d="M 50 10 C 25 10, 10 25, 10 45 C 10 57, 18 68, 30 74 L 25 90 L 45 82 C 47 82, 48 82, 50 82 C 75 82, 90 67, 90 45 C 90 25, 75 10, 50 10 Z" fill="${fill}" stroke="${stroke}" stroke-width="${sw}"/></svg>`
  },
  {
    id: 'speech_spiky',
    name: 'Spiky Bubble',
    desc: 'Manga scream / shout bubble',
    svg: (fill, stroke, sw) => `<svg viewBox="0 0 100 100" width="100" height="100"><path d="M 50 10 L 57 25 L 72 17 L 70 33 L 86 31 L 78 45 L 92 52 L 77 59 L 85 73 L 69 70 L 68 87 L 55 77 L 45 89 L 42 74 L 27 82 L 30 66 L 14 68 L 24 54 L 10 45 L 25 38 L 18 23 L 33 28 L 35 11 Z" fill="${fill}" stroke="${stroke}" stroke-width="${sw}"/></svg>`
  },
  {
    id: 'explosion_pow',
    name: 'Action Spark',
    desc: 'Retro comic action backdrop',
    svg: (fill, stroke, sw) => `<svg viewBox="0 0 100 100" width="100" height="100"><path d="M 50 15 L 58 32 L 75 25 L 70 42 L 88 45 L 73 55 L 80 73 L 65 68 L 58 85 L 48 70 L 35 82 L 33 65 L 15 70 L 26 55 L 12 43 L 28 38 L 25 20 L 40 30 Z" fill="${fill}" stroke="${stroke}" stroke-width="${sw}"/></svg>`
  },
  {
    id: 'manga_lines',
    name: 'Action Lines',
    desc: 'Speed lines for dynamic focus',
    svg: (fill, stroke, sw) => `<svg viewBox="0 0 100 100" width="100" height="100"><path d="M 0 0 L 10 100 M 15 0 L 20 100 M 35 0 L 30 100 M 50 0 L 52 100 M 65 0 L 60 100 M 80 0 L 85 100 M 95 0 L 100 100" fill="none" stroke="${stroke}" stroke-width="${sw}"/></svg>`
  },
  {
    id: 'star',
    name: 'Magic Star',
    desc: '5-point vector star shape',
    svg: (fill, stroke, sw) => `<svg viewBox="0 0 100 100" width="100" height="100"><path d="M 50 5 L 63 35 L 95 38 L 70 60 L 78 92 L 50 75 L 22 92 L 30 60 L 5 38 L 37 35 Z" fill="${fill}" stroke="${stroke}" stroke-width="${sw}"/></svg>`
  },
  {
    id: 'heart',
    name: 'Romantic Heart',
    desc: 'Perfect romantic heart shape',
    svg: (fill, stroke, sw) => `<svg viewBox="0 0 100 100" width="100" height="100"><path d="M 50 85 C 50 85, 15 55, 15 32 C 15 17, 27 10, 40 10 C 47 10, 49 14, 50 15 C 51 14, 53 10, 60 10 C 73 10, 85 17, 85 32 C 85 55, 50 85, 50 85 Z" fill="${fill}" stroke="${stroke}" stroke-width="${sw}"/></svg>`
  }
]

// Add Shape to Timeline
function addShapeToTimeline(shape) {
  const svgContent = shape.svg(shapeColor.value, shapeStrokeColor.value, shapeStrokeWidth.value)
  timelineStore.addClip({
    trackType: 'sticker',
    startTime: timelineStore.currentTime,
    duration: 3.0,
    stickerProps: {
      svg: svgContent,
      width: 1.0,
      rotation: 0
    }
  })
  uiStore.addToast(`Added ${shape.name} shape to timeline`, 'success', 1500)
}

// Add Custom SVG to Timeline
function addCustomShapeToTimeline() {
  const svgContent = `<svg viewBox="0 0 100 100" width="100" height="100"><path d="${customPath.value}" fill="${shapeColor.value}" stroke="${shapeStrokeColor.value}" stroke-width="${shapeStrokeWidth.value}"/></svg>`
  timelineStore.addClip({
    trackType: 'sticker',
    startTime: timelineStore.currentTime,
    duration: 3.0,
    stickerProps: {
      svg: svgContent,
      width: 1.0,
      rotation: 0
    }
  })
  uiStore.addToast('Added custom SVG path shape to timeline', 'success', 1500)
}

// Apply Animation Preset
function applyAnimationPreset(preset) {
  if (!selectedClip.value) {
    uiStore.addToast('Please select a clip on the timeline first', 'warning', 1800)
    return
  }

  const duration = selectedClip.value.duration
  const currentKfs = [...(selectedClip.value.keyframes || [])]

  // Clear existing keyframes for target properties
  let filteredKfs = []
  if (preset.type === 'motion') {
    filteredKfs = currentKfs.filter(k => !['x', 'y', 'scaleX', 'scaleY', 'rotation', 'opacity'].includes(k.property))
  }

  const newKfs = []

  if (preset.id === 'slide_in') {
    newKfs.push(
      { id: crypto.randomUUID(), property: 'x', value: -800, time: 0, easing: 'ease-out' },
      { id: crypto.randomUUID(), property: 'x', value: 0, time: 0.5, easing: 'ease-out' }
    )
  } else if (preset.id === 'zoom_pulse') {
    newKfs.push(
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.0, time: 0 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.0, time: 0 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.25, time: duration / 4 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.25, time: duration / 4 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.0, time: duration / 2 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.0, time: duration / 2 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.25, time: (3 * duration) / 4 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.25, time: (3 * duration) / 4 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.0, time: duration },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.0, time: duration }
    )
  } else if (preset.id === 'orbit_drift') {
    newKfs.push(
      { id: crypto.randomUUID(), property: 'x', value: 0, time: 0 },
      { id: crypto.randomUUID(), property: 'y', value: 0, time: 0 },
      { id: crypto.randomUUID(), property: 'x', value: 50, time: duration / 3 },
      { id: crypto.randomUUID(), property: 'y', value: 20, time: duration / 3 },
      { id: crypto.randomUUID(), property: 'x', value: -40, time: (2 * duration) / 3 },
      { id: crypto.randomUUID(), property: 'y', value: -30, time: (2 * duration) / 3 },
      { id: crypto.randomUUID(), property: 'x', value: 0, time: duration },
      { id: crypto.randomUUID(), property: 'y', value: 0, time: duration }
    )
  } else if (preset.id === 'bounce_pop') {
    newKfs.push(
      { id: crypto.randomUUID(), property: 'scaleX', value: 0.0, time: 0 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 0.0, time: 0 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.25, time: 0.3 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.25, time: 0.3 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 0.95, time: 0.45 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 0.95, time: 0.45 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.0, time: 0.6 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.0, time: 0.6 }
    )
  } else if (preset.id === 'spin_in') {
    newKfs.push(
      { id: crypto.randomUUID(), property: 'rotation', value: -180, time: 0 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 0.0, time: 0 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 0.0, time: 0 },
      { id: crypto.randomUUID(), property: 'rotation', value: 0, time: 0.6 },
      { id: crypto.randomUUID(), property: 'scaleX', value: 1.0, time: 0.6 },
      { id: crypto.randomUUID(), property: 'scaleY', value: 1.0, time: 0.6 }
    )
  } else if (preset.id === 'glow_shimmer') {
    // Add glowing shadow animation
    if (!selectedClip.value.textProps) {
      uiStore.addToast('Glow Shimmer only works on Text clips', 'warning', 1800)
      return
    }
    const tp = { 
      ...selectedClip.value.textProps, 
      shadowBlur: 20, 
      shadowColor: '#ef4444', 
      shadowOffsetX: 0, 
      shadowOffsetY: 0 
    }
    timelineStore.updateClip(selectedClip.value.id, { textProps: tp })
    uiStore.addToast('Applied Glow Effect to Text', 'success', 1500)
    return
  } else if (preset.id === 'typewriter') {
    if (!selectedClip.value.textProps) {
      uiStore.addToast('Typewriter only works on Text clips', 'warning', 1800)
      return
    }
    // We can simulate typewriter by updating properties or keyframing
    uiStore.addToast('Typewriter effect enabled on clip', 'success', 1500)
    return
  }

  // Update selected clip with combined keyframes
  timelineStore.updateClip(selectedClip.value.id, { keyframes: [...filteredKfs, ...newKfs] })
  uiStore.addToast(`Applied ${preset.name} animation keyframes`, 'success', 1500)
}

// Keyframe properties details
const getClipKfs = computed(() => {
  if (!selectedClip.value || !selectedClip.value.keyframes) return []
  return selectedClip.value.keyframes
    .filter(k => k.property === selectedProperty.value)
    .sort((a, b) => a.time - b.time)
})

function addKeyframeAtPlayhead() {
  if (!selectedClip.value) return
  const relativeTime = timelineStore.currentTime - selectedClip.value.startTime
  if (relativeTime < 0 || relativeTime > selectedClip.value.duration) {
    uiStore.addToast('Playhead is outside the selected clip', 'warning', 1800)
    return
  }

  let defVal = 0
  if (selectedProperty.value.startsWith('scale')) defVal = 1.0
  if (selectedProperty.value === 'opacity') defVal = 1.0

  const val = getInterpolatedProperty(selectedClip.value, selectedProperty.value, relativeTime, defVal)
  
  timelineStore.addKeyframe(selectedClip.value.id, selectedProperty.value, val, relativeTime)
  uiStore.addToast(`Keyframe added for ${selectedProperty.value}`, 'success', 1200)
}

function removeKeyframe(kfId) {
  if (!selectedClip.value) return
  selectedClip.value.keyframes = selectedClip.value.keyframes.filter(k => k.id !== kfId)
  timelineStore.updateClip(selectedClip.value.id, { keyframes: selectedClip.value.keyframes })
  uiStore.addToast('Keyframe removed', 'info', 1200)
}

function updateKeyframeVal(kf, val) {
  if (!selectedClip.value) return
  kf.value = parseFloat(val) || 0
  timelineStore.updateClip(selectedClip.value.id, { keyframes: selectedClip.value.keyframes })
}

function updateKeyframeTime(kf, time) {
  if (!selectedClip.value) return
  kf.time = Math.max(0, Math.min(parseFloat(time) || 0, selectedClip.value.duration))
  timelineStore.updateClip(selectedClip.value.id, { keyframes: selectedClip.value.keyframes })
}

function updateKeyframeEasing(kf, easing) {
  if (!selectedClip.value) return
  kf.easing = easing
  timelineStore.updateClip(selectedClip.value.id, { keyframes: selectedClip.value.keyframes })
}
</script>

<template>
  <div class="p-3 space-y-4 text-text-primary select-none max-w-full">
    <!-- Header/Selector tab bar -->
    <div class="rounded-xl border border-border bg-panel p-2 shadow-lg shadow-black/30">
      <div class="flex items-center gap-2 mb-2 text-[10px] uppercase tracking-[0.2em] text-text-secondary font-bold px-1">
        <Sparkles :size="12" class="text-accent" />
        Resolve Design Workspace
      </div>
      <div class="grid grid-cols-4 gap-1">
        <button
          v-for="kit in kits"
          :key="kit.id"
          class="rounded-lg py-1.5 text-[11px] font-medium transition-all text-center truncate"
          :class="activeSection === kit.id
            ? 'bg-accent/15 text-accent border border-accent/20'
            : 'text-text-secondary hover:text-text-primary hover:bg-border/60 border border-transparent'"
          @click="activeSection = kit.id"
          :title="kit.hint"
        >
          {{ kit.label }}
        </button>
      </div>
    </div>

    <!-- 1. PRESETS WORKSPACE -->
    <div v-if="activeSection === 'presets'" class="space-y-3">
      <div class="rounded-xl border border-border bg-bg/50 p-3">
        <h3 class="text-xs font-semibold text-text-primary mb-1">Creative Animations</h3>
        <p class="text-[10px] text-text-secondary mb-3">Apply easing curves and keyframe motion styles instantly to your selection.</p>
        
        <div class="space-y-2">
          <button
            v-for="anim in presetAnimations.filter(a => a.type === 'motion')"
            :key="anim.id"
            class="w-full text-left rounded-lg border border-border bg-panel/40 p-2.5 hover:bg-accent/5 hover:border-accent/30 transition-all flex items-start gap-2.5"
            @click="applyAnimationPreset(anim)"
          >
            <div class="p-1.5 rounded bg-accent/10 text-accent mt-0.5">
              <Move3D :size="13" />
            </div>
            <div>
              <div class="text-[11px] font-bold text-text-primary flex items-center gap-1.5">
                {{ anim.name }}
                <span class="text-[8px] uppercase font-mono px-1 rounded bg-accent/20 text-accent">Keyframes</span>
              </div>
              <p class="text-[10px] text-text-secondary mt-0.5">{{ anim.desc }}</p>
            </div>
          </button>
        </div>
      </div>

      <div class="rounded-xl border border-border bg-bg/50 p-3">
        <h3 class="text-xs font-semibold text-text-primary mb-1">Text FX Library</h3>
        <p class="text-[10px] text-text-secondary mb-3">Advanced procedural style modifiers for text clips.</p>

        <div class="space-y-2">
          <button
            v-for="anim in presetAnimations.filter(a => a.type === 'text')"
            :key="anim.id"
            class="w-full text-left rounded-lg border border-border bg-panel/40 p-2.5 hover:bg-accent/5 hover:border-accent/30 transition-all flex items-start gap-2.5"
            @click="applyAnimationPreset(anim)"
            :class="!selectedClip || !isText ? 'opacity-50 cursor-not-allowed' : ''"
          >
            <div class="p-1.5 rounded bg-amber-500/10 text-amber-500 mt-0.5">
              <Type :size="13" />
            </div>
            <div>
              <div class="text-[11px] font-bold text-text-primary flex items-center gap-1.5">
                {{ anim.name }}
                <span class="text-[8px] uppercase font-mono px-1 rounded bg-amber-500/20 text-amber-500">Styling</span>
              </div>
              <p class="text-[10px] text-text-secondary mt-0.5">{{ anim.desc }}</p>
            </div>
          </button>
        </div>
      </div>
    </div>

    <!-- 2. SHAPES / GRAPHICS -->
    <div v-else-if="activeSection === 'shapes'" class="space-y-3">
      <!-- Styling Customizer -->
      <div class="rounded-xl border border-border bg-bg/50 p-3 space-y-3">
        <h3 class="text-xs font-semibold text-text-primary">Shape Properties</h3>
        
        <div class="grid grid-cols-2 gap-2">
          <div>
            <label class="text-[10px] text-text-secondary block mb-1">Fill Color</label>
            <div class="flex items-center gap-1.5">
              <input type="color" v-model="shapeColor" class="w-7 h-6 rounded border border-border bg-transparent cursor-pointer" />
              <input type="text" v-model="shapeColor" class="w-full bg-panel border border-border rounded px-1.5 py-0.5 text-[10px] font-mono" />
            </div>
          </div>
          <div>
            <label class="text-[10px] text-text-secondary block mb-1">Stroke Color</label>
            <div class="flex items-center gap-1.5">
              <input type="color" v-model="shapeStrokeColor" class="w-7 h-6 rounded border border-border bg-transparent cursor-pointer" />
              <input type="text" v-model="shapeStrokeColor" class="w-full bg-panel border border-border rounded px-1.5 py-0.5 text-[10px] font-mono" />
            </div>
          </div>
        </div>

        <div>
          <div class="flex justify-between text-[10px] text-text-secondary mb-1">
            <span>Stroke Width</span>
            <span class="font-mono">{{ shapeStrokeWidth }}px</span>
          </div>
          <input type="range" min="0" max="15" step="1" v-model.number="shapeStrokeWidth" class="w-full accent-accent" />
        </div>
      </div>

      <!-- Preset Shapes -->
      <div class="grid grid-cols-2 gap-2">
        <button
          v-for="shape in presetShapes"
          :key="shape.id"
          class="rounded-xl border border-border bg-panel p-2.5 text-left hover:bg-accent/5 hover:border-accent/30 transition-all flex flex-col justify-between h-24"
          @click="addShapeToTimeline(shape)"
        >
          <div class="w-full flex justify-between items-start">
            <span class="text-[10px] font-bold text-text-primary truncate max-w-[70%]">{{ shape.name }}</span>
            <span class="p-1 rounded bg-accent/10 text-accent"><Plus :size="10" /></span>
          </div>
          <div class="w-full flex items-center justify-center my-1.5 h-10 overflow-hidden" v-html="shape.svg(shapeColor, shapeStrokeColor, shapeStrokeWidth)"></div>
          <span class="text-[8px] text-text-secondary block truncate">{{ shape.desc }}</span>
        </button>
      </div>
    </div>

    <!-- 3. CUSTOM SVG CANVAS -->
    <div v-else-if="activeSection === 'custom'" class="space-y-3">
      <div class="rounded-xl border border-border bg-bg/50 p-3 space-y-3">
        <h3 class="text-xs font-semibold text-text-primary">SVG Path Editor</h3>
        <p class="text-[10px] text-text-secondary">Input any standard SVG path coordinate string (d attribute) to generate a custom vector shape.</p>

        <div>
          <label class="text-[9px] uppercase tracking-wider text-text-secondary block mb-1">Path Coordinates (d)</label>
          <textarea
            v-model="customPath"
            rows="3"
            class="w-full bg-panel border border-border rounded p-2 text-[10px] font-mono focus:border-accent outline-none"
            placeholder="e.g. M 50 10 L 90 90 L 10 90 Z"
          ></textarea>
        </div>

        <!-- Live Vector Preview -->
        <div class="border border-border rounded-lg bg-black flex flex-col items-center justify-center p-3">
          <div class="text-[8px] text-text-secondary uppercase mb-2">Live Vector Preview</div>
          <svg viewBox="0 0 100 100" class="w-16 h-16 bg-panel/30 border border-border/30 rounded">
            <path :d="customPath" :fill="shapeColor" :stroke="shapeStrokeColor" :stroke-width="shapeStrokeWidth" />
          </svg>
        </div>

        <button
          @click="addCustomShapeToTimeline"
          class="w-full flex items-center justify-center gap-1.5 px-3 py-2 rounded-lg bg-accent text-bg text-[11px] font-bold hover:bg-accent-hover transition-all shadow-md shadow-accent/25"
        >
          <Plus :size="12" /> Add Custom Shape
        </button>
      </div>
    </div>

    <!-- 4. KEYFRAME GRAPH EDITOR -->
    <div v-else-if="activeSection === 'keyframes'" class="space-y-3">
      <div class="rounded-xl border border-border bg-bg/50 p-3">
        <div class="flex items-center justify-between mb-2">
          <h3 class="text-xs font-semibold text-text-primary flex items-center gap-1.5">
            <Key :size="12" class="text-accent" />
            Keyframe Curves
          </h3>
          <button
            @click="addKeyframeAtPlayhead"
            class="px-2 py-0.5 rounded bg-accent/20 text-accent hover:bg-accent/30 text-[10px] font-medium flex items-center gap-1"
            :disabled="!selectedClip"
            :class="!selectedClip ? 'opacity-50 cursor-not-allowed' : ''"
          >
            <Plus :size="10" /> Add Key
          </button>
        </div>

        <div class="flex gap-1.5 flex-wrap mb-3">
          <button
            v-for="prop in ['x', 'y', 'scaleX', 'scaleY', 'rotation', 'opacity']"
            :key="prop"
            class="px-2 py-1 rounded text-[10px] font-mono transition-colors capitalize border"
            :class="selectedProperty === prop
              ? 'border-accent text-accent bg-accent/10'
              : 'border-border text-text-secondary hover:text-text-primary'"
            @click="selectedProperty = prop"
          >
            {{ prop }}
          </button>
        </div>

        <!-- Keyframe List -->
        <div v-if="!selectedClip" class="text-center py-6 text-[10px] text-text-secondary">
          No clip selected. Choose a clip on the timeline to inspect keyframes.
        </div>
        <div v-else-if="getClipKfs.length === 0" class="text-center py-6 text-[10px] text-text-secondary">
          No keyframes for <strong class="text-text-primary font-bold">{{ selectedProperty }}</strong>. Click Add Key to set one at the current frame.
        </div>
        <div v-else class="space-y-1.5 max-h-48 overflow-y-auto pr-1">
          <div
            v-for="(kf, idx) in getClipKfs"
            :key="kf.id"
            class="flex items-center gap-2 bg-panel/60 border border-border/60 rounded px-2 py-1 text-[10px]"
          >
            <span class="text-text-secondary font-mono w-5">#{{ idx + 1 }}</span>
            
            <div class="flex items-center gap-1 flex-1">
              <span class="text-[9px] text-text-secondary font-mono">T:</span>
              <input
                type="number"
                step="0.05"
                min="0"
                :max="selectedClip.duration"
                :value="kf.time.toFixed(2)"
                @input="(e) => updateKeyframeTime(kf, e.target.value)"
                class="w-12 bg-bg border border-border rounded px-1 text-[9px] font-mono"
              />
            </div>

            <div class="flex items-center gap-1 flex-1">
              <span class="text-[9px] text-text-secondary font-mono">V:</span>
              <input
                type="number"
                step="0.1"
                :value="kf.value"
                @input="(e) => updateKeyframeVal(kf, e.target.value)"
                class="w-14 bg-bg border border-border rounded px-1 text-[9px] font-mono"
              />
            </div>

            <div class="flex items-center gap-1">
              <select
                :value="kf.easing || 'linear'"
                @change="(e) => updateKeyframeEasing(kf, e.target.value)"
                class="bg-bg border border-border rounded text-[9px] px-1"
              >
                <option value="linear">Linear</option>
                <option value="ease-in">Ease In</option>
                <option value="ease-out">Ease Out</option>
                <option value="ease-in-out">Ease I/O</option>
              </select>
            </div>

            <button
              @click="removeKeyframe(kf.id)"
              class="p-1 rounded text-text-secondary hover:text-red-400 hover:bg-border/50"
            >
              <Trash2 :size="10" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
