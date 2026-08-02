<script setup>
import { ref, computed, watch } from 'vue'
import { useDesignStore, getNodeType, EASING_TYPES } from '../../stores/designStore'
import { Plus, Trash2, ChevronDown, ChevronRight } from 'lucide-vue-next'

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
  designStore.addKeyframe(node.value.id, paramId, props.playheadTime, value, 'linear')
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
  <div class="flex flex-col h-full overflow-hidden">
    <div class="p-2 border-b border-border">
      <div class="text-[10px] text-text-secondary uppercase tracking-wider">Node Inspector</div>
    </div>
    <div v-if="!node" class="flex-1 flex items-center justify-center text-[11px] text-text-secondary">
      <div class="text-center">
        <div class="text-2xl mb-1">🔗</div>
        Select a node on the canvas
      </div>
    </div>
    <div v-else class="flex-1 overflow-y-auto p-2 space-y-2.5">
      <!-- Node identity -->
      <div class="flex items-center gap-2">
        <div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{ background: type?.color }" />
        <div>
          <div class="text-[12px] text-text-primary font-medium">{{ node.label }}</div>
          <div class="text-[10px] text-text-secondary">{{ type?.category }} · {{ type?.label }}</div>
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
