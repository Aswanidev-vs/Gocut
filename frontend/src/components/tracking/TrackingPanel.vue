<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useTrackingStore } from '../../stores/trackingStore'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import { Crosshair, Play, Square, CheckCircle, AlertCircle, Loader2 } from 'lucide-vue-next'

const trackingStore = useTrackingStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

const selectedAsset = ref(null)
const startTime = ref(0)
const duration = ref(5)

const assets = computed(() => projectStore.project?.assets || [])

function onTrackingProgress(e) {
  if (e.assetId === selectedAsset.value) {
    trackingStore.progress = e.progress
  }
}

function onTrackingComplete(e) {
  if (e.assetId === selectedAsset.value) {
    uiStore.addToast(`Tracking complete: ${e.frameCount} frames, ${(e.confidence * 100).toFixed(0)}% confidence`, 'success', 3000)
  }
}

onMounted(() => {
  window.addEventListener('tracking:progress', onTrackingProgress)
  window.addEventListener('tracking:complete', onTrackingComplete)
})

onUnmounted(() => {
  window.removeEventListener('tracking:progress', onTrackingProgress)
  window.removeEventListener('tracking:complete', onTrackingComplete)
})

async function startAnalysis() {
  if (!selectedAsset.value) {
    uiStore.addToast('Select a clip first', 'error', 2000)
    return
  }
  await trackingStore.analyze(selectedAsset.value, startTime.value, duration.value)
}

