<script setup>
import { computed, watch } from 'vue'
import { useTimelineStore, getInterpolatedProperty } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import { ChevronRight, Trash2, Copy, RotateCcw, Diamond } from 'lucide-vue-next'
import ColorInspector from '../inspector/ColorInspector.vue'

const timelineStore = useTimelineStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

const hasSelection = computed(() => timelineStore.selectedClips.length > 0)
const selectedClip = computed(() => timelineStore.selectedClips[0])
const asset = computed(() => selectedClip.value ? projectStore.getAsset(selectedClip.value.assetId) : null)
const isText = computed(() => !!selectedClip.value?.textProps)
const isVideo = computed(() => {
  if (!selectedClip.value) return false
  const track = timelineStore.tracks.find(t => t.id === selectedClip.value.trackId)
  return track?.type === 'video'
})
const isAudio = computed(() => {
  if (!selectedClip.value) return false
  const track = timelineStore.tracks.find(t => t.id === selectedClip.value.trackId)
  return track?.type === 'audio'
})

const activeTab = computed({
  get: () => uiStore.activeInspectorTab,
  set: (v) => uiStore.setActiveInspectorTab(v),
})

// Sync workspace changes to inspector tab
watch(() => uiStore.activeWorkspace, (ws) => {
  if (ws === 'color') activeTab.value = 'color'
  else if (ws === 'audio') activeTab.value = 'audio'
  else activeTab.value = 'edit'
})

const tabs = computed(() => {
  const list = [{ id: 'edit', label: 'Edit' }]
  if (isVideo.value || isText.value) list.push({ id: 'color', label: 'Color' })
  if (isVideo.value || isAudio.value) list.push({ id: 'audio', label: 'Audio' })
  if (isText.value) list.push({ id: 'text', label: 'Text' })
  return list
})

const clipTime = computed(() => {
  if (!selectedClip.value) return 0
  return timelineStore.currentTime - selectedClip.value.startTime
})

function hasKeyframe(prop) {
  if (!selectedClip.value || !selectedClip.value.keyframes) return false
  const t = clipTime.value
  return selectedClip.value.keyframes.some(k => k.property === prop && Math.abs(k.time - t) < 0.001)
}

function hasAnyKeyframe(prop) {
  if (!selectedClip.value || !selectedClip.value.keyframes) return false
  return selectedClip.value.keyframes.some(k => k.property === prop)
}

function getPropValue(prop, defaultValue) {
  if (!selectedClip.value) return defaultValue
  if (hasAnyKeyframe(prop)) {
    return getInterpolatedProperty(selectedClip.value, prop, clipTime.value, defaultValue)
  }
  if (prop === 'opacity') {
    return selectedClip.value.opacity !== undefined ? selectedClip.value.opacity : 1.0
  }
  return selectedClip.value.transform?.[prop] !== undefined ? selectedClip.value.transform[prop] : defaultValue
}

function toggleKeyframe(prop, value) {
  if (!selectedClip.value) return
  if (hasKeyframe(prop)) {
    timelineStore.removeKeyframeAtTime(selectedClip.value.id, prop, clipTime.value)
  } else {
    timelineStore.addKeyframe(selectedClip.value.id, prop, value, clipTime.value)
  }
}

