<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { useTimelineStore } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import {
  Plus, Trash2, Copy, Play, Pause, RotateCcw,
  ZoomIn, ZoomOut, Maximize2, Magnet,
  Sparkles, Wand2, Zap, Box, Crosshair,
} from 'lucide-vue-next'
import NodeGraph from './NodeGraph.vue'
import NodeInspector from './NodeInspector.vue'
import NodeLibrary from './NodeLibrary.vue'
import AnimationCurves from './AnimationCurves.vue'
import TemplateGallery from './TemplateGallery.vue'
import DesignOnboarding from './DesignOnboarding.vue'
import SimpleEffectsPanel from './SimpleEffectsPanel.vue'
import TrackingPanel from '../tracking/TrackingPanel.vue'
import CompositingCanvas from './CompositingCanvas.vue'
import { useDesignHotkeys } from '../../composables/useDesignHotkeys'

useDesignHotkeys()

const designStore = useDesignStore()
const timelineStore = useTimelineStore()
const uiStore = useUiStore()

// Resizable panel widths & heights
const MIN_CENTER_WIDTH = 360
const MIN_LIBRARY_WIDTH = 160
const MIN_INSPECTOR_WIDTH = 200

const isNarrowScreen = ref(typeof window !== 'undefined' && window.innerWidth < 1100)
const isSmallScreen = ref(typeof window !== 'undefined' && window.innerWidth < 800)

const libraryWidth = ref(220)
const inspectorWidth = ref(280)
const showCurves = ref(false)
const curvesHeight = ref(180)
const viewerHeight = ref(320)
const showOnboarding = computed(() => designStore.nodes.length === 0)
const viewMode = ref('dual') // 'dual' | 'single' | 'graph'

const isPlaying = ref(false)
let playInterval = null
const playheadTime = ref(0)

const tabs = [
  { id: 'nodes', label: 'Tools', icon: Box },
  { id: 'templates', label: 'Templates', icon: Sparkles },
  { id: 'effects', label: 'Effects', icon: Wand2 },
  { id: 'tracking', label: 'Track', icon: Crosshair },
  { id: 'presets', label: 'Presets', icon: Zap },
]
const activeTab = ref('nodes')

// Fusion Toolbar Quick Tool Definitions
const FUSION_SHELF_TOOLS = [
  { type: 'background', label: 'Bg', title: 'Background Generator (Bg)', color: '#F59E0B' },
  { type: 'fastNoise', label: 'Noise', title: 'FastNoise (FN)', color: '#F59E0B' },
  { type: 'textPlus', label: 'Text+', title: 'Text+ Generator', color: '#F59E0B' },
  { type: 'maskPolygon', label: 'Mask', title: 'Polygon Mask (Poly)', color: '#3B82F6' },
  { type: 'transform', label: 'Xf', title: 'Transform (Xf)', color: '#8B5CF6' },
  { type: 'merge', label: 'Mrg', title: 'Merge (Mrg)', color: '#10B981' },
  { type: 'colorCorrector', label: 'CC', title: 'ColorCorrector (CC)', color: '#EC4899' },
  { type: 'blur', label: 'Blur', title: 'Blur Effect', color: '#EC4899' },
]

function quickAddTool(toolType) {
  designStore.addNode(toolType)
}

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
    libraryWidth.value = Math.max(MIN_LIBRARY_WIDTH, Math.min(e.clientX - 260, 400))
  } else if (activeDrag === 'inspector') {
    inspectorWidth.value = Math.max(MIN_INSPECTOR_WIDTH, Math.min(window.innerWidth - e.clientX, 450))
  } else if (activeDrag === 'viewer') {
    viewerHeight.value = Math.max(160, Math.min(e.clientY - 80, window.innerHeight - 250))
  } else if (activeDrag === 'curves') {
    curvesHeight.value = Math.max(100, Math.min(window.innerHeight - e.clientY, window.innerHeight - 150))
  }
}

