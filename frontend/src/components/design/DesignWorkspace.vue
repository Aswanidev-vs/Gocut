<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { useUiStore } from '../../stores/uiStore'
import { RECIPES, applyRecipe } from '../../lib/designRecipes'
import {
  Play, Pause, RotateCcw, ZoomIn, ZoomOut, Maximize2, Magnet, Activity,
} from 'lucide-vue-next'
import DesignRail from './DesignRail.vue'
import DesignOnboarding from './DesignOnboarding.vue'
import NodeGraph from './NodeGraph.vue'
import NodeInspector from './NodeInspector.vue'
import AnimationCurves from './AnimationCurves.vue'
import CompositingCanvas from './CompositingCanvas.vue'
import { useDesignHotkeys } from '../../composables/useDesignHotkeys'

useDesignHotkeys()

const designStore = useDesignStore()
const uiStore = useUiStore()

// Resizable panel widths & heights (unchanged bounds from the Fusion layout).
const MIN_LIBRARY_WIDTH = 180
const MIN_INSPECTOR_WIDTH = 240
const MAX_LIBRARY_WIDTH = 420
const MAX_INSPECTOR_WIDTH = 460

const libraryWidth = ref(240)
const inspectorWidth = ref(300)
const showCurves = ref(false)
const curvesHeight = ref(180)
const viewerHeight = ref(340)

const isPlaying = ref(false)
let playInterval = null
const playheadTime = ref(0)

// Simple leads: one viewer, moves on the left, parameters on the right. Flow is
// the full node graph for people who want to wire it by hand.
const MODES = [
  { id: 'simple', label: 'Simple', hint: 'One viewer, moves and parameters' },
  { id: 'flow', label: 'Flow', hint: 'Node graph, viewers and spline editor' },
]
const mode = ref('simple')

const VIEW_MODES = [
  { id: 'single', label: 'Output' },
  { id: 'dual', label: 'A/B' },
  { id: 'graph', label: 'Graph only' },
]
const viewMode = ref('single')

const showOnboarding = computed(() => designStore.nodes.length === 0)
const fps = computed(() => designStore.composition.fps || 30)
const duration = computed(() => designStore.composition.duration || 5)
const frame = computed(() => Math.round(playheadTime.value * fps.value))

function timecode(seconds) {
  const f = Math.max(0, Math.round(seconds * fps.value))
  const total = Math.floor(f / fps.value)
  const mm = String(Math.floor(total / 60)).padStart(2, '0')
  const ss = String(total % 60).padStart(2, '0')
  const ff = String(f % fps.value).padStart(2, '0')
  return `${mm}:${ss}:${ff}`
}
const timecodeNow = computed(() => timecode(playheadTime.value))
const timecodeEnd = computed(() => timecode(duration.value))

function setMode(next) {
  mode.value = next
  if (next === 'simple') showCurves.value = false
}

// ---- Playback ----
function togglePlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) {
    playInterval = setInterval(() => {
      playheadTime.value += 1 / fps.value
      if (playheadTime.value > duration.value) playheadTime.value = 0
    }, 1000 / fps.value)
  } else {
    clearInterval(playInterval)
  }
}

function resetPlayhead() {
  playheadTime.value = 0
  isPlaying.value = false
  clearInterval(playInterval)
}

function onStepPlayhead(e) {
  const step = (e.detail || 1) / fps.value
  playheadTime.value = Math.max(0, Math.min(duration.value, playheadTime.value + step))
}

// ---- Panel resizing ----
let activeDrag = null

function startDrag(e, panel) {
  activeDrag = panel
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  e.preventDefault()
}

function onDrag(e) {
  if (!activeDrag) return
  if (activeDrag === 'library') {
    libraryWidth.value = Math.max(MIN_LIBRARY_WIDTH, Math.min(e.clientX - 260, MAX_LIBRARY_WIDTH))
  } else if (activeDrag === 'inspector') {
    inspectorWidth.value = Math.max(MIN_INSPECTOR_WIDTH, Math.min(window.innerWidth - e.clientX, MAX_INSPECTOR_WIDTH))
  } else if (activeDrag === 'viewer') {
    viewerHeight.value = Math.max(180, Math.min(e.clientY - 80, window.innerHeight - 250))
  } else if (activeDrag === 'curves') {
    curvesHeight.value = Math.max(100, Math.min(window.innerHeight - e.clientY, window.innerHeight - 150))
  }
}

