<script setup>
import { computed } from 'vue'
import { useDesignStore, getNodeType } from '../../stores/designStore'
import { NODE_WIDTH, NODE_HEIGHT } from '../../lib/designConstants'

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
    maxX = Math.max(maxX, node.x + NODE_WIDTH)
    minY = Math.min(minY, node.y)
    maxY = Math.max(maxY, node.y + NODE_HEIGHT)
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
  Color: '#10B981',
  Effects: '#EC4899',
  Mask: '#14B8A6',
  Output: '#F472B6',
  Time: '#60A5FA',
  Utility: '#94A3B8',
}
</script>

<template>
  <div class="bg-cr-void border border-cr-line rounded-none overflow-hidden" :style="{ width: width + 'px', height: height + 'px' }">
    <svg :width="width" :height="height" class="w-full h-full">
      <!-- Connections -->
      <line
        v-for="conn in designStore.connections"
        :key="conn.id"
        :x1="toMinimap(designStore.nodes.find(n => n.id === conn.fromNode)?.x + NODE_WIDTH || 0, designStore.nodes.find(n => n.id === conn.fromNode)?.y + 30 || 0).x"
        :y1="toMinimap(designStore.nodes.find(n => n.id === conn.fromNode)?.x + NODE_WIDTH || 0, designStore.nodes.find(n => n.id === conn.fromNode)?.y + 30 || 0).y"
        :x2="toMinimap(designStore.nodes.find(n => n.id === conn.toNode)?.x || 0, designStore.nodes.find(n => n.id === conn.toNode)?.y + 30 || 0).x"
        :y2="toMinimap(designStore.nodes.find(n => n.id === conn.toNode)?.x || 0, designStore.nodes.find(n => n.id === conn.toNode)?.y + 30 || 0).y"
        stroke="#2E2E36"
        stroke-width="1"
      />

      <!-- Nodes -->
      <rect
        v-for="node in designStore.nodes"
        :key="node.id"
        :x="toMinimap(node.x, node.y).x"
        :y="toMinimap(node.x, node.y).y"
        :width="NODE_WIDTH * scale"
        :height="NODE_HEIGHT * scale"
        :rx="2"
        :fill="designStore.selectedNodeIds.has(node.id) ? '#00D4FF' : (nodeColorMap[getNodeType(node.type)?.category] || '#26262D')"
        :opacity="designStore.selectedNodeIds.has(node.id) ? 0.9 : 0.75"
      />
    </svg>
  </div>
</template>
