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
    <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="emit('close')" />

    <!-- Palette -->
    <div class="relative w-[340px] bg-[#1E1E2E] border border-[#3a3a4a] rounded-xl shadow-2xl shadow-black/60 overflow-hidden" @mousedown.stop>
      <!-- Search input -->
      <div class="flex items-center gap-2 px-3 py-2.5 border-b border-[#3a3a4a]">
        <Search :size="14" class="text-text-secondary flex-shrink-0" />
        <input
          ref="inputRef"
          v-model="search"
          type="text"
          placeholder="Search nodes… (e.g. blur, merge, text)"
          class="flex-1 bg-transparent text-[13px] text-text-primary outline-none placeholder:text-text-secondary/50"
          @keydown="onKeyDown"
        />
        <kbd class="text-[9px] text-text-secondary/60 bg-bg/60 px-1.5 py-0.5 rounded border border-border/60 font-mono">ESC</kbd>
      </div>

      <!-- Results -->
      <div class="max-h-[280px] overflow-y-auto py-1">
        <div v-if="filteredNodes.length === 0" class="px-4 py-6 text-center text-[11px] text-text-secondary/60">
          No nodes match "{{ search }}"
        </div>
        <button
          v-for="(node, i) in filteredNodes"
          :key="node.type"
          class="w-full flex items-center gap-2.5 px-3 py-2 text-left transition-colors"
          :class="i === selectedIndex ? 'bg-accent/10 text-accent' : 'text-text-primary hover:bg-bg/60'"
          @click="selectNode(node)"
          @mouseenter="selectedIndex = i"
        >
          <div
            class="w-6 h-6 rounded flex items-center justify-center flex-shrink-0 text-[10px] font-bold"
            :style="{ backgroundColor: node.color + '20', color: node.color }"
          >
            {{ node.label.charAt(0) }}
          </div>
          <div class="flex-1 min-w-0">
            <div class="text-[12px] font-medium truncate">{{ node.label }}</div>
            <div class="text-[9px] text-text-secondary/60 truncate">{{ node.category }}</div>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <span v-if="node.inputs.length" class="text-[8px] text-text-secondary/40 bg-bg/60 px-1 rounded">{{ node.inputs.length }} in</span>
            <span v-if="node.outputs.length" class="text-[8px] text-text-secondary/40 bg-bg/60 px-1 rounded">{{ node.outputs.length }} out</span>
          </div>
        </button>
      </div>

      <!-- Footer hint -->
      <div class="px-3 py-1.5 border-t border-[#3a3a4a] flex items-center gap-3 text-[9px] text-text-secondary/50">
        <span><kbd class="bg-bg/60 px-1 rounded border border-border/60">↑↓</kbd> navigate</span>
        <span><kbd class="bg-bg/60 px-1 rounded border border-border/60">Enter</kbd> select</span>
        <span><kbd class="bg-bg/60 px-1 rounded border border-border/60">Esc</kbd> close</span>
      </div>
    </div>
  </div>
</template>
