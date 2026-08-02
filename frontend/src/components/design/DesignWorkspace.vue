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

// Resizable panel widths
const libraryWidth = ref(240)
const inspectorWidth = ref(300)
const showCurves = ref(false)
const curvesHeight = ref(160)
const showOnboarding = computed(() => designStore.nodes.length === 0)
const viewMode = ref('graph') // 'graph' | 'preview' | 'split'

const isPlaying = ref(false)
let playInterval = null
const playheadTime = ref(0)

const tabs = [
  { id: 'nodes', label: 'Nodes', icon: Box },
  { id: 'templates', label: 'Templates', icon: Sparkles },
  { id: 'effects', label: 'Effects', icon: Wand2 },
  { id: 'tracking', label: 'Track', icon: Crosshair },
  { id: 'presets', label: 'Presets', icon: Zap },
]
const activeTab = ref('nodes')

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
    libraryWidth.value = Math.max(180, Math.min(e.clientX, window.innerWidth - inspectorWidth.value - 200))
  } else if (activeDrag === 'inspector') {
    inspectorWidth.value = Math.max(220, Math.min(window.innerWidth - e.clientX, window.innerWidth - libraryWidth.value - 200))
  } else if (activeDrag === 'curves') {
    curvesHeight.value = Math.max(80, Math.min(window.innerHeight - e.clientY, window.innerHeight - 200))
  }
}
function stopDrag() {
  activeDrag = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}

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

onMounted(() => {
  window.addEventListener('design:togglePlay', togglePlay)
  window.addEventListener('design:stepPlayhead', onStepPlayhead)
})

onUnmounted(() => {
  clearInterval(playInterval)
  window.removeEventListener('design:togglePlay', togglePlay)
  window.removeEventListener('design:stepPlayhead', onStepPlayhead)
})

function onStepPlayhead(e) {
  const step = (e.detail || 1) * (1 / (designStore.composition.fps || 30))
  playheadTime.value = Math.max(0, Math.min(designStore.composition.duration, playheadTime.value + step))
}

// Watch timeline playhead for sync
watch(() => timelineStore.currentTime, (t) => {
  // Optionally sync from main timeline
})

// Onboarding handlers
function startWithTemplate() {
  activeTab.value = 'templates'
}
function startWithEffect() {
  activeTab.value = 'effects'
}
function startWithNodes() {
  activeTab.value = 'nodes'
  // Seed a basic graph so the user isn't staring at emptiness
  if (designStore.nodes.length === 0) {
    designStore.addNode('text', { x: 120, y: 200, label: 'Title', params: { text: 'Hello', fontSize: 64, color: '#FFFFFF' } })
    designStore.addNode('output', { x: 500, y: 200, label: 'Output' })
  }
}

defineExpose({ libraryWidth, inspectorWidth, curvesHeight, showCurves })
</script>