function setTransform(key, val) {
  if (!selectedClip.value) return
  if (hasAnyKeyframe(key)) {
    timelineStore.addKeyframe(selectedClip.value.id, key, val, clipTime.value)
  }
  timelineStore.updateClipTransform(selectedClip.value.id, { [key]: val })
}
function setColor(key, val) {
  if (!selectedClip.value) return
  timelineStore.updateClipColor(selectedClip.value.id, { [key]: val })
}
function setTextProp(key, val) {
  if (!selectedClip.value) return
  const tp = { ...selectedClip.value.textProps, [key]: val }
  timelineStore.updateClip(selectedClip.value.id, { textProps: tp })
}
function updateClipField(key, val) {
  if (!selectedClip.value) return
  if (key === 'opacity' && hasAnyKeyframe('opacity')) {
    timelineStore.addKeyframe(selectedClip.value.id, 'opacity', val, clipTime.value)
  }
  timelineStore.updateClip(selectedClip.value.id, { [key]: val })
}
function resetTransform() {
  if (!selectedClip.value) return
  timelineStore.updateClipTransform(selectedClip.value.id, {
    x: 0, y: 0, scaleX: 1, scaleY: 1, rotation: 0, flipH: false, flipV: false,
    cropX: 0, cropY: 0, cropW: 0, cropH: 0,
  })
}
function resetColor() {
  if (!selectedClip.value) return
  timelineStore.updateClipColor(selectedClip.value.id, {
    brightness: 0, contrast: 0, saturation: 0, hue: 0,
    sharpness: 0, vignette: 0, grain: 0, blur: 0,
    tint: 0, temp: 0, highlights: 0, shadows: 0,
    liftR: 0, liftG: 0, liftB: 0,
    gammaR: 0, gammaG: 0, gammaB: 0,
    gainR: 0, gainG: 0, gainB: 0,
    curves: '',
    chromaKeyColor: '',
    chromaKeySimilarity: 0.01,
    chromaKeyBlend: 0.0,
  })
}
function deleteSelected() { timelineStore.removeSelected() }
function duplicateSelected() {
  if (!selectedClip.value) return
  const src = selectedClip.value
  timelineStore.addClip({
    assetId: src.assetId, trackId: src.trackId, trackType: null,
    startTime: src.startTime + src.duration, duration: src.duration,
    trimStart: src.trimStart, trimEnd: src.trimEnd,
    textProps: src.textProps ? JSON.parse(JSON.stringify(src.textProps)) : null,
    stickerProps: src.stickerProps ? JSON.parse(JSON.stringify(src.stickerProps)) : null,
  })
}

const filterPresets = [
  { name: 'Natural',   color: { brightness: 0,  contrast: 5,   saturation: 0,  hue: 0 } },
  { name: 'Cinema',    color: { brightness: -5, contrast: 15,  saturation: -10, hue: 0 } },
  { name: 'Warm',      color: { brightness: 5,  contrast: 0,   saturation: 10, hue: 5, temp: 10 } },
  { name: 'Cool',      color: { brightness: 0,  contrast: 5,   saturation: -5, hue: -5, temp: -10 } },
  { name: 'Vintage',   color: { brightness: 5,  contrast: -5,  saturation: -25, hue: 10, temp: 15 } },
  { name: 'B&W',       color: { saturation: -100 } },
  { name: 'Vivid',     color: { brightness: 5,  contrast: 20,  saturation: 30 } },
  { name: 'Matte',     color: { brightness: 5,  contrast: -10, saturation: -15, hue: -5 } },
  { name: 'Golden',    color: { brightness: 10, contrast: 0,   saturation: 15, hue: 8, temp: 20 } },
  { name: 'Cyberpunk', color: { brightness: 0,  contrast: 20,  saturation: 25, hue: -15, temp: -20 } },
  { name: 'Soft',      color: { brightness: 8,  contrast: -10, saturation: -10 } },
  { name: 'Fade',      color: { brightness: 15, contrast: -20, saturation: -30 } },
]
function applyPreset(p) {
  if (!selectedClip.value) return
  timelineStore.updateClipColor(selectedClip.value.id, p.color)
  uiStore.addToast('Applied ' + p.name, 'success', 1200)
}
function fileName(p) { if (!p) return ''; return p.split(/[\\/]/).pop() || p }
</script>