function applyToTransform() {
  if (!trackingStore.results || !trackingStore.results.points.length) {
    uiStore.addToast('No tracking data to apply', 'error', 2000)
    return
  }
  // Emit event for DesignWorkspace to handle
  window.dispatchEvent(new CustomEvent('tracking:applyToTransform', {
    detail: trackingStore.results
  }))
  uiStore.addToast('Tracking applied to transform', 'success', 2000)
}
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <div class="p-3 border-b border-border bg-panel/50">
      <div class="flex items-center gap-2">
        <Crosshair :size="14" class="text-accent" />
        <div>
          <div class="text-sm font-bold text-text-primary">Motion Tracking</div>
          <div class="text-[10px] text-text-secondary">Track points or stabilize footage</div>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-3 space-y-4">
      <!-- Asset selector -->
      <div class="space-y-1.5">
        <label class="text-[11px] text-text-secondary font-medium">Source Clip</label>
        <select
          v-model="selectedAsset"
          class="w-full bg-bg border border-border rounded px-2 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent"
        >
          <option :value="null" disabled>Select a clip…</option>
          <option v-for="asset in assets" :key="asset.id" :value="asset.id">
            {{ asset.name || asset.path.split(/[/\\]/).pop() }}
          </option>
        </select>
      </div>

      <!-- Method selector -->
      <div class="space-y-1.5">
        <label class="text-[11px] text-text-secondary font-medium">Method</label>
        <div class="flex gap-1">
          <button
            class="flex-1 py-1.5 rounded text-[10px] font-medium transition-colors"
            :class="trackingStore.method === 'stabilize' ? 'bg-accent/15 text-accent border border-accent/30' : 'bg-bg border border-border text-text-secondary hover:text-text-primary'"
            @click="trackingStore.method = 'stabilize'"
          >
            Stabilize
          </button>
          <button
            class="flex-1 py-1.5 rounded text-[10px] font-medium transition-colors"
            :class="trackingStore.method === 'point' ? 'bg-accent/15 text-accent border border-accent/30' : 'bg-bg border border-border text-text-secondary hover:text-text-primary'"
            @click="trackingStore.method = 'point'"
          >
            Point Track
          </button>
        </div>
      </div>

      <!-- Time range -->
      <div class="grid grid-cols-2 gap-2">
        <div class="space-y-1.5">
          <label class="text-[11px] text-text-secondary font-medium">Start (s)</label>
          <input
            v-model.number="startTime"
            type="number"
            min="0"
            step="0.1"
            class="w-full bg-bg border border-border rounded px-2 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent"
          />
        </div>
        <div class="space-y-1.5">
          <label class="text-[11px] text-text-secondary font-medium">Duration (s)</label>
          <input
            v-model.number="duration"
            type="number"
            min="0.1"
            step="0.1"
            class="w-full bg-bg border border-border rounded px-2 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent"
          />
        </div>
      </div>

      <!-- Stabilization settings -->
      <div v-if="trackingStore.method === 'stabilize'" class="space-y-3">
        <div class="space-y-1.5">
          <label class="text-[11px] text-text-secondary font-medium">Shaking Intensity</label>
          <input
            v-model.number="trackingStore.settings.shaking"
            type="range"
            min="0"
            max="2"
            step="1"
            class="w-full h-1 accent-accent"
          />
          <div class="flex justify-between text-[9px] text-text-secondary">
            <span>Low</span>
            <span>Medium</span>
            <span>High</span>
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] text-text-secondary font-medium">Accuracy</label>
          <input
            v-model.number="trackingStore.settings.accuracy"
            type="range"
            min="0"
            max="2"
            step="1"
            class="w-full h-1 accent-accent"
          />
          <div class="flex justify-between text-[9px] text-text-secondary">
            <span>Fast</span>
            <span>Accurate</span>
            <span>Very Accurate</span>
          </div>
        </div>
      </div>

      <!-- Point tracking region -->
      <div v-if="trackingStore.method === 'point'" class="space-y-2">
        <div class="text-[11px] text-text-secondary font-medium">Tracking Region</div>
        <div class="grid grid-cols-2 gap-2">
          <div class="space-y-1">
            <label class="text-[9px] text-text-secondary">X</label>
            <input v-model.number="trackingStore.settings.regionX" type="number" class="w-full bg-bg border border-border rounded px-2 py-1 text-[10px] text-text-primary outline-none" />
          </div>
          <div class="space-y-1">
            <label class="text-[9px] text-text-secondary">Y</label>
            <input v-model.number="trackingStore.settings.regionY" type="number" class="w-full bg-bg border border-border rounded px-2 py-1 text-[10px] text-text-primary outline-none" />
          </div>
          <div class="space-y-1">
            <label class="text-[9px] text-text-secondary">Width</label>
            <input v-model.number="trackingStore.settings.regionW" type="number" class="w-full bg-bg border border-border rounded px-2 py-1 text-[10px] text-text-primary outline-none" />
          </div>
          <div class="space-y-1">
            <label class="text-[9px] text-text-secondary">Height</label>
            <input v-model.number="trackingStore.settings.regionH" type="number" class="w-full bg-bg border border-border rounded px-2 py-1 text-[10px] text-text-primary outline-none" />
          </div>
        </div>
      </div>

      <!-- Analyze button -->
      <button
        class="w-full flex items-center justify-center gap-2 py-2 rounded-lg bg-accent text-bg text-[11px] font-semibold hover:bg-accent-hover transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        :disabled="trackingStore.isAnalyzing || !selectedAsset"
        @click="startAnalysis"
      >
        <Loader2 v-if="trackingStore.isAnalyzing" :size="12" class="animate-spin" />
        <Play v-else :size="12" />
        {{ trackingStore.isAnalyzing ? 'Analyzing…' : 'Analyze Motion' }}
      </button>

      <!-- Progress bar -->
      <div v-if="trackingStore.isAnalyzing" class="space-y-1">
        <div class="h-1.5 bg-bg rounded-full overflow-hidden">
          <div
            class="h-full bg-accent transition-all duration-300"
            :style="{ width: (trackingStore.progress * 100) + '%' }"
          />
        </div>
        <div class="text-[9px] text-text-secondary text-center">
          {{ (trackingStore.progress * 100).toFixed(0) }}%
        </div>
      </div>

      <!-- Results -->
      <div v-if="trackingStore.results" class="space-y-2">
        <div class="flex items-center gap-2 text-[11px]">
          <CheckCircle :size="12" class="text-green-400" />
          <span class="text-text-primary font-medium">Results</span>
        </div>
        <div class="bg-bg/60 rounded-lg p-2 space-y-1 text-[10px]">
          <div class="flex justify-between">
            <span class="text-text-secondary">Frames tracked</span>
            <span class="text-text-primary font-mono">{{ trackingStore.results.frameCount }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-text-secondary">Confidence</span>
            <span class="text-text-primary font-mono">{{ (trackingStore.results.confidence * 100).toFixed(0) }}%</span>
          </div>
          <div class="flex justify-between">
            <span class="text-text-secondary">Method</span>
            <span class="text-text-primary font-mono">{{ trackingStore.results.method }}</span>
          </div>
        </div>

        <!-- Apply button -->
        <button
          class="w-full flex items-center justify-center gap-2 py-1.5 rounded border border-accent/30 text-accent text-[10px] font-medium hover:bg-accent/10 transition-colors"
          @click="applyToTransform"
        >
          Apply to Transform
        </button>
      </div>

      <!-- Error -->
      <div v-if="trackingStore.error" class="flex items-start gap-2 p-2 rounded-lg bg-red-500/10 border border-red-500/20">
        <AlertCircle :size="12" class="text-red-400 flex-shrink-0 mt-0.5" />
        <div class="text-[10px] text-red-400">{{ trackingStore.error }}</div>
      </div>
    </div>
  </div>
</template>
