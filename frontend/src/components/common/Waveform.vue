<script setup>
import { computed } from 'vue'

const props = defineProps({
  // 500-sample []float32 amplitude array (values ~0..1, signed)
  samples: { type: Array, default: null },
  color: { type: String, default: '#00D4FF' },
})

const W = 500
const H = 100

// Build a single closed SVG path: top envelope left→right, then mirrored
// bottom envelope right→left, so it fills into a symmetric waveform shape.
const path = computed(() => {
  const wf = props.samples
  if (!Array.isArray(wf) || wf.length === 0) return ''
  const N = wf.length
  const mid = H / 2
  const half = H / 2 - 2
  let d = ''
  for (let i = 0; i < N; i++) {
    const x = (i / (N - 1)) * W
    const amp = Math.min(1, Math.abs(Number(wf[i]) || 0))
    const y = mid - amp * half
    d += (i === 0 ? 'M' : 'L') + x.toFixed(2) + ' ' + y.toFixed(2) + ' '
  }
  for (let i = N - 1; i >= 0; i--) {
    const x = (i / (N - 1)) * W
    const amp = Math.min(1, Math.abs(Number(wf[i]) || 0))
    const y = mid + amp * half
    d += 'L' + x.toFixed(2) + ' ' + y.toFixed(2) + ' '
  }
  return d + 'Z'
})
</script>

<template>
  <svg
    :viewBox="`0 0 ${W} ${H}`"
    preserveAspectRatio="none"
    class="w-full h-full block"
  >
    <path
      v-if="path"
      :d="path"
      :fill="color"
      fill-opacity="0.5"
      :stroke="color"
      stroke-opacity="0.9"
      stroke-width="1"
      vector-effect="non-scaling-stroke"
    />
    <line
      v-else
      :x1="0" :y1="H / 2" :x2="W" :y2="H / 2"
      :stroke="color" stroke-opacity="0.3" stroke-width="1"
      stroke-dasharray="4 4" vector-effect="non-scaling-stroke"
    />
  </svg>
</template>
