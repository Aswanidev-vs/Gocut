<script setup>
import { ref, computed, watch } from 'vue'
import { useDesignStore, getNodeType, EASING_TYPES } from '../../stores/designStore'
import { Plus, Trash2, ChevronDown, ChevronRight, Sliders, Play, RotateCcw, Key } from 'lucide-vue-next'

const props = defineProps({ playheadTime: { type: Number, default: 0 } })
const designStore = useDesignStore()

const node = computed(() => designStore.selectedNode)
const type = computed(() => node.value ? getNodeType(node.value.type) : null)
const localParams = ref({})
const activeKeyframeParam = ref(null)
const expandedKeyframes = ref(new Set())

watch(node, (n) => {
  localParams.value = n ? JSON.parse(JSON.stringify(n.params)) : {}
  activeKeyframeParam.value = null
  expandedKeyframes.value = new Set()
}, { immediate: true })

function updateParam(paramId, value) {
  if (!node.value) return
  localParams.value[paramId] = value
  designStore.updateNodeParam(node.value.id, paramId, value)
}

function addKeyframeForParam(paramId) {
  if (!node.value) return
  const value = localParams.value[paramId] ?? 0
  designStore.addKeyframe(node.value.id, paramId, props.playheadTime, value, 'smooth')
  activeKeyframeParam.value = paramId
}

function hasKeyframeAtTime(paramId) {
  if (!node.value || !node.value.keyframes[paramId]) return false
  return node.value.keyframes[paramId].some(k => Math.abs(k.time - props.playheadTime) < 0.02)
}

function removeKeyframe(paramId, kfId) {
  if (!node.value) return
  designStore.removeKeyframe(node.value.id, paramId, kfId)
}

function updateKeyframeEasing(paramId, kfId, easing) {
  if (!node.value) return
  const kfs = node.value.keyframes[paramId]
  if (!kfs) return
  const kf = kfs.find(k => k.id === kfId)
  if (kf) kf.easing = easing
}

function toggleKeyframeExpand(paramId) {
  if (expandedKeyframes.value.has(paramId)) {
    expandedKeyframes.value.delete(paramId)
  } else {
    expandedKeyframes.value.add(paramId)
  }
}

