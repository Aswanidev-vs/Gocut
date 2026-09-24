<script setup>
import { ref, computed } from 'vue'
import { useDesignStore, NODE_TYPES, getNodeType } from '../../stores/designStore'
import { Plus } from 'lucide-vue-next'

const designStore = useDesignStore()
const search = ref('')

const categories = computed(() => {
  const groups = {}
  for (const [type, def] of Object.entries(NODE_TYPES)) {
    if (search.value && !def.label.toLowerCase().includes(search.value.toLowerCase())) continue
    if (!groups[def.cat]) groups[def.cat] = []
    groups[def.cat].push({ type, ...def })
  }
  return groups
})

function addToGraph(type) {
  designStore.addNode(type)
}
</script>

<template>
  <div class="p-2 space-y-2">
    <input
      v-model="search"
      type="text"
      placeholder="Search nodes…"
      class="w-full bg-cr-void border border-cr-line px-2.5 py-1.5 text-[11px] text-ink outline-none focus:border-accent placeholder:text-ink-faint"
    />
    <div v-for="(items, cat) in categories" :key="cat" class="space-y-px">
      <div class="text-[9px] uppercase tracking-[0.14em] text-ink-faint px-0.5 py-1 flex items-center justify-between">
        <span>{{ cat }}</span>
        <span class="font-jetbrains-mono tabular-nums">{{ items.length }}</span>
      </div>
      <button
        v-for="item in items"
        :key="item.type"
        class="w-full flex items-center gap-2 px-2.5 py-1.5 text-[11px] text-left transition-colors border-l-2 border-transparent hover:border-accent hover:bg-white/5"
        @click="addToGraph(item.type)"
      >
        <div class="w-2 h-2 flex-shrink-0" :style="{ background: item.col }" />
        <div class="flex-1 min-w-0">
          <div class="text-ink truncate">{{ item.label }}</div>
          <div class="text-[9px] text-ink-faint font-jetbrains-mono">
            {{ item.in.length }} in · {{ item.out.length }} out
          </div>
        </div>
        <Plus :size="10" class="text-ink-faint flex-shrink-0" />
      </button>
    </div>
    <div v-if="Object.keys(categories).length === 0" class="text-[10px] text-ink-faint text-center py-4">
      No nodes match “{{ search }}”
    </div>
  </div>
</template>