<template>
  <div class="flex flex-col h-full bg-bg overflow-hidden">
    <!-- Design Toolbar -->
    <div class="h-10 bg-panel border-b border-border flex items-center px-2 gap-1 flex-shrink-0">
      <div class="flex items-center gap-1.5 px-2">
        <div class="w-1.5 h-1.5 rounded-full bg-pink-500 animate-pulse" />
        <span class="text-[11px] font-semibold text-text-primary">Design</span>
        <span class="text-[10px] text-text-secondary px-1.5 py-0.5 rounded border border-border">Fusion-style</span>
      </div>

      <div class="w-px h-5 bg-border mx-1" />

      <div class="flex items-center gap-0.5">
        <button
          class="flex items-center gap-1 px-2 py-1 rounded text-[11px] text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          @click="designStore.addNode('media')"
          title="Add Media Node"
        >
          <Plus :size="11" /> Add Node
        </button>
        <button
          class="flex items-center gap-1 px-2 py-1 rounded text-[11px] text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          :disabled="!designStore.selectedNodeId"
          :class="!designStore.selectedNodeId && 'opacity-40 cursor-not-allowed'"
          @click="designStore.duplicateSelectedNode()"
          title="Duplicate (Ctrl+D)"
        >
          <Copy :size="11" /> Duplicate
        </button>
        <button
          class="flex items-center gap-1 px-2 py-1 rounded text-[11px] text-text-secondary hover:text-red-400 hover:bg-red-500/10 transition-colors"
          :disabled="!designStore.selectedNodeId"
          :class="!designStore.selectedNodeId && 'opacity-40 cursor-not-allowed'"
          @click="designStore.removeSelectedNode()"
          title="Delete (Del)"
        >
          <Trash2 :size="11" /> Delete
        </button>
      </div>

      <div class="w-px h-5 bg-border mx-1" />

      <div class="flex items-center gap-0.5">
        <button
          class="p-1.5 rounded transition-colors"
          :class="designStore.snapEnabled ? 'text-accent bg-accent/10' : 'text-text-secondary hover:text-text-primary hover:bg-border/60'"
          @click="designStore.snapEnabled = !designStore.snapEnabled"
          title="Snap to Grid"
        >
          <Magnet :size="12" />
        </button>
        <button
          class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          @click="designStore.zoomFit()"
          title="Zoom to Fit"
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
        <div class="text-[10px] text-text-secondary px-1.5 font-mono">
          {{ Math.round(designStore.zoom * 100) }}%
        </div>
      </div>

      <div class="flex-1" />

      <!-- Composition info -->
      <div class="flex items-center gap-2 px-2">
        <div class="text-[10px] text-text-secondary">
          <span class="text-text-primary font-medium">{{ designStore.composition.name }}</span>
          <span class="mx-1">·</span>
          <span>{{ designStore.composition.width }}×{{ designStore.composition.height }}</span>
          <span class="mx-1">·</span>
          <span class="font-mono">{{ designStore.composition.fps }}fps</span>
        </div>
      </div>

      <div class="w-px h-5 bg-border mx-1" />

      <!-- Playback -->
      <div class="flex items-center gap-0.5">
        <button
          class="p-1.5 rounded transition-colors"
          :class="isPlaying ? 'text-pink-400 bg-pink-500/10' : 'text-text-secondary hover:text-text-primary hover:bg-border/60'"
          @click="togglePlay"
          :title="isPlaying ? 'Pause' : 'Play'"
        >
          <Pause v-if="isPlaying" :size="12" />
          <Play v-else :size="12" />
        </button>
        <button
          class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
          @click="resetPlayhead"
          title="Reset to Start"
        >
          <RotateCcw :size="12" />
        </button>
        <div class="text-[10px] text-text-secondary px-2 font-mono">
          {{ playheadTime.toFixed(2) }}s / {{ designStore.composition.duration.toFixed(2) }}s
        </div>
      </div>

      <div class="w-px h-5 bg-border mx-1" />

      <!-- View mode toggle -->
      <div class="flex items-center gap-0.5 bg-bg/60 rounded p-0.5">
        <button
          class="px-2 py-1 rounded text-[10px] font-medium transition-colors"
          :class="viewMode === 'graph' ? 'bg-accent/15 text-accent' : 'text-text-secondary hover:text-text-primary'"
          @click="viewMode = 'graph'"
          title="Node Graph"
        >
          Graph
        </button>
        <button
          class="px-2 py-1 rounded text-[10px] font-medium transition-colors"
          :class="viewMode === 'preview' ? 'bg-accent/15 text-accent' : 'text-text-secondary hover:text-text-primary'"
          @click="viewMode = 'preview'"
          title="Compositing Preview"
        >
          Preview
        </button>
        <button
          class="px-2 py-1 rounded text-[10px] font-medium transition-colors"
          :class="viewMode === 'split' ? 'bg-accent/15 text-accent' : 'text-text-secondary hover:text-text-primary'"
          @click="viewMode = 'split'"
          title="Split View"
        >
          Split
        </button>
      </div>

      <div class="w-px h-5 bg-border mx-1" />

      <button
        class="flex items-center gap-1 px-2 py-1 rounded text-[11px] text-text-secondary hover:text-text-primary hover:bg-border/60 transition-colors"
        :class="showCurves && 'text-accent bg-accent/10'"
        @click="showCurves = !showCurves"
        title="Toggle Animation Curves"
      >
        <Zap :size="11" /> Curves
      </button>
    </div>

    <!-- Main Content: Library | Graph | Inspector -->
    <div class="flex-1 flex overflow-hidden">
      <!-- Left: Node Library -->
      <div class="bg-panel border-r border-border flex flex-col flex-shrink-0" :style="{ width: libraryWidth + 'px' }">
        <div class="flex items-center gap-0.5 p-1.5 border-b border-border">
          <button
            v-for="t in tabs"
            :key="t.id"
            class="flex-1 flex items-center justify-center gap-1 py-1.5 rounded text-[10px] transition-colors"
            :class="activeTab === t.id ? 'bg-accent/10 text-accent' : 'text-text-secondary hover:text-text-primary hover:bg-border/50'"
            @click="activeTab = t.id"
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
              <div v-if="designStore.presets.length === 0" class="text-[10px] text-text-secondary text-center py-4">
                No presets saved yet
              </div>
            </div>
            <button
              v-if="designStore.selectedNodeId"
              class="w-full mt-2 py-1.5 rounded border border-dashed border-border hover:border-accent/50 text-[10px] text-text-secondary hover:text-text-primary transition-colors"
              @click="designStore.saveAsPreset()"
            >
              + Save current as preset
            </button>
          </div>
        </div>
      </div>

      <div class="w-1 cursor-col-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'library')" />

      <!-- Center: Node Graph + Curves (or Onboarding when empty) -->
      <div class="flex-1 flex flex-col overflow-hidden bg-bg">
        <DesignOnboarding
          v-if="showOnboarding"
          @start-template="startWithTemplate"
          @start-effect="startWithEffect"
          @start-blank="startWithNodes"
        />
        <template v-else>
          <!-- Split view: preview on top, graph on bottom -->
          <template v-if="viewMode === 'split'">
            <div class="flex-1 min-h-0 border-b border-border">
              <CompositingCanvas ref="compositingRef" :playhead-time="playheadTime" :is-playing="isPlaying" />
            </div>
            <div class="flex-1 min-h-0">
              <NodeGraph :playhead-time="playheadTime" :is-playing="isPlaying" />
            </div>
          </template>
          <!-- Preview only -->
          <template v-else-if="viewMode === 'preview'">
            <div class="flex-1 overflow-hidden">
              <CompositingCanvas ref="compositingRef" :playhead-time="playheadTime" :is-playing="isPlaying" />
            </div>
          </template>
          <!-- Graph only (default) -->
          <template v-else>
            <div class="flex-1 overflow-hidden" :style="showCurves ? { height: `calc(100% - ${curvesHeight}px - 4px)` } : {}">
              <NodeGraph :playhead-time="playheadTime" :is-playing="isPlaying" />
            </div>
            <template v-if="showCurves">
              <div class="h-1 cursor-row-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'curves')" />
              <div :style="{ height: curvesHeight + 'px' }" class="flex-shrink-0">
                <AnimationCurves :playhead-time="playheadTime" @seek="(t) => playheadTime = t" />
              </div>
            </template>
          </template>
        </template>
      </div>

      <div class="w-1 cursor-col-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'inspector')" />

      <!-- Right: Node Inspector -->
      <div class="bg-panel border-l border-border flex flex-col flex-shrink-0" :style="{ width: inspectorWidth + 'px' }">
        <NodeInspector :playhead-time="playheadTime" />
      </div>
    </div>
  </div>
</template>
