<script setup>
import { ref, computed } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { useUiStore } from '../../stores/uiStore'
import {
  Sun, Droplets, Sparkles, Palette, Contrast, Eye, RotateCcw,
  ChevronDown, ChevronRight, Wand2, Plus
} from 'lucide-vue-next'

const designStore = useDesignStore()
const uiStore = useUiStore()

const emit = defineEmits(['applied'])

// Effect categories with friendly names
const effectGroups = [
  {
    id: 'enhance',
    label: 'Enhance',
    icon: Sun,
    effects: [
      { id: 'glow', label: 'Glow', icon: Sparkles, desc: 'Add a soft glow halo', nodeType: 'glow', params: { intensity: 1.2, color: '#00D4FF', radius: 15 } },
      { id: 'blur', label: 'Blur', icon: Droplets, desc: 'Soften the image', nodeType: 'blur', params: { radius: 8 } },
      { id: 'shadow', label: 'Drop Shadow', icon: Eye, desc: 'Add depth with a shadow', nodeType: 'shadow', params: { color: '#000000', blur: 12, offsetX: 0, offsetY: 6 } },
    ],
  },
  {
    id: 'color',
    label: 'Color',
    icon: Palette,
    effects: [
      { id: 'colorCorrect', label: 'Color Adjust', icon: Contrast, desc: 'Brightness, contrast, saturation', nodeType: 'colorCorrect', params: { brightness: 0.1, contrast: 0.15, saturation: 0.2, hue: 0 } },
      { id: 'chromaKey', label: 'Green Screen', icon: Wand2, desc: 'Remove a background color', nodeType: 'chromaKey', params: { keyColor: '#00FF00', similarity: 0.4, smoothness: 0.1 } },
    ],
  },
]

const expandedGroup = ref('enhance')
const appliedEffects = ref([]) // { id, nodeType, label, params }

function applyEffect(effect) {
  // Create a source + effect + output chain if graph is empty
  if (designStore.nodes.length === 0) {
    const src = designStore.addNode('text', { x: 100, y: 200, label: 'Source', params: { text: 'Your Text', fontSize: 64, color: '#FFFFFF' } })
    const fx = designStore.addNode(effect.nodeType, { x: 350, y: 200, label: effect.label, params: { ...effect.params } })
    designStore.addNode('output', { x: 600, y: 200, label: 'Output' })
    if (src && fx) designStore.addConnection(src.id, 'out', fx.id, 'in')
  } else {
    // Insert effect after the currently selected node
    const fx = designStore.addNode(effect.nodeType, { x: 350, y: 250, label: effect.label, params: { ...effect.params } })
  }
  appliedEffects.value.push({ id: crypto.randomUUID(), nodeType: effect.nodeType, label: effect.label, params: { ...effect.params } })
  uiStore.addToast(`Added ${effect.label} effect`, 'success', 1500)
  emit('applied')
}

function removeEffect(idx) {
  appliedEffects.value.splice(idx, 1)
}

function updateEffectParam(idx, key, val) {
  appliedEffects.value[idx].params[key] = val
  // Sync to the actual node in the graph
  const fx = appliedEffects.value[idx]
  const node = designStore.nodes.find(n => n.label === fx.label && n.type === fx.nodeType)
  if (node) designStore.updateNodeParam(node.id, key, val)
}
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- Header -->
    <div class="px-3 py-2.5 border-b border-border bg-panel/50">
      <div class="flex items-center gap-2">
        <Wand2 :size="14" class="text-accent" />
        <div>
          <div class="text-sm font-bold text-text-primary">Quick Effects</div>
          <div class="text-[10px] text-text-secondary">Click an effect to add it. Adjust with sliders.</div>
        </div>
      </div>
    </div>

    <div class="flex-1 overflow-y-auto p-3 space-y-3">
      <!-- Effect categories -->
      <div v-for="group in effectGroups" :key="group.id" class="rounded-lg border border-border/60 overflow-hidden">
        <!-- Group header -->
        <button
          class="w-full flex items-center gap-2 px-3 py-2 bg-panel/40 hover:bg-panel/70 transition-colors"
          @click="expandedGroup = expandedGroup === group.id ? null : group.id"
        >
          <component :is="group.icon" :size="14" class="text-accent" />
          <span class="text-[11px] font-bold text-text-primary uppercase tracking-wider">{{ group.label }}</span>
          <div class="flex-1" />
          <ChevronDown v-if="expandedGroup === group.id" :size="12" class="text-text-secondary" />
          <ChevronRight v-else :size="12" class="text-text-secondary" />
        </button>

        <!-- Effects list -->
        <div v-if="expandedGroup === group.id" class="p-2 space-y-1.5 bg-bg/30">
          <button
            v-for="effect in group.effects"
            :key="effect.id"
            class="w-full flex items-center gap-2.5 p-2.5 rounded-lg border border-border/40 hover:border-accent/40 hover:bg-accent/5 transition-all text-left group"
            @click="applyEffect(effect)"
          >
            <div class="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center flex-shrink-0">
              <component :is="effect.icon" :size="14" class="text-accent" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-[11px] font-semibold text-text-primary">{{ effect.label }}</div>
              <div class="text-[9px] text-text-secondary truncate">{{ effect.desc }}</div>
            </div>
            <Plus :size="12" class="text-text-secondary group-hover:text-accent flex-shrink-0" />
          </button>
        </div>
      </div>

      <!-- Applied effects with sliders -->
      <div v-if="appliedEffects.length" class="space-y-2 pt-2">
        <div class="text-[10px] text-text-secondary uppercase tracking-wider font-bold px-1">Applied Effects</div>
        <div
          v-for="(fx, idx) in appliedEffects"
          :key="fx.id"
          class="rounded-lg border border-border/60 bg-panel/40 p-3 space-y-2"
        >
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-semibold text-text-primary">{{ fx.label }}</span>
            <button
              class="p-1 rounded text-text-secondary hover:text-red-400 hover:bg-red-500/10 transition-colors"
              @click="removeEffect(idx)"
              title="Remove effect"
            >
              <RotateCcw :size="11" />
            </button>
          </div>
          <!-- Param sliders -->
          <div v-for="(val, key) in fx.params" :key="key" class="space-y-1">
            <div class="flex items-center justify-between text-[10px]">
              <span class="text-text-secondary capitalize">{{ key }}</span>
              <!-- Color value: show color picker -->
              <input
                v-if="typeof val === 'string' && val.startsWith('#')"
                type="color"
                :value="val"
                @input="(e) => updateEffectParam(idx, key, e.target.value)"
                class="w-6 h-4 rounded border border-border bg-transparent cursor-pointer"
              />
              <!-- Number value: show numeric readout -->
              <span v-else class="text-text-primary font-mono">{{ typeof val === 'number' ? val.toFixed(2) : val }}</span>
            </div>
            <!-- Range slider for numeric params -->
            <input
              v-if="typeof val === 'number'"
              type="range"
              min="0"
              :max="key === 'blur' || key === 'radius' ? 100 : key === 'intensity' ? 5 : key === 'similarity' || key === 'smoothness' ? 1 : 1"
              step="0.01"
              :value="val"
              @input="(e) => updateEffectParam(idx, key, parseFloat(e.target.value))"
              class="w-full h-1 accent-accent"
            />
          </div>
        </div>
      </div>

      <!-- Empty hint -->
      <div v-else class="text-center py-8 text-[11px] text-text-secondary/60">
        <Wand2 :size="24" class="mx-auto mb-2 opacity-40" />
        No effects applied yet.<br />Pick one above to get started.
      </div>
    </div>
  </div>
</template>