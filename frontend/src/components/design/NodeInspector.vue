<script setup>
// Parameter panel for the selected node. In Simple mode the workspace supplies
// the header, so `embedded` drops this component's own chrome.
import { ref, computed, watch } from 'vue'
import { useDesignStore, getNodeType, EASING_TYPES } from '../../stores/designStore'
import { ChevronDown, ChevronRight, Trash2 } from 'lucide-vue-next'

const props = defineProps({
  playheadTime: { type: Number, default: 0 },
  embedded: { type: Boolean, default: false },
})

const designStore = useDesignStore()

const node = computed(() => designStore.selectedNode)
const type = computed(() => (node.value ? getNodeType(node.value.type) : null))
const localParams = ref({})
const expandedKeyframes = ref(new Set())

watch(node, (n) => {
  localParams.value = n ? JSON.parse(JSON.stringify(n.params)) : {}
  expandedKeyframes.value = new Set()
}, { immediate: true })

const animatedParams = computed(() => (node.value && node.value.keyframes) || {})

function updateParam(paramId, value) {
  if (!node.value) return
  localParams.value[paramId] = value
  designStore.updateNodeParam(node.value.id, paramId, value)
}

function renameNode(event) {
  if (!node.value) return
  node.value.label = event.target.value || node.value.label
}

function addKeyframeForParam(paramId) {
  if (!node.value) return
  const value = localParams.value[paramId] ?? 0
  designStore.addKeyframe(node.value.id, paramId, props.playheadTime, value, 'smooth')
  const next = new Set(expandedKeyframes.value)
  next.add(paramId)
  expandedKeyframes.value = next
}

function hasKeyframeAtTime(paramId) {
  const list = animatedParams.value[paramId]
  if (!list) return false
  return list.some(k => Math.abs(k.time - props.playheadTime) < 0.02)
}

function removeKeyframe(paramId, kfId) {
  if (!node.value) return
  designStore.removeKeyframe(node.value.id, paramId, kfId)
}

function updateKeyframeEasing(paramId, kfId, easing) {
  if (!node.value) return
  const kf = (animatedParams.value[paramId] || []).find(k => k.id === kfId)
  if (kf) kf.easing = easing
}

function toggleKeyframeExpand(paramId) {
  const next = new Set(expandedKeyframes.value)
  if (next.has(paramId)) next.delete(paramId)
  else next.add(paramId)
  expandedKeyframes.value = next
}

// Live value at the playhead: amber whenever the parameter is animated, which
// is the one thing worth shouting about in this panel.
function liveValue(paramId) {
  if (!node.value) return ''
  const v = designStore.getParamValue(node.value.id, paramId, props.playheadTime)
  return typeof v === 'number' ? Number(v.toFixed(3)) : v
}

function isAnimated(paramId) {
  return (animatedParams.value[paramId] || []).length > 0
}
</script>