function stopDrag() {
  activeDrag = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

// Lifecycle: responsive resize + design playback events
onMounted(() => {
  window.addEventListener('design:togglePlay', togglePlay)
  window.addEventListener('design:stepPlayhead', onStepPlayhead)
})

onUnmounted(() => {
  clearInterval(playInterval)
  window.removeEventListener('design:togglePlay', togglePlay)
  window.removeEventListener('design:stepPlayhead', onStepPlayhead)
})

function togglePlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) {
    const fps = 30
    playInterval = setInterval(() => {
      playheadTime.value += 1 / fps
      if (playheadTime.value > designStore.composition.duration) {
        playheadTime.value = 0
      }
    }, 1000 / fps)
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
  const step = (e.detail || 1) * (1 / (designStore.composition.fps || 30))
  playheadTime.value = Math.max(0, Math.min(designStore.composition.duration, playheadTime.value + step))
}

// Onboarding handlers
function startWithTemplate() {
  activeTab.value = 'templates'
}
function startWithEffect() {
  activeTab.value = 'effects'
}
function startWithNodes() {
  activeTab.value = 'nodes'
  if (designStore.nodes.length === 0) {
    const bg = designStore.addNode('background', { x: 100, y: 150, label: 'Background1' })
    const txt = designStore.addNode('textPlus', { x: 100, y: 280, label: 'Text1' })
    const mrg = designStore.addNode('merge', { x: 340, y: 200, label: 'Merge1' })
    const out = designStore.addNode('mediaOut', { x: 560, y: 200, label: 'MediaOut1' })
    
    if (bg && txt && mrg && out) {
      designStore.addConnection(bg.id, 'out', mrg.id, 'bg')
      designStore.addConnection(txt.id, 'out', mrg.id, 'fg')
      designStore.addConnection(mrg.id, 'out', out.id, 'in')
      designStore.setViewer1(txt.id)
      designStore.setViewer2(mrg.id)
    }
  }
}

defineExpose({ libraryWidth, inspectorWidth, curvesHeight, showCurves })
</script>