<template>
  <div class="bg-panel border-l border-border flex flex-col overflow-hidden flex-shrink-0">
    <div v-if="!hasSelection" class="flex-1 flex flex-col items-center justify-center text-text-secondary text-xs gap-2 px-4 text-center">
      <ChevronRight :size="20" class="opacity-40" />
      <div>Select a clip on the timeline to inspect it.</div>
    </div>
    <div v-else class="flex flex-col h-full">
      <!-- Header -->
      <div class="px-3 py-2 border-b border-border flex items-center gap-2">
        <div class="flex-1 min-w-0">
          <div class="text-[10px] text-text-secondary uppercase tracking-wider">{{ asset?.type || (isText ? 'Text' : 'Clip') }}</div>
          <div class="text-sm text-text-primary truncate">{{ isText ? selectedClip.textProps.text : fileName(asset?.path || '') }}</div>
        </div>
        <button class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border" @click="duplicateSelected" title="Duplicate"><Copy :size="12" /></button>
        <button class="p-1.5 rounded text-text-secondary hover:text-red-400 hover:bg-border" @click="deleteSelected" title="Delete"><Trash2 :size="12" /></button>
      </div>

      <!-- Tab bar -->
      <div class="flex items-center gap-0.5 px-2 py-1.5 border-b border-border">
        <button v-for="s in tabs" :key="s.id"
          class="px-2.5 py-1 rounded text-[11px] font-medium transition-colors"
          :class="activeTab === s.id ? 'bg-accent/15 text-accent' : 'text-text-secondary hover:text-text-primary hover:bg-border/60'"
          @click="activeTab = s.id">{{ s.label }}</button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-3 space-y-4">

        <!-- ======== EDIT TAB (Hub) ======== -->
        <template v-if="activeTab === 'edit'">
          <!-- Quick Transform -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider">Transform</h4>
              <button class="p-0.5 rounded text-text-secondary hover:text-accent" @click="resetTransform" title="Reset"><RotateCcw :size="10" /></button>
            </div>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <div class="flex items-center justify-between mb-1"><label class="text-[10px] text-text-secondary">X</label><button class="text-text-secondary hover:text-accent" :class="{'text-accent': hasKeyframe('x')}" @click="toggleKeyframe('x', getPropValue('x', 0))"><Diamond :size="10" :fill="hasKeyframe('x') ? 'currentColor' : 'none'" /></button></div>
                <input type="number" :value="getPropValue('x', 0)" @input="(e) => setTransform('x', parseFloat(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
              <div>
                <div class="flex items-center justify-between mb-1"><label class="text-[10px] text-text-secondary">Y</label><button class="text-text-secondary hover:text-accent" :class="{'text-accent': hasKeyframe('y')}" @click="toggleKeyframe('y', getPropValue('y', 0))"><Diamond :size="10" :fill="hasKeyframe('y') ? 'currentColor' : 'none'" /></button></div>
                <input type="number" :value="getPropValue('y', 0)" @input="(e) => setTransform('y', parseFloat(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-2 mt-2">
              <div>
                <div class="flex items-center justify-between mb-1"><label class="text-[10px] text-text-secondary">Scale W</label><button class="text-text-secondary hover:text-accent" :class="{'text-accent': hasKeyframe('scaleX')}" @click="toggleKeyframe('scaleX', getPropValue('scaleX', 1))"><Diamond :size="10" :fill="hasKeyframe('scaleX') ? 'currentColor' : 'none'" /></button></div>
                <input type="number" step="0.1" :value="getPropValue('scaleX', 1)" @input="(e) => setTransform('scaleX', parseFloat(e.target.value) || 1)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
              <div>
                <div class="flex items-center justify-between mb-1"><label class="text-[10px] text-text-secondary">Scale H</label><button class="text-text-secondary hover:text-accent" :class="{'text-accent': hasKeyframe('scaleY')}" @click="toggleKeyframe('scaleY', getPropValue('scaleY', 1))"><Diamond :size="10" :fill="hasKeyframe('scaleY') ? 'currentColor' : 'none'" /></button></div>
                <input type="number" step="0.1" :value="getPropValue('scaleY', 1)" @input="(e) => setTransform('scaleY', parseFloat(e.target.value) || 1)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
            </div>
            <div class="mt-2">
              <div class="flex items-center justify-between mb-1"><label class="text-[10px] text-text-secondary">Rotation</label><button class="text-text-secondary hover:text-accent" :class="{'text-accent': hasKeyframe('rotation')}" @click="toggleKeyframe('rotation', getPropValue('rotation', 0))"><Diamond :size="10" :fill="hasKeyframe('rotation') ? 'currentColor' : 'none'" /></button></div>
              <div class="flex items-center gap-2">
                <input type="range" min="-180" max="180" step="1" :value="getPropValue('rotation', 0)" @input="(e) => setTransform('rotation', parseFloat(e.target.value))" class="flex-1 accent-accent" />
                <input type="number" :value="getPropValue('rotation', 0)" @input="(e) => setTransform('rotation', parseFloat(e.target.value) || 0)" class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
            </div>
            <div class="flex items-center gap-2 mt-2">
              <button class="px-2 py-1 rounded text-xs border" :class="selectedClip.transform.flipH ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTransform('flipH', !selectedClip.transform.flipH)">Flip H</button>
              <button class="px-2 py-1 rounded text-xs border" :class="selectedClip.transform.flipV ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTransform('flipV', !selectedClip.transform.flipV)">Flip V</button>
            </div>
          </div>

          <hr class="border-border" />

          <!-- Quick Color (essentials only) -->
          <div v-if="isVideo || isText">
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Quick Color</h4>
            <div v-for="k in [['brightness',-100,100],['contrast',-100,100],['saturation',-100,100]]" :key="k[0]">
              <div class="flex justify-between text-[10px] text-text-secondary capitalize"><span>{{ k[0] }}</span><span class="font-mono">{{ selectedClip.color[k[0]] }}</span></div>
              <input type="range" :min="k[1]" :max="k[2]" :value="selectedClip.color[k[0]]" @input="(e) => setColor(k[0], parseInt(e.target.value))" class="w-full accent-accent" />
            </div>
          </div>

          <hr class="border-border" />

          <!-- Quick Audio -->
          <div v-if="isVideo || isAudio">
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Volume</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0" max="2" step="0.01" :value="selectedClip.volume" @input="(e) => updateClipField('volume', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="Math.round(selectedClip.volume * 100)" :step="1" min="0" max="200" @input="(e) => updateClipField('volume', Math.max(0, Math.min(2, (parseFloat(e.target.value) || 0) / 100)))" class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>

          <hr class="border-border" />

          <!-- Speed & Opacity -->
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Speed</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0.1" max="4" step="0.1" :value="selectedClip.speed" @input="(e) => updateClipField('speed', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="selectedClip.speed" :step="0.1" min="0.1" @input="(e) => updateClipField('speed', parseFloat(e.target.value) || 1)" class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider">Opacity</h4>
              <button class="text-text-secondary hover:text-accent" :class="{'text-accent': hasKeyframe('opacity')}" @click="toggleKeyframe('opacity', getPropValue('opacity', 1.0))"><Diamond :size="10" :fill="hasKeyframe('opacity') ? 'currentColor' : 'none'" /></button>
            </div>
            <div class="flex items-center gap-2">
              <input type="range" min="0" max="1" step="0.01" :value="getPropValue('opacity', 1.0)" @input="(e) => updateClipField('opacity', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="Math.round(getPropValue('opacity', 1.0) * 100)" :step="1" min="0" max="100" @input="(e) => updateClipField('opacity', Math.max(0, Math.min(1, (parseFloat(e.target.value) || 0) / 100)))" class="w-14 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
        </template>

        <!-- ======== COLOR TAB (Advanced) ======== -->
        <template v-if="activeTab === 'color'">
          <ColorInspector class="mb-4" />

          <hr class="border-border my-3" />

          <div class="flex items-center justify-between mb-1">
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider font-bold">Primary Adjustments</h4>
            <button class="p-0.5 rounded text-text-secondary hover:text-accent" @click="resetColor" title="Reset all"><RotateCcw :size="10" /></button>
          </div>

          <!-- Primary sliders -->
          <div v-for="k in [
            ['brightness',-100,100],['contrast',-100,100],['saturation',-100,100],['hue',-180,180],
            ['temp',-100,100],['tint',-100,100],['highlights',-100,100],['shadows',-100,100],
            ['sharpness',0,100],['vignette',0,100],['grain',0,100],['blur',0,20]
          ]" :key="k[0]">
            <div class="flex justify-between text-[10px] text-text-secondary capitalize"><span>{{ k[0] }}</span><span class="font-mono">{{ selectedClip.color[k[0]] }}</span></div>
            <input type="range" :min="k[1]" :max="k[2]" :value="selectedClip.color[k[0]]" @input="(e) => setColor(k[0], parseInt(e.target.value))" class="w-full accent-accent" />
          </div>

          <hr class="border-border" />

          <!-- Color Wheels (Lift / Gamma / Gain) -->
          <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mt-2 mb-2">Color Wheels</h4>
          <div v-for="wheel in ['lift', 'gamma', 'gain']" :key="wheel" class="mb-3">
            <div class="text-[10px] text-text-secondary capitalize mb-1">{{ wheel }}</div>
            <div class="grid grid-cols-3 gap-1.5">
              <div v-for="ch in ['R','G','B']" :key="ch">
                <label class="text-[9px] font-mono" :class="ch === 'R' ? 'text-red-400' : ch === 'G' ? 'text-green-400' : 'text-blue-400'">{{ ch }}</label>
                <input type="range" min="-100" max="100" :value="selectedClip.color[wheel + ch]" @input="(e) => setColor(wheel + ch, parseInt(e.target.value))" class="w-full" :class="ch === 'R' ? 'accent-red-400' : ch === 'G' ? 'accent-green-400' : 'accent-blue-400'" />
              </div>
            </div>
          </div>

          <hr class="border-border" />

          <!-- Filter Presets -->
          <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mt-2 mb-2">Filter Presets</h4>
          <div class="grid grid-cols-3 gap-1.5">
            <button v-for="p in filterPresets" :key="p.name" class="aspect-video rounded border border-border/60 hover:border-accent transition-colors flex items-center justify-center" :style="{ background: 'linear-gradient(135deg, ' + (((p.color.temp||0) > 0 ? '#F59E0B' : (p.color.temp||0) < 0 ? '#3B82F6' : '#888')) + ', ' + (((p.color.saturation||0) < -50 ? '#888' : (p.color.saturation||0) > 50 ? '#EC4899' : '#00D4FF')) + ')' }" @click="applyPreset(p)" :title="p.name">
              <span class="bg-bg/70 px-1 rounded text-[9px] text-text-primary">{{ p.name }}</span>
            </button>
          </div>
        </template>

        <!-- ======== AUDIO TAB (Advanced) ======== -->
        <template v-if="activeTab === 'audio'">
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Volume</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0" max="2" step="0.01" :value="selectedClip.volume" @input="(e) => updateClipField('volume', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="Math.round(selectedClip.volume * 100)" :step="1" min="0" max="200" @input="(e) => updateClipField('volume', Math.max(0, Math.min(2, (parseFloat(e.target.value) || 0) / 100)))" class="w-16 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>

          <hr class="border-border" />

          <!-- Speed (also relevant for audio) -->
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Playback Speed</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0.1" max="4" step="0.1" :value="selectedClip.speed" @input="(e) => updateClipField('speed', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="selectedClip.speed" :step="0.1" min="0.1" @input="(e) => updateClipField('speed', parseFloat(e.target.value) || 1)" class="w-16 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
            <div class="flex gap-1.5 mt-2 flex-wrap">
              <button v-for="sp in [0.25, 0.5, 1, 1.5, 2, 4]" :key="sp"
                class="px-2 py-0.5 rounded text-[10px] border transition-colors"
                :class="selectedClip.speed === sp ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary hover:border-accent/40'"
                @click="updateClipField('speed', sp)">{{ sp }}x</button>
            </div>
          </div>

          <hr class="border-border" />

          <!-- Waveform -->
          <div v-if="asset && asset.waveform && asset.waveform.length">
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Waveform</h4>
            <div class="h-16 flex items-center gap-px bg-bg/50 border border-border rounded p-1">
              <div v-for="(v, i) in asset.waveform" :key="i" class="flex-1 bg-accent/60 rounded-sm" :style="{ height: Math.max(2, Math.abs(v) * 100) + '%' }" />
            </div>
          </div>

          <hr class="border-border" />

          <!-- Fade controls -->
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Fade</h4>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="text-[10px] text-text-secondary">Fade In (s)</label>
                <input type="number" min="0" max="10" step="0.1" :value="selectedClip.transition?.duration || 0"
                  @input="(e) => updateClipField('transition', { type: 'fade', duration: parseFloat(e.target.value) || 0 })"
                  class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
              <div>
                <label class="text-[10px] text-text-secondary">Type</label>
                <select :value="selectedClip.transition?.type || 'none'"
                  @change="(e) => updateClipField('transition', { type: e.target.value, duration: selectedClip.transition?.duration || 0.5 })"
                  class="w-full bg-bg border border-border rounded px-2 py-1 text-xs">
                  <option value="none">None</option>
                  <option value="fade">Fade</option>
                  <option value="dissolve">Dissolve</option>
                  <option value="wipeleft">Wipe Left</option>
                  <option value="wiperight">Wipe Right</option>
                </select>
              </div>
            </div>
          </div>
        </template>

        <!-- ======== TEXT TAB ======== -->
        <template v-if="activeTab === 'text' && isText">
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Content</h4>
            <textarea :value="selectedClip.textProps.text" @input="(e) => setTextProp('text', e.target.value)" rows="3" class="w-full bg-bg border border-border rounded px-2 py-1.5 text-sm resize-none focus:border-accent outline-none"></textarea>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Size</h4>
            <input type="range" min="8" max="300" :value="selectedClip.textProps.fontSize" @input="(e) => setTextProp('fontSize', parseInt(e.target.value))" class="w-full accent-accent" />
            <div class="text-[10px] text-text-secondary text-right font-mono mt-1">{{ selectedClip.textProps.fontSize }}px</div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Color</h4>
            <div class="flex items-center gap-2">
              <input type="color" :value="selectedClip.textProps.color || '#FFFFFF'" @input="(e) => setTextProp('color', e.target.value)" class="w-10 h-8 rounded border border-border bg-transparent" />
              <input type="text" :value="selectedClip.textProps.color || '#FFFFFF'" @input="(e) => setTextProp('color', e.target.value)" class="flex-1 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Style</h4>
            <div class="flex items-center gap-2 flex-wrap">
              <button class="px-2 py-1 rounded text-xs border" :class="selectedClip.textProps.bold ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTextProp('bold', !selectedClip.textProps.bold)"><b>B</b></button>
              <button class="px-2 py-1 rounded text-xs border italic" :class="selectedClip.textProps.italic ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTextProp('italic', !selectedClip.textProps.italic)">I</button>
              <button class="px-2 py-1 rounded text-xs border underline" :class="selectedClip.textProps.underline ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTextProp('underline', !selectedClip.textProps.underline)">U</button>
              <button v-for="a in ['left','center','right']" :key="a" class="px-2 py-1 rounded text-xs border" :class="selectedClip.textProps.align === a ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTextProp('align', a)">{{ a }}</button>
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Stroke</h4>
            <div class="grid grid-cols-2 gap-2">
              <div>
                <label class="text-[10px] text-text-secondary">Width</label>
                <input type="number" min="0" max="20" :value="selectedClip.textProps.strokeWidth" @input="(e) => setTextProp('strokeWidth', parseInt(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
              <div>
                <label class="text-[10px] text-text-secondary">Color</label>
                <input type="color" :value="selectedClip.textProps.strokeColor || '#000000'" @input="(e) => setTextProp('strokeColor', e.target.value)" class="w-full h-7 rounded border border-border bg-transparent" />
              </div>
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Shadow</h4>
            <div class="grid grid-cols-3 gap-2">
              <div>
                <label class="text-[10px] text-text-secondary">Blur</label>
                <input type="number" min="0" max="50" :value="selectedClip.textProps.shadowBlur" @input="(e) => setTextProp('shadowBlur', parseInt(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
              <div>
                <label class="text-[10px] text-text-secondary">X</label>
                <input type="number" :value="selectedClip.textProps.shadowOffsetX" @input="(e) => setTextProp('shadowOffsetX', parseInt(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
              <div>
                <label class="text-[10px] text-text-secondary">Y</label>
                <input type="number" :value="selectedClip.textProps.shadowOffsetY" @input="(e) => setTextProp('shadowOffsetY', parseInt(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
              </div>
            </div>
          </div>
        </template>

      </div>
    </div>
  </div>
</template>