<template>
  <div class="flex flex-col h-full bg-cr-panel select-none min-h-0">
    <div v-if="!embedded" class="h-8 flex-shrink-0 px-3 border-b border-cr-line flex items-center justify-between bg-cr-raise">
      <span class="text-[9px] uppercase tracking-[0.14em] text-ink-faint">Parameters</span>
      <div v-if="node" class="flex items-center gap-px">
        <button
          class="px-1.5 py-0.5 text-[9px] font-jetbrains-mono border transition-colors"
          :class="designStore.viewer1NodeId === node.id ? 'border-accent/40 bg-accent/10 text-accent' : 'border-cr-line text-ink-faint hover:text-ink'"
          title="Send to viewer A"
          @click="designStore.setViewer1(node.id)"
        >A</button>
        <button
          class="px-1.5 py-0.5 text-[9px] font-jetbrains-mono border transition-colors"
          :class="designStore.viewer2NodeId === node.id ? 'border-signal/40 bg-signal/10 text-signal' : 'border-cr-line text-ink-faint hover:text-ink'"
          title="Send to viewer B"
          @click="designStore.setViewer2(node.id)"
        >B</button>
      </div>
    </div>

    <div v-if="!node" class="flex-1 flex items-center justify-center px-4 text-center">
      <p class="text-[10px] text-ink-faint leading-relaxed">
        No node selected.<br />Pick a move on the left, then click a node in Flow to tune it.
      </p>
    </div>

    <div v-else class="flex-1 overflow-y-auto">
      <!-- Identity -->
      <div class="px-3 py-2 border-b border-cr-line flex items-center gap-2">
        <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :style="{ background: type?.color }" />
        <input
          :value="node.label"
          class="flex-1 min-w-0 bg-transparent text-[11px] text-ink outline-none border-b border-transparent focus:border-accent"
          @change="renameNode"
        />
        <span class="text-[9px] font-jetbrains-mono text-ink-faint flex-shrink-0">{{ type?.label }}</span>
      </div>

      <div v-if="!type?.params?.length" class="px-3 py-3 text-[10px] text-ink-faint">
        This node has nothing to tune.
      </div>

      <!-- Parameters -->
      <div v-for="p in type?.params" :key="p.id" class="px-3 py-2 border-b border-cr-line-soft">
        <div class="flex items-center justify-between gap-2 mb-1">
          <label class="text-[10px] text-ink-dim truncate">{{ p.label }}</label>
          <div class="flex items-center gap-1.5 flex-shrink-0">
            <span
              class="text-[9px] font-jetbrains-mono tabular-nums"
              :class="isAnimated(p.id) ? 'text-signal' : 'text-ink-faint'"
            >{{ liveValue(p.id) }}</span>
            <button
              class="text-[10px] leading-none transition-colors"
              :class="hasKeyframeAtTime(p.id) ? 'text-signal' : 'text-ink-faint hover:text-signal'"
              :title="hasKeyframeAtTime(p.id) ? 'Keyframe already set at the playhead' : 'Set a keyframe at the playhead'"
              @click="addKeyframeForParam(p.id)"
            >◆</button>
          </div>
        </div>

        <input
          v-if="p.type === 'number'"
          type="number"
          :min="p.min"
          :max="p.max"
          :step="p.step || 0.01"
          :value="localParams[p.id] ?? p.def ?? 0"
          class="w-full bg-cr-void border border-cr-line px-2 py-1 text-[11px] font-jetbrains-mono tabular-nums text-ink outline-none focus:border-accent"
          @input="updateParam(p.id, Number($event.target.value))"
        />

        <input
          v-else-if="p.type === 'color'"
          type="color"
          :value="localParams[p.id] || '#000000'"
          class="w-full h-6 bg-cr-void border border-cr-line cursor-pointer"
          @input="updateParam(p.id, $event.target.value)"
        />

        <select
          v-else-if="p.type === 'select'"
          :value="localParams[p.id] ?? p.def ?? ''"
          class="w-full bg-cr-void border border-cr-line px-2 py-1 text-[11px] text-ink outline-none focus:border-accent"
          @change="updateParam(p.id, $event.target.value)"
        >
          <option v-for="opt in p.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>

        <button
          v-else-if="p.type === 'toggle'"
          class="px-2 py-1 text-[10px] uppercase tracking-[0.14em] border transition-colors"
          :class="localParams[p.id] ? 'bg-accent/10 text-accent border-accent/40' : 'bg-cr-void text-ink-dim border-cr-line'"
          @click="updateParam(p.id, !localParams[p.id])"
        >{{ localParams[p.id] ? 'On' : 'Off' }}</button>

        <!-- Text, font family and asset ids all edit as free text -->
        <input
          v-else
          type="text"
          :value="localParams[p.id] ?? p.def ?? ''"
          class="w-full bg-cr-void border border-cr-line px-2 py-1 text-[11px] text-ink outline-none focus:border-accent"
          @input="updateParam(p.id, $event.target.value)"
        />

        <!-- Keyframes for this parameter -->
        <div v-if="animatedParams[p.id]?.length" class="mt-1.5">
          <button
            class="flex items-center gap-1 text-[9px] uppercase tracking-[0.14em] text-signal/80 hover:text-signal transition-colors"
            @click="toggleKeyframeExpand(p.id)"
          >
            <ChevronDown v-if="expandedKeyframes.has(p.id)" :size="10" />
            <ChevronRight v-else :size="10" />
            {{ animatedParams[p.id].length }} keyframes
          </button>
          <div v-if="expandedKeyframes.has(p.id)" class="mt-1 space-y-px">
            <div
              v-for="kf in animatedParams[p.id]"
              :key="kf.id"
              class="flex items-center gap-1.5 text-[9px] bg-cr-void border-l-2 border-signal/60 px-1.5 py-1"
            >
              <span class="font-jetbrains-mono tabular-nums text-ink-dim w-10">{{ kf.time.toFixed(2) }}s</span>
              <span class="font-jetbrains-mono tabular-nums text-ink flex-1 truncate">
                {{ typeof kf.value === 'number' ? Number(kf.value.toFixed(3)) : kf.value }}
              </span>
              <select
                :value="kf.easing"
                class="bg-transparent text-[9px] text-ink-dim border border-cr-line outline-none px-1 py-0.5"
                @change="updateKeyframeEasing(p.id, kf.id, $event.target.value)"
              >
                <option v-for="e in EASING_TYPES" :key="e.id" :value="e.id">{{ e.label }}</option>
              </select>
              <button class="text-ink-faint hover:text-red-400 transition-colors" @click="removeKeyframe(p.id, kf.id)">
                <Trash2 :size="9" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Wiring -->
      <div v-if="type?.inputs?.length || type?.outputs?.length" class="px-3 py-2">
        <div class="text-[9px] uppercase tracking-[0.14em] text-ink-faint mb-1">Wiring</div>
        <div v-if="type.inputs.length" class="text-[10px] text-ink-dim">
          In: <span class="font-jetbrains-mono">{{ type.inputs.join(', ') }}</span>
          <span v-if="!designStore.connections.find(c => c.toNode === node.id)" class="text-ink-faint"> · nothing connected</span>
        </div>
        <div v-if="type.outputs.length" class="text-[10px] text-ink-dim">
          Out: <span class="font-jetbrains-mono">{{ type.outputs.join(', ') }}</span>
          <span v-if="!designStore.connections.find(c => c.fromNode === node.id)" class="text-ink-faint"> · not wired onward</span>
        </div>
      </div>
    </div>
  </div>
</template>