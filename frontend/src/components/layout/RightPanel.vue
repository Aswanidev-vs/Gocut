<script setup>
import { computed } from 'vue'
import { useTimelineStore } from '../../stores/timelineStore'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import { ChevronRight, Trash2, Copy } from 'lucide-vue-next'

const timelineStore = useTimelineStore()
const projectStore = useProjectStore()
const uiStore = useUiStore()

const hasSelection = computed(() => timelineStore.selectedClips.length > 0)
const selectedClip = computed(() => timelineStore.selectedClips[0])
const asset = computed(() => selectedClip.value ? projectStore.getAsset(selectedClip.value.assetId) : null)
const isText = computed(() => !!selectedClip.value?.textProps)

const sections = computed(() => {
  if (!selectedClip.value) return []
  const list = [
    { id: 'transform', label: 'Transform' },
    { id: 'color',     label: 'Color' },
    { id: 'audio',     label: 'Audio' },
  ]
  if (isText.value) list.push({ id: 'text', label: 'Text' })
  return list
})

const activeTab = computed({
  get: () => uiStore.activeInspectorTab,
  set: (v) => uiStore.setActiveInspectorTab(v),
})

function setTransform(key, val) {
  if (!selectedClip.value) return
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
  timelineStore.updateClip(selectedClip.value.id, { [key]: val })
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
      <div class="px-3 py-2 border-b border-border flex items-center gap-2">
        <div class="flex-1 min-w-0">
          <div class="text-[10px] text-text-secondary uppercase tracking-wider">{{ asset?.type || (isText ? 'Text' : 'Clip') }}</div>
          <div class="text-sm text-text-primary truncate">{{ isText ? selectedClip.textProps.text : fileName(asset?.path || '') }}</div>
        </div>
        <button class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border" @click="duplicateSelected" title="Duplicate"><Copy :size="12" /></button>
        <button class="p-1.5 rounded text-text-secondary hover:text-red-400 hover:bg-border" @click="deleteSelected" title="Delete"><Trash2 :size="12" /></button>
      </div>
      <div class="flex items-center gap-0.5 px-2 py-1.5 border-b border-border">
        <button v-for="s in sections" :key="s.id"
          class="px-2 py-1 rounded text-[11px]"
          :class="activeTab === s.id ? 'bg-accent/10 text-accent' : 'text-text-secondary hover:text-text-primary hover:bg-border'"
          @click="activeTab = s.id">{{ s.label }}</button>
      </div>
      <div class="flex-1 overflow-y-auto p-3 space-y-4">
        <template v-if="activeTab === 'transform'">
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Position</h4>
            <div class="grid grid-cols-2 gap-2">
              <div><label class="text-[10px] text-text-secondary">X</label><input type="number" :value="selectedClip.transform.x" @input="(e) => setTransform('x', parseFloat(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" /></div>
              <div><label class="text-[10px] text-text-secondary">Y</label><input type="number" :value="selectedClip.transform.y" @input="(e) => setTransform('y', parseFloat(e.target.value) || 0)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" /></div>
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Scale</h4>
            <div class="grid grid-cols-2 gap-2">
              <div><label class="text-[10px] text-text-secondary">W</label><input type="number" step="0.1" :value="selectedClip.transform.scaleX" @input="(e) => setTransform('scaleX', parseFloat(e.target.value) || 1)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" /></div>
              <div><label class="text-[10px] text-text-secondary">H</label><input type="number" step="0.1" :value="selectedClip.transform.scaleY" @input="(e) => setTransform('scaleY', parseFloat(e.target.value) || 1)" class="w-full bg-bg border border-border rounded px-2 py-1 text-xs font-mono" /></div>
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Rotation</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="-180" max="180" step="1" :value="selectedClip.transform.rotation" @input="(e) => setTransform('rotation', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="selectedClip.transform.rotation" @input="(e) => setTransform('rotation', parseFloat(e.target.value) || 0)" class="w-16 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Flip</h4>
            <div class="flex items-center gap-2">
              <button class="px-2 py-1 rounded text-xs border" :class="selectedClip.transform.flipH ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTransform('flipH', !selectedClip.transform.flipH)">H</button>
              <button class="px-2 py-1 rounded text-xs border" :class="selectedClip.transform.flipV ? 'border-accent text-accent bg-accent/10' : 'border-border text-text-secondary'" @click="setTransform('flipV', !selectedClip.transform.flipV)">V</button>
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Speed</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0.1" max="4" step="0.1" :value="selectedClip.speed" @input="(e) => updateClipField('speed', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="selectedClip.speed" :step="0.1" min="0.1" @input="(e) => updateClipField('speed', parseFloat(e.target.value) || 1)" class="w-16 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Opacity</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0" max="1" step="0.01" :value="selectedClip.opacity" @input="(e) => updateClipField('opacity', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="Math.round(selectedClip.opacity * 100)" :step="1" min="0" max="100" @input="(e) => updateClipField('opacity', Math.max(0, Math.min(1, (parseFloat(e.target.value) || 0) / 100)))" class="w-16 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
        </template>
        <template v-if="activeTab === 'color'">
          <div v-for="k in [['brightness',-100,100],['contrast',-100,100],['saturation',-100,100],['hue',-180,180],['temp',-100,100],['tint',-100,100],['sharpness',0,100],['vignette',0,100],['grain',0,100]]" :key="k[0]">
            <div class="flex justify-between text-[10px] text-text-secondary capitalize"><span>{{ k[0] }}</span><span class="font-mono">{{ selectedClip.color[k[0]] }}</span></div>
            <input type="range" :min="k[1]" :max="k[2]" :value="selectedClip.color[k[0]]" @input="(e) => setColor(k[0], parseInt(e.target.value))" class="w-full accent-accent" />
          </div>
          <div>
            <div class="flex justify-between text-[10px] text-text-secondary"><span>Blur</span><span class="font-mono">{{ selectedClip.color.blur }}</span></div>
            <input type="range" min="0" max="20" :value="selectedClip.color.blur" @input="(e) => setColor('blur', parseInt(e.target.value))" class="w-full accent-accent" />
          </div>
          <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mt-3 mb-2">Filter Presets</h4>
          <div class="grid grid-cols-3 gap-1.5">
            <button v-for="p in filterPresets" :key="p.name" class="aspect-video rounded border border-border/60 hover:border-accent transition-colors flex items-center justify-center" :style="{ background: 'linear-gradient(135deg, ' + (((p.color.temp||0) > 0 ? '#F59E0B' : (p.color.temp||0) < 0 ? '#3B82F6' : '#888')) + ', ' + (((p.color.saturation||0) < -50 ? '#888' : (p.color.saturation||0) > 50 ? '#EC4899' : '#00D4FF')) + ')' }" @click="applyPreset(p)" :title="p.name">
              <span class="bg-bg/70 px-1 rounded text-[9px] text-text-primary">{{ p.name }}</span>
            </button>
          </div>
        </template>
        <template v-if="activeTab === 'audio'">
          <div>
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2">Volume</h4>
            <div class="flex items-center gap-2">
              <input type="range" min="0" max="2" step="0.01" :value="selectedClip.volume" @input="(e) => updateClipField('volume', parseFloat(e.target.value))" class="flex-1 accent-accent" />
              <input type="number" :value="Math.round(selectedClip.volume * 100)" :step="1" min="0" max="200" @input="(e) => updateClipField('volume', Math.max(0, Math.min(2, (parseFloat(e.target.value) || 0) / 100)))" class="w-16 bg-bg border border-border rounded px-2 py-1 text-xs font-mono" />
            </div>
          </div>
          <div v-if="asset && asset.waveform && asset.waveform.length">
            <h4 class="text-[10px] font-semibold text-text-secondary uppercase tracking-wider mb-2 mt-2">Waveform</h4>
            <div class="h-12 flex items-center gap-px bg-bg/50 border border-border rounded p-1">
              <div v-for="(v, i) in asset.waveform" :key="i" class="flex-1 bg-accent/60 rounded-sm" :style="{ height: Math.max(2, Math.abs(v) * 100) + '%' }" />
            </div>
          </div>
        </template>
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
        </template>
      </div>
    </div>
  </div>
</template>
