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
      class="w-full bg-bg border border-border rounded px-2.5 py-1.5 text-[11px] text-text-primary outline-none focus:border-accent placeholder:text-text-secondary/60"
    />
    <div v-for="(items, cat) in categories" :key="cat" class="space-y-1">
      <div class="text-[10px] text-text-secondary uppercase tracking-wider px-1 py-1">{{ cat }} ({{ items.length }})</div>
      <button
        v-for="item in items"
        :key="item.type"
        class="w-full flex items-center gap-2 px-2.5 py-2 rounded text-[11px] text-left transition-colors hover:bg-accent/5 border border-border/50 hover:border-accent/30"
        @click="addToGraph(item.type)"
      >
        <div class="w-2.5 h-2.5 rounded-full flex-shrink-0" :style="{ background: item.col }" />
        <div class="flex-1 min-w-0">
          <div class="text-text-primary truncate">{{ item.label }}</div>
          <div class="text-[9px] text-text-secondary">Inputs: {{ item.in.length }} · Outputs: {{ item.out.length }}</div>
        </div>
        <Plus :size="10" class="text-text-secondary flex-shrink-0" />
      </button>
    </div>
    <div v-if="Object.keys(categories).length === 0" class="text-[10px] text-text-secondary text-center py-4">
      No nodes match "{{ search }}"
    </div>
  </div>
</template>