function stopDrag() {
  activeDrag = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

// ---- Starting a composition ----
function startFromRecipe(id) {
  const recipe = RECIPES.find(r => r.id === id)
  if (!recipe) return
  applyRecipe(designStore, recipe)
  setMode('simple')
  uiStore.addToast(`${recipe.name} added`, 'success', 1600)
}

function startBlank() {
  designStore.clearSelection()
  const bg = designStore.addNode('background', { x: 80, y: 110, label: 'Backdrop' })
  designStore.clearSelection()
  const text = designStore.addNode('text', { x: 80, y: 230, label: 'Text', params: { text: 'Hello', fontSize: 96 } })
  designStore.clearSelection()
  const merge = designStore.addNode('merge', { x: 400, y: 170, label: 'Over' })
  designStore.clearSelection()
  const out = designStore.addNode('output', { x: 660, y: 170, label: 'Output' })
  if (bg && text && merge && out) {
    designStore.addConnection(bg.id, 'out', merge.id, 'bg')
    designStore.addConnection(text.id, 'out', merge.id, 'fg')
    designStore.addConnection(merge.id, 'out', out.id, 'in')
  }
  setMode('flow')
}

onMounted(() => {
  window.addEventListener('design:togglePlay', togglePlay)
  window.addEventListener('design:stepPlayhead', onStepPlayhead)
})

onUnmounted(() => {
  clearInterval(playInterval)
  window.removeEventListener('design:togglePlay', togglePlay)
  window.removeEventListener('design:stepPlayhead', onStepPlayhead)
})

defineExpose({ libraryWidth, inspectorWidth, curvesHeight, showCurves })
</script>

<template>
  <div class="flex flex-col h-full bg-cr-void overflow-hidden select-none">
    <!-- Control rail: identity, mode, tools, transport -->
    <div class="h-9 flex-shrink-0 bg-cr-raise border-b border-cr-line flex items-center px-2 gap-3">
      <div class="flex items-baseline gap-2 flex-shrink-0">
        <span class="text-[10px] uppercase tracking-[0.18em] text-ink">Design</span>
        <span class="text-[9px] font-jetbrains-mono text-ink-faint">{{ designStore.composition.name }}</span>
      </div>

      <div class="flex items-stretch border border-cr-line flex-shrink-0">
        <button
          v-for="m in MODES"
          :key="m.id"
          :title="m.hint"
          class="px-2.5 py-1 text-[10px] uppercase tracking-[0.14em] transition-colors"
          :class="mode === m.id ? 'bg-accent/10 text-accent' : 'text-ink-dim hover:text-ink hover:bg-white/5'"
          @click="setMode(m.id)"
        >
          {{ m.label }}
        </button>
      </div>

      <div class="flex-1" />

      <!-- Graph tools only matter in Flow -->
      <div v-if="mode === 'flow'" class="flex items-center gap-0.5 flex-shrink-0">
        <button
          class="p-1.5 transition-colors"
          :class="designStore.snapEnabled ? 'text-accent bg-accent/10' : 'text-ink-dim hover:text-ink hover:bg-white/5'"
          title="Snap to grid"
          @click="designStore.snapEnabled = !designStore.snapEnabled"
        >
          <Magnet :size="12" />
        </button>
        <button class="p-1.5 text-ink-dim hover:text-ink hover:bg-white/5 transition-colors" title="Zoom to fit" @click="designStore.zoomFit()">
          <Maximize2 :size="12" />
        </button>
        <button class="p-1.5 text-ink-dim hover:text-ink hover:bg-white/5 transition-colors" title="Zoom in" @click="designStore.zoomIn()">
          <ZoomIn :size="12" />
        </button>
        <button class="p-1.5 text-ink-dim hover:text-ink hover:bg-white/5 transition-colors" title="Zoom out" @click="designStore.zoomOut()">
          <ZoomOut :size="12" />
        </button>
        <button
          class="p-1.5 transition-colors"
          :class="showCurves ? 'text-signal bg-signal/10' : 'text-ink-dim hover:text-ink hover:bg-white/5'"
          title="Spline editor"
          @click="showCurves = !showCurves"
        >
          <Activity :size="12" />
        </button>
      </div>

      <!-- Transport: time reads amber so it never competes with selection -->
      <div class="flex items-center gap-2 flex-shrink-0 border-l border-cr-line pl-3">
        <button class="p-1 text-ink-dim hover:text-ink transition-colors" title="Back to start" @click="resetPlayhead">
          <RotateCcw :size="12" />
        </button>
        <button
          class="p-1 transition-colors"
          :class="isPlaying ? 'text-signal' : 'text-ink-dim hover:text-ink'"
          :title="isPlaying ? 'Pause (Space)' : 'Play (Space)'"
          @click="togglePlay"
        >
          <Pause v-if="isPlaying" :size="13" />
          <Play v-else :size="13" />
        </button>
        <span class="text-[10px] font-jetbrains-mono tabular-nums text-signal">{{ timecodeNow }}</span>
        <span class="text-[10px] font-jetbrains-mono tabular-nums text-ink-faint">/ {{ timecodeEnd }}</span>
      </div>
    </div>

    <div class="flex-1 flex min-h-0">
      <!-- Moves / Looks / Nodes -->
      <div class="bg-cr-panel flex flex-col flex-shrink-0 overflow-hidden" :style="{ width: libraryWidth + 'px' }">
        <DesignRail />
      </div>
      <div class="w-px cursor-col-resize hover:bg-accent/60 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'library')" />

      <!-- Stage -->
      <div class="flex-1 flex flex-col min-w-0 min-h-0 bg-cr-void">
        <DesignOnboarding
          v-if="showOnboarding"
          @pick="startFromRecipe"
          @blank="startBlank"
        />
        <template v-else>
          <!-- Flow: viewer selection strip -->
          <div v-if="mode === 'flow'" class="h-7 flex-shrink-0 bg-cr-panel border-b border-cr-line flex items-center gap-2 px-2">
            <span class="text-[9px] uppercase tracking-[0.14em] text-ink-faint">View</span>
            <button
              v-for="v in VIEW_MODES"
              :key="v.id"
              class="px-2 py-0.5 text-[10px] border transition-colors"
              :class="viewMode === v.id ? 'border-accent/40 bg-accent/10 text-accent' : 'border-cr-line text-ink-dim hover:text-ink'"
              @click="viewMode = v.id"
            >
              {{ v.label }}
            </button>
            <div class="flex-1" />
            <span class="text-[9px] font-jetbrains-mono tabular-nums text-ink-faint">
              {{ designStore.nodes.length }} nodes · {{ designStore.connections.length }} links
            </span>
          </div>

          <!-- Viewers -->
          <div
            v-if="mode === 'simple' || viewMode !== 'graph'"
            class="flex-shrink-0 flex overflow-hidden bg-black border-b border-cr-line"
            :style="{ height: (mode === 'simple' ? viewerHeight + 120 : viewerHeight) + 'px' }"
          >
            <template v-if="mode === 'flow' && viewMode === 'dual'">
              <div class="flex-1 min-w-0 border-r border-cr-line overflow-hidden">
                <CompositingCanvas :playhead-time="playheadTime" :is-playing="isPlaying" :target-node-id="designStore.viewer1NodeId" viewer-label="A" />
              </div>
              <div class="flex-1 min-w-0 overflow-hidden">
                <CompositingCanvas :playhead-time="playheadTime" :is-playing="isPlaying" :target-node-id="designStore.viewer2NodeId" viewer-label="B" />
              </div>
            </template>
            <div v-else class="flex-1 min-w-0 overflow-hidden">
              <CompositingCanvas :playhead-time="playheadTime" :is-playing="isPlaying" viewer-label="Output" />
            </div>
          </div>

          <!-- Simple: scrub without leaving the viewer -->
          <div v-if="mode === 'simple'" class="h-8 flex-shrink-0 bg-cr-panel border-b border-cr-line flex items-center gap-3 px-3">
            <span class="text-[9px] uppercase tracking-[0.14em] text-ink-faint">Scrub</span>
            <input
              v-model.number="playheadTime"
              type="range"
              min="0"
              :max="duration"
              :step="1 / fps"
              class="flex-1 accent-accent cursor-pointer"
            />
            <span class="text-[9px] font-jetbrains-mono tabular-nums text-ink-dim">f{{ frame }}</span>
          </div>

          <!-- Flow: graph plus optional spline editor -->
          <template v-if="mode === 'flow'">
            <div
              v-if="viewMode !== 'graph'"
              class="h-px cursor-row-resize hover:bg-accent/60 z-20 transition-colors flex-shrink-0"
              @mousedown="(e) => startDrag(e, 'viewer')"
            />
            <div class="flex-1 flex flex-col min-h-0 overflow-hidden relative">
              <div class="flex-1 overflow-hidden" :style="showCurves ? { height: 'calc(100% - ' + curvesHeight + 'px - 4px)' } : {}">
                <NodeGraph :playhead-time="playheadTime" :is-playing="isPlaying" />
              </div>
              <template v-if="showCurves">
                <div class="h-px cursor-row-resize hover:bg-accent/60 z-20 transition-colors flex-shrink-0" @mousedown="(e) => startDrag(e, 'curves')" />
                <div :style="{ height: curvesHeight + 'px' }" class="flex-shrink-0">
                  <AnimationCurves :playhead-time="playheadTime" @seek="(t) => playheadTime = t" />
                </div>
              </template>
            </div>
          </template>

          <!-- Simple: the selected node's parameters, right under the picture -->
          <div v-else class="flex-1 min-h-0 border-t border-cr-line flex flex-col">
            <div class="h-6 flex-shrink-0 bg-cr-panel border-b border-cr-line flex items-center px-2">
              <span class="text-[9px] uppercase tracking-[0.14em] text-ink-faint">
                {{ designStore.selectedNode ? 'Parameters · ' + designStore.selectedNode.label : 'Parameters · no node selected' }}
              </span>
            </div>
            <div class="flex-1 min-h-0 overflow-hidden">
              <NodeInspector :playhead-time="playheadTime" embedded />
            </div>
          </div>
        </template>
      </div>

      <!-- Inspector: only in Flow, where parameters are not already on screen -->
      <template v-if="mode === 'flow'">
        <div class="w-px cursor-col-resize hover:bg-accent/60 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'inspector')" />
        <div class="bg-cr-panel flex flex-col flex-shrink-0" :style="{ width: inspectorWidth + 'px' }">
          <NodeInspector :playhead-time="playheadTime" />
        </div>
      </template>
    </div>
  </div>
</template>