const keyframedParams = computed(() => {
  if (!node.value) return {}
  return node.value.keyframes || {}
})
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden bg-[#111119] select-none">
    <!-- Header -->
    <div class="h-8 px-3 border-b border-border/80 flex items-center justify-between bg-[#141420]">
      <div class="flex items-center gap-1.5 font-bold text-[10px] uppercase tracking-wider text-text-primary">
        <Sliders :size="12" class="text-accent" />
        <span>Fusion Inspector</span>
      </div>
      <div v-if="node" class="flex items-center gap-1 text-[9px] font-mono">
        <button
          class="px-1.5 py-0.5 rounded transition-colors"
          :class="designStore.viewer1NodeId === node.id ? 'bg-cyan-500/20 text-cyan-400 font-bold border border-cyan-500/40' : 'text-text-secondary bg-[#1C1C28] hover:text-white'"
          @click="designStore.setViewer1(node.id)"
          title="Send to Viewer 1 (Key 1)"
        >
          [1]
        </button>
        <button
          class="px-1.5 py-0.5 rounded transition-colors"
          :class="designStore.viewer2NodeId === node.id ? 'bg-pink-500/20 text-pink-400 font-bold border border-pink-500/40' : 'text-text-secondary bg-[#1C1C28] hover:text-white'"
          @click="designStore.setViewer2(node.id)"
          title="Send to Viewer 2 (Key 2)"
        >
          [2]
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!node" class="flex-1 flex flex-col items-center justify-center p-4 text-center">
      <div class="w-10 h-10 rounded-xl bg-white/5 border border-border/60 flex items-center justify-center text-text-secondary/50 mb-2">
        <Sliders :size="18" />
      </div>
      <div class="text-[11px] font-semibold text-text-primary mb-1">No Node Selected</div>
      <p class="text-[10px] text-text-secondary max-w-[180px]">
        Click any tool node in the flow graph to tweak its properties & keyframes.
      </p>
    </div>

    <!-- Active Node Inspector -->
    <div v-else class="flex-1 overflow-y-auto p-3 space-y-3">
      <!-- Node Identity Banner -->
      <div class="p-2.5 rounded-lg bg-[#181826] border border-border/70 flex items-center justify-between">
        <div class="flex items-center gap-2 min-w-0">
          <div class="w-3 h-3 rounded-full flex-shrink-0 shadow-sm" :style="{ background: type?.color }" />
          <div class="min-w-0">
            <input
              v-model="node.label"
              class="text-[11px] font-bold text-text-primary bg-transparent outline-none border-b border-transparent focus:border-accent w-full"
              placeholder="Node Name"
            />
            <div class="text-[9px] text-text-secondary font-mono">{{ type?.category }} / {{ type?.label }}</div>
          </div>
        </div>
      </div>

      <!-- Params -->
      <div v-for="p in type?.params" :key="p.id" class="space-y-1">
        <div class="flex items-center justify-between">
          <label class="text-[11px] text-text-secondary">{{ p.label }}</label>
          <button
            class="p-1 rounded text-[9px] transition-colors"
            :class="hasKeyframeAtTime(p.id) ? 'text-amber-400 bg-amber-500/10' : 'text-text-secondary/40 hover:text-amber-400 hover:bg-amber-500/10'"
            @click="addKeyframeForParam(p.id)"
            title="Add keyframe at current time"
          >
            ◆
          </button>
        </div>

        <!-- Text -->
        <input v-if="p.type === 'text'"
          :value="localParams[p.id]"
          @input="updateParam(p.id, $event.target.value)"
          class="w-full bg-bg border border-border rounded px-2 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent"
        />

        <!-- Number -->
        <div v-else-if="p.type === 'number'" class="flex items-center gap-2">
          <input
            type="range"
            :min="p.min ?? 0"
            :max="p.max ?? (p.id === 'opacity' ? 1 : p.id === 'scaleX' || p.id === 'scaleY' ? 10 : 1000)"
            :step="p.step ?? 0.01"
            :value="localParams[p.id] ?? p.def ?? 0"
            @input="updateParam(p.id, parseFloat($event.target.value))"
            class="flex-1 h-1 accent-accent"
          />
          <div class="text-[10px] text-text-secondary font-mono w-12 text-right">
            {{ typeof localParams[p.id] === 'number' ? localParams[p.id].toFixed(2) : localParams[p.id] }}{{ p.suffix || '' }}
          </div>
        </div>

        <!-- Color -->
        <input v-else-if="p.type === 'color'"
          :value="localParams[p.id] || '#FFFFFF'"
          @input="updateParam(p.id, $event.target.value)"
          type="color"
          class="w-full h-6 bg-bg border border-border rounded cursor-pointer"
        />

        <!-- Select -->
        <select v-else-if="p.type === 'select'"
          :value="localParams[p.id] ?? p.def ?? ''"
          @change="updateParam(p.id, $event.target.value)"
          class="w-full bg-bg border border-border rounded px-2 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent"
        >
          <option v-for="opt in p.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>

        <!-- Toggle -->
        <button v-else-if="p.type === 'toggle'"
          class="px-2 py-1 rounded text-[10px] border transition-colors"
          :class="localParams[p.id] ? 'bg-accent/10 text-accent border-accent/30' : 'bg-bg text-text-secondary border-border'"
          @click="updateParam(p.id, !localParams[p.id])"
        >
          {{ localParams[p.id] ? 'ON' : 'OFF' }}
        </button>

        <!-- Keyframe list for this param -->
        <div v-if="keyframedParams[p.id]?.length" class="mt-1">
          <button
            class="flex items-center gap-1 text-[9px] text-amber-500/70 hover:text-amber-400 transition-colors"
            @click="toggleKeyframeExpand(p.id)"
          >
            <ChevronRight v-if="!expandedKeyframes.has(p.id)" :size="10" />
            <ChevronDown v-else :size="10" />
            {{ keyframedParams[p.id].length }} keyframe(s)
          </button>
          <div v-if="expandedKeyframes.has(p.id)" class="mt-1 space-y-1 pl-2">
            <div
              v-for="kf in keyframedParams[p.id]"
              :key="kf.id"
              class="flex items-center gap-1.5 text-[9px] bg-bg/60 rounded px-1.5 py-1"
            >
              <span class="text-text-secondary font-mono w-8">{{ kf.time.toFixed(1) }}s</span>
              <span class="text-text-primary font-mono flex-1">{{ typeof kf.value === 'number' ? kf.value.toFixed(2) : kf.value }}</span>
              <select
                :value="kf.easing"
                @change="updateKeyframeEasing(p.id, kf.id, $event.target.value)"
                class="bg-transparent text-[8px] text-text-secondary border border-border/60 rounded px-1 py-0.5 outline-none"
              >
                <option v-for="e in EASING_TYPES" :key="e.id" :value="e.id">{{ e.label }}</option>
              </select>
              <button
                class="text-text-secondary/40 hover:text-red-400 transition-colors"
                @click="removeKeyframe(p.id, kf.id)"
              >
                <Trash2 :size="9" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Connections info -->
      <div class="pt-2 border-t border-border space-y-1">
        <div class="text-[10px] text-text-secondary uppercase tracking-wider">Connections</div>
        <div v-if="type?.inputs.length" class="text-[10px] text-text-secondary">
          Inputs: {{ type.inputs.join(', ') }}
          <span v-if="!designStore.connections.find(c => c.toNode === node.id)" class="text-text-secondary/50"> (none connected)</span>
        </div>
        <div v-if="type?.outputs.length" class="text-[10px] text-text-secondary">
          Outputs: {{ type.outputs.join(', ') }}
          <span v-if="!designStore.connections.find(c => c.fromNode === node.id)" class="text-text-secondary/50"> (not connected)</span>
        </div>
      </div>
    </div>
  </div>
</template>
