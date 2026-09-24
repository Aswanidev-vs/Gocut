<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { NODE_TYPES } from '../../stores/designStore'
import { Search, Box, Sparkles, Zap, Hash, ArrowRight } from 'lucide-vue-next'

const emit = defineEmits(['close', 'select'])

const search = ref('')
const selectedIndex = ref(0)
const inputRef = ref(null)

const categoryIcons = {
  Sources: Box,
  Transform: Zap,
  Composite: Sparkles,
  Effects: Zap,
  Math: Hash,
  Output: ArrowRight,
}

const allNodes = computed(() => {
  const list = []
  for (const [key, def] of Object.entries(NODE_TYPES)) {
    list.push({
      type: key,
      label: def.label,
      category: def.cat,
      color: def.col,
      inputs: def.in,
      outputs: def.out,
    })
  }
  return list
})

const filteredNodes = computed(() => {
  if (!search.value) return allNodes.value
  const q = search.value.toLowerCase()
  return allNodes.value.filter(n =>
    n.label.toLowerCase().includes(q) ||
    n.type.toLowerCase().includes(q) ||
    n.category.toLowerCase().includes(q)
  )
})

watch(filteredNodes, () => {
  selectedIndex.value = 0
})

function onKeyDown(e) {
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedIndex.value = Math.min(selectedIndex.value + 1, filteredNodes.value.length - 1)
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedIndex.value = Math.max(selectedIndex.value - 1, 0)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    if (filteredNodes.value[selectedIndex.value]) {
      selectNode(filteredNodes.value[selectedIndex.value])
    }
  } else if (e.key === 'Escape') {
    e.preventDefault()
    emit('close')
  }
}

function selectNode(node) {
  emit('select', node.type)
  emit('close')
}

onMounted(() => {
  nextTick(() => {
    inputRef.value?.focus()
  })
})
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-start justify-center pt-[15vh]" @mousedown.self="emit('close')">
    <!-- Backdrop -->
    <div class="absolute inset-0 bg-cr-void/70 backdrop-blur-sm" @click="emit('close')" />

    <!-- Palette -->
    <div class="relative w-[340px] bg-cr-panel border border-cr-line rounded-none overflow-hidden" @mousedown.stop>
      <!-- Search input -->
      <div class="flex items-center gap-2 px-3 py-2.5 border-b border-cr-line">
        <Search :size="14" class="text-ink-faint flex-shrink-0" />
        <input
          ref="inputRef"
          v-model="search"
          type="text"
          placeholder="Search nodes… (e.g. blur, merge, text)"
          class="flex-1 bg-transparent text-[13px] text-ink outline-none placeholder:text-ink-faint"
          @keydown="onKeyDown"
        />
        <kbd class="text-[9px] text-ink-faint bg-cr-raise px-1.5 py-0.5 rounded-none border border-cr-line font-jetbrains-mono">ESC</kbd>
      </div>

      <!-- Results -->
      <div class="max-h-[280px] overflow-y-auto py-1">
        <div v-if="filteredNodes.length === 0" class="px-4 py-6 text-center text-[11px] text-ink-faint">
          No nodes match "{{ search }}"
        </div>
        <button
          v-for="(node, i) in filteredNodes"
          :key="node.type"
          class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors border-l-2"
          :class="i === selectedIndex ? 'bg-accent/10 text-accent border-accent' : 'bg-cr-raise text-ink border-transparent hover:bg-white/5'"
          @click="selectNode(node)"
          @mouseenter="selectedIndex = i"
        >
          <div
            class="w-6 h-6 rounded-none flex items-center justify-center flex-shrink-0 text-[10px] font-bold"
            :style="{ backgroundColor: node.color + '20', color: node.color }"
          >
            {{ node.label.charAt(0) }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-[12px] font-medium truncate">{{ node.label }}</div>
            <div class="text-[9px] uppercase tracking-[0.14em] text-ink-faint truncate">{{ node.category }}</div>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <span v-if="node.inputs.length" class="text-[8px] text-ink-dim bg-cr-void px-1 rounded-none font-jetbrains-mono tabular-nums">{{ node.inputs.length }} in</span>
            <span v-if="node.outputs.length" class="text-[8px] text-ink-dim bg-cr-void px-1 rounded-none font-jetbrains-mono tabular-nums">{{ node.outputs.length }} out</span>
          </div>
        </button>
      </div>

      <!-- Footer hint -->
      <div class="px-3 py-1.5 border-t border-cr-line flex items-center gap-3 text-[9px] text-ink-faint">
        <span><kbd class="bg-cr-raise px-1 rounded-none border border-cr-line font-jetbrains-mono">↑↓</kbd> navigate</span>
        <span><kbd class="bg-cr-raise px-1 rounded-none border border-cr-line font-jetbrains-mono">Enter</kbd> select</span>
        <span><kbd class="bg-cr-raise px-1 rounded-none border border-cr-line font-jetbrains-mono">Esc</kbd> close</span>
      </div>
    </div>
  </div>
</template>
