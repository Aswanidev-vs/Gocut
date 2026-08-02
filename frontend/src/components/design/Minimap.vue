<script setup>
import { computed } from 'vue'
import { useDesignStore, getNodeType } from '../../stores/designStore'

const props = defineProps({
  width: { type: Number, default: 160 },
  height: { type: Number, default: 100 },
})

const designStore = useDesignStore()

const bounds = computed(() => {
  if (designStore.nodes.length === 0) return { minX: 0, maxX: 1920, minY: 0, maxY: 1080 }
  let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity
  for (const node of designStore.nodes) {
    minX = Math.min(minX, node.x)
    maxX = Math.max(maxX, node.x + 160)
    minY = Math.min(minY, node.y)
    maxY = Math.max(maxY, node.y + 60)
  }
  const pad = 200
  return { minX: minX - pad, maxX: maxX + pad, minY: minY - pad, maxY: maxY + pad }
})

const scale = computed(() => {
  const bw = bounds.value.maxX - bounds.value.minX
  const bh = bounds.value.maxY - bounds.value.minY
  return Math.min(props.width / bw, props.height / bh)
})

function toMinimap(nx, ny) {
  const b = bounds.value
  return {
    x: (nx - b.minX) * scale.value,
    y: (ny - b.minY) * scale.value,
  }
}

const nodeColorMap = {
  Sources: '#F59E0B',
  Transform: '#8B5CF6',
  Composite: '#10B981',
  Effects: '#EC4899',
  Math: '#A78BFA',
  Output: '#F472B6',
}
</script>

<template>
  <div class="bg-[#0a0a0a] border border-[#3a3a4a] rounded-lg overflow-hidden" :style="{ width: width + 'px', height: height + 'px' }">
    <svg :width="width" :height="height" class="w-full h-full">
      <!-- Connections -->
      <line
        v-for="conn in designStore.connections"
        :key="conn.id"
        :x1="toMinimap(designStore.nodes.find(n => n.id === conn.fromNode)?.x + 160 || 0, designStore.nodes.find(n => n.id === conn.fromNode)?.y + 30 || 0).x"
        :y1="toMinimap(designStore.nodes.find(n => n.id === conn.fromNode)?.x + 160 || 0, designStore.nodes.find(n => n.id === conn.fromNode)?.y + 30 || 0).y"
        :x2="toMinimap(designStore.nodes.find(n => n.id === conn.toNode)?.x || 0, designStore.nodes.find(n => n.id === conn.toNode)?.y + 30 || 0).x"
        :y2="toMinimap(designStore.nodes.find(n => n.id === conn.toNode)?.x || 0, designStore.nodes.find(n => n.id === conn.toNode)?.y + 30 || 0).y"
        stroke="#3a3a4a"
        stroke-width="1"
      />

      <!-- Nodes -->
      <rect
        v-for="node in designStore.nodes"
        :key="node.id"
        :x="toMinimap(node.x, node.y).x"
        :y="toMinimap(node.y, node.y).y"
        :width="160 * scale"
        :height="60 * scale"
        :rx="2"
        :fill="designStore.selectedNodeIds.has(node.id) ? '#00D4FF' : (nodeColorMap[getNodeType(node.type)?.category] || '#4a4a4a')"
        :opacity="designStore.selectedNodeIds.has(node.id) ? 0.9 : 0.6"
      />
    </svg>
  </div>
</template>