<template>
  <div class="flex flex-col h-full bg-[#0A0A10] overflow-hidden select-none">
    <!-- Fusion Shelf / Quick Tool Bar -->
    <div class="h-10 bg-[#14141E] border-b border-border/80 flex items-center px-2 gap-1.5 flex-shrink-0 overflow-x-auto">
      <div class="flex items-center gap-1 px-1.5 flex-shrink-0">
        <div class="w-2 h-2 rounded-full bg-pink-500 shadow-[0_0_8px_rgba(236,72,153,0.8)]" />
        <span class="text-[11px] font-bold text-text-primary uppercase tracking-wider">Fusion</span>
      </div>

      <div class="w-px h-5 bg-border/60 mx-1 flex-shrink-0" />

      <!-- Resolve Fusion Quick Tool Icons -->
      <div class="flex items-center gap-1 flex-shrink-0">
        <button
          v-for="tool in FUSION_SHELF_TOOLS"
          :key="tool.type"
          class="flex items-center gap-1 px-2 py-1 rounded bg-[#1C1C28] hover:bg-accent/20 hover:border-accent/50 border border-border/60 text-[10px] font-medium text-text-primary transition-colors shadow-sm"
          :title="tool.title"
          @click="quickAddTool(tool.type)"
        >
          <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: tool.color }" />
          <span>{{ tool.label }}</span>
        </button>
      </div>

      <div class="w-px h-5 bg-border/60 mx-1 flex-shrink-0" />

      <!-- Graph Operations -->
      <div class="flex items-center gap-0.5 flex-shrink-0">
        <button
          class="p-1.5 rounded transition-colors"
          :class="designStore.snapEnabled ? 'text-accent bg-accent/10' : 'text-text-secondary hover:text-text-primary hover:bg-border/60'"
          @click="designStore.snapEnabled = !designStore.snapEnabled"
          title="Snap to Grid (S)"
        >
          <Magnet :size="12" />
        </button>
        <button
          class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          @click="designStore.zoomFit()"
          title="Zoom to Fit (Ctrl+0)"
        >
          <Maximize2 :size="12" />
        </button>
        <button
          class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          @click="designStore.zoomIn()"
          title="Zoom In"
        >
          <ZoomIn :size="12" />
        </button>
        <button
          class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          @click="designStore.zoomOut()"
          title="Zoom Out"
        >
          <ZoomOut :size="12" />
        </button>
      </div>

      <div class="flex-1 flex-shrink-0" />

      <!-- Transport Playback -->
      <div class="flex items-center gap-1 flex-shrink-0 bg-[#0F0F17] px-2 py-0.5 rounded border border-border/60">
        <button
          class="p-1 rounded transition-colors"
          :class="isPlaying ? 'text-pink-400 bg-pink-500/10' : 'text-text-secondary hover:text-text-primary'"
          @click="togglePlay"
          :title="isPlaying ? 'Pause (Space)' : 'Play (Space)'"
        >
          <Pause v-if="isPlaying" :size="12" />
          <Play v-else :size="12" />
        </button>
        <button
          class="p-1 rounded text-text-secondary hover:text-text-primary transition-colors"
          @click="resetPlayhead"
          title="Reset to 0s"
        >
          <RotateCcw :size="12" />
        </button>
        <div class="text-[10px] text-text-primary px-1 font-mono">
          {{ playheadTime.toFixed(2) }}s / {{ designStore.composition.duration.toFixed(2) }}s
        </div>
      </div>

      <div class="w-px h-5 bg-border/60 mx-1 flex-shrink-0" />

      <!-- View layout switcher -->
      <div class="flex items-center gap-0.5 bg-[#0F0F17] rounded p-0.5 flex-shrink-0 border border-border/60">
        <button
          class="px-2 py-0.5 rounded text-[10px] font-medium transition-colors"
          :class="viewMode === 'dual' ? 'bg-accent/20 text-accent font-semibold' : 'text-text-secondary hover:text-text-primary'"
          @click="viewMode = 'dual'"
          title="Fusion Dual Viewers (Viewer 1 & 2)"
        >
          Dual
        </button>
        <button
          class="px-2 py-0.5 rounded text-[10px] font-medium transition-colors"
          :class="viewMode === 'single' ? 'bg-accent/20 text-accent font-semibold' : 'text-text-secondary hover:text-text-primary'"
          @click="viewMode = 'single'"
          title="Single Output Viewer"
        >
          Single
        </button>
        <button
          class="px-2 py-0.5 rounded text-[10px] font-medium transition-colors"
          :class="viewMode === 'graph' ? 'bg-accent/20 text-accent font-semibold' : 'text-text-secondary hover:text-text-primary'"
          @click="viewMode = 'graph'"
          title="Full Flow Graph"
        >
          Flow Only
        </button>
      </div>

      <button
        class="flex items-center gap-1 px-2 py-1 rounded text-[10px] font-medium border border-border/60 transition-colors flex-shrink-0"
        :class="showCurves ? 'text-accent bg-accent/15 border-accent/40' : 'text-text-secondary hover:text-text-primary bg-[#1C1C28]'"
        @click="showCurves = !showCurves"
        title="Toggle Fusion Spline / Keyframe Editor"
      >
        <Zap :size="11" /> <span>Spline</span>
      </button>
    </div>

    <!-- Main Workspace Area -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Left Panel: Tools / Templates / Presets -->
      <div class="bg-[#12121A] border-r border-border flex flex-col flex-shrink-0" :style="{ width: libraryWidth + 'px' }">
        <div class="flex items-center gap-0.5 p-1 border-b border-border bg-[#161622]">
          <button
            v-for="t in tabs"
            :key="t.id"
            class="flex-1 flex flex-col items-center justify-center gap-0.5 py-1.5 rounded text-[9px] font-medium transition-colors"
            :class="activeTab === t.id ? 'bg-accent/15 text-accent font-semibold' : 'text-text-secondary hover:text-text-primary'"
            @click="activeTab = t.id"
            :title="t.label"
          >
            <component :is="t.icon" :size="11" />
            <span>{{ t.label }}</span>
          </button>
        </div>
        <div class="flex-1 overflow-y-auto">
          <NodeLibrary v-if="activeTab === 'nodes'" />
          <TemplateGallery v-else-if="activeTab === 'templates'" @insert="activeTab = 'nodes'" />
          <SimpleEffectsPanel v-else-if="activeTab === 'effects'" @applied="activeTab = 'nodes'" />
          <TrackingPanel v-else-if="activeTab === 'tracking'" />
          <div v-else class="p-2">
            <div class="text-[10px] text-text-secondary uppercase tracking-wider mb-2 px-1">Saved Presets</div>
            <div class="space-y-1">
              <div
                v-for="p in designStore.presets"
                :key="p.id"
                class="p-2 rounded bg-bg/60 border border-border/60 hover:border-accent/50 transition-colors cursor-pointer"
                @click="designStore.loadPreset(p.id)"
              >
                <div class="text-[11px] text-text-primary font-medium">{{ p.name }}</div>
                <div class="text-[9px] text-text-secondary mt-0.5">{{ p.nodes.length }} nodes</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="w-1 cursor-col-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'library')" />

      <!-- Center: DaVinci Resolve Fusion Viewers (Top) + Node Graph & Spline (Bottom) -->
      <div class="flex-1 flex flex-col overflow-hidden bg-[#0A0A10]">
        <DesignOnboarding
          v-if="showOnboarding"
          @start-template="startWithTemplate"
          @start-effect="startWithEffect"
          @start-blank="startWithNodes"
        />
        <template v-else>
          <!-- Top: Fusion Viewers (Dual or Single) -->
          <div
            v-if="viewMode !== 'graph'"
            class="flex-shrink-0 border-b border-border flex overflow-hidden bg-black"
            :style="{ height: viewerHeight + 'px' }"
          >
            <!-- Dual Viewers Mode -->
            <template v-if="viewMode === 'dual'">
              <!-- Viewer 1 (Left) -->
              <div class="flex-1 min-w-0 border-r border-border/70 overflow-hidden">
                <CompositingCanvas
                  :playhead-time="playheadTime"
                  :is-playing="isPlaying"
                  :target-node-id="designStore.viewer1NodeId"
                  viewer-label="Viewer 1 [1]"
                />
              </div>
              <!-- Viewer 2 (Right) -->
              <div class="flex-1 min-w-0 overflow-hidden">
                <CompositingCanvas
                  :playhead-time="playheadTime"
                  :is-playing="isPlaying"
                  :target-node-id="designStore.viewer2NodeId"
                  viewer-label="Viewer 2 [2]"
                />
              </div>
            </template>
            <!-- Single Output Mode -->
            <template v-else-if="viewMode === 'single'">
              <div class="flex-1 min-w-0 overflow-hidden">
                <CompositingCanvas
                  :playhead-time="playheadTime"
                  :is-playing="isPlaying"
                  viewer-label="Final Output"
                />
              </div>
            </template>
          </div>

          <div
            v-if="viewMode !== 'graph'"
            class="h-1 cursor-row-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors"
            @mousedown="(e) => startDrag(e, 'viewer')"
          />

          <!-- Bottom: Node Graph Flow Area -->
          <div class="flex-1 flex flex-col min-h-0 overflow-hidden relative">
            <div class="flex-1 overflow-hidden" :style="showCurves ? { height: `calc(100% - ${curvesHeight}px - 4px)` } : {}">
              <NodeGraph :playhead-time="playheadTime" :is-playing="isPlaying" />
            </div>

            <!-- Spline / Curves Editor -->
            <template v-if="showCurves">
              <div class="h-1 cursor-row-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'curves')" />
              <div :style="{ height: curvesHeight + 'px' }" class="flex-shrink-0">
                <AnimationCurves :playhead-time="playheadTime" @seek="(t) => playheadTime = t" />
              </div>
            </template>
          </div>
        </template>
      </div>

      <div class="w-1 cursor-col-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'inspector')" />

      <!-- Right: Fusion Inspector / Modifier Settings -->
      <div class="bg-[#12121A] border-l border-border flex flex-col flex-shrink-0" :style="{ width: inspectorWidth + 'px' }">
        <NodeInspector :playhead-time="playheadTime" />
      </div>
    </div>
  </div>
</template>
