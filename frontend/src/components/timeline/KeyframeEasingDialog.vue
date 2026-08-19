<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useTimelineStore, getInterpolatedProperty } from '../../stores/timelineStore'
import { X, Activity, Check } from 'lucide-vue-next'

const props = defineProps({
  isOpen: Boolean,
})
const emit = defineEmits(['close'])

const timelineStore = useTimelineStore()
const canvasRef = ref(null)

const selectedClip = computed(() => timelineStore.selectedClips[0])
const properties = computed(() => {
  if (!selectedClip.value || !selectedClip.value.keyframes) return []
  const propsSet = new Set(selectedClip.value.keyframes.map(k => k.property))
  return Array.from(propsSet)
})

const selectedProperty = ref('')

watch(properties, (list) => {
  if (list.length && (!selectedProperty.value || !list.includes(selectedProperty.value))) {
    selectedProperty.value = list[0]
  }
}, { immediate: true })

const currentKeyframes = computed(() => {
  if (!selectedClip.value || !selectedClip.value.keyframes || !selectedProperty.value) return []
  return selectedClip.value.keyframes
    .filter(k => k.property === selectedProperty.value)
    .sort((a, b) => a.time - b.time)
})

const easingOptions = [
  { id: 'linear', label: 'Linear' },
  { id: 'easeIn', label: 'Ease In (Quadratic)' },
  { id: 'easeOut', label: 'Ease Out (Quadratic)' },
  { id: 'easeInOut', label: 'Ease In-Out (Smooth)' },
  { id: 'bounce', label: 'Bounce' },
  { id: 'elastic', label: 'Elastic' },
]

function setEasing(keyframeId, easingMode) {
  if (!selectedClip.value) return
  timelineStore.updateKeyframe(selectedClip.value.id, keyframeId, { easing: easingMode })
}

function drawCurve() {
  const canvas = canvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  const W = canvas.width
  const H = canvas.height
  ctx.clearRect(0, 0, W, H)

  // Background
  ctx.fillStyle = '#0F0F17'
  ctx.fillRect(0, 0, W, H)

  // Grid
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)'
  ctx.lineWidth = 1
  for (let i = 0; i <= 4; i++) {
    const y = (H / 4) * i
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(W, y); ctx.stroke()
    const x = (W / 4) * i
    ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, H); ctx.stroke()
  }

  const kfs = currentKeyframes.value
  if (kfs.length < 2) {
    ctx.fillStyle = 'rgba(255, 255, 255, 0.3)'
    ctx.font = '11px sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText('Add at least 2 keyframes to visualize curve', W / 2, H / 2)
    return
  }

  const minT = kfs[0].time
  const maxT = kfs[kfs.length - 1].time
  const tRange = Math.max(0.01, maxT - minT)

  const vals = kfs.map(k => parseFloat(k.value) || 0)
  let minV = Math.min(...vals)
  let maxV = Math.max(...vals)

  // Bounce/elastic overshoot past the keyframe values; sample the real
  // interpolated property (same engine the preview uses) to include the
  // extremes in the plot range.
  const rangeSteps = 100
  for (let i = 0; i <= rangeSteps; i++) {
    const t = minT + (i / rangeSteps) * tRange
    const v = getInterpolatedProperty(selectedClip.value, selectedProperty.value, t, minV)
    if (v < minV) minV = v
    if (v > maxV) maxV = v
  }
  const vRange = Math.max(0.001, maxV - minV)

  const pad = 24
  const plotW = W - pad * 2
  const plotH = H - pad * 2

  const toScreen = (t, val) => ({
    x: pad + ((t - minT) / tRange) * plotW,
    y: H - pad - ((val - minV) / vRange) * plotH,
  })

  // Draw smooth curve
  ctx.strokeStyle = '#00D4FF'
  ctx.lineWidth = 2
  ctx.beginPath()

  const steps = 100
  for (let i = 0; i <= steps; i++) {
    const t = minT + (i / steps) * tRange
    // Route through the shared interpolator so the drawn curve always
    // matches the preview engine (including bounce/elastic).
    const interpVal = getInterpolatedProperty(selectedClip.value, selectedProperty.value, t, minV)

    const pos = toScreen(t, interpVal)
    if (i === 0) ctx.moveTo(pos.x, pos.y)
    else ctx.lineTo(pos.x, pos.y)
  }
  ctx.stroke()

  // Draw points
  kfs.forEach((kf) => {
    const pos = toScreen(kf.time, parseFloat(kf.value) || 0)
    ctx.fillStyle = '#F59E0B'
    ctx.beginPath()
    ctx.arc(pos.x, pos.y, 4, 0, Math.PI * 2)
    ctx.fill()
  })
}

watch([currentKeyframes, selectedProperty], drawCurve, { deep: true })
onMounted(drawCurve)
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-bg/70 backdrop-blur-sm" @click.self="emit('close')">
    <div class="w-[580px] bg-panel border border-border rounded-lg shadow-2xl overflow-hidden flex flex-col">
      <!-- Header -->
      <div class="px-4 py-3 border-b border-border flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Activity :size="16" class="text-accent" />
          <h3 class="text-sm font-semibold text-text-primary">Keyframe Easing Graph Editor</h3>
        </div>
        <button class="p-1 rounded hover:bg-border text-text-secondary" @click="emit('close')"><X :size="14" /></button>
      </div>

      <div class="p-4 space-y-4">
        <!-- Property Selector -->
        <div v-if="properties.length" class="flex items-center gap-2">
          <label class="text-xs text-text-secondary uppercase tracking-wider font-semibold">Property:</label>
          <select v-model="selectedProperty" class="bg-bg border border-border rounded px-2 py-1 text-xs text-text-primary outline-none focus:border-accent">
            <option v-for="p in properties" :key="p" :value="p">{{ p.toUpperCase() }}</option>
          </select>
        </div>
        <div v-else class="text-xs text-text-secondary text-center py-4">
          No keyframes on selected clip. Add keyframes using inspector diamond buttons first.
        </div>

        <!-- Curve Canvas -->
        <div class="border border-border/80 rounded-lg overflow-hidden bg-[#0F0F17] flex items-center justify-center p-2">
          <canvas ref="canvasRef" width="530" height="180" class="w-full h-auto select-none" />
        </div>

        <!-- Keyframe Easing List -->
        <div v-if="currentKeyframes.length" class="space-y-2 max-h-48 overflow-y-auto pr-1">
          <div v-for="(kf, idx) in currentKeyframes" :key="kf.id" class="flex items-center justify-between bg-bg/60 border border-border/60 rounded px-3 py-2 text-xs">
            <div class="flex items-center gap-2">
              <span class="font-mono text-accent">#{{ idx + 1 }}</span>
              <span class="text-text-secondary font-mono">t={{ kf.time.toFixed(2) }}s</span>
              <span class="text-text-primary font-mono font-semibold">val={{ kf.value }}</span>
            </div>
            <div class="flex items-center gap-1">
              <span class="text-[10px] text-text-secondary uppercase mr-1">Easing:</span>
              <select
                :value="kf.easing || 'linear'"
                @change="setEasing(kf.id, $event.target.value)"
                class="bg-bg border border-border rounded px-2 py-1 text-xs text-text-primary outline-none focus:border-accent"
              >
                <option v-for="opt in easingOptions" :key="opt.id" :value="opt.id">{{ opt.label }}</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end px-4 py-3 border-t border-border bg-bg/30">
        <button class="px-4 py-1.5 rounded bg-accent text-bg text-xs font-semibold hover:bg-accent-hover transition-colors" @click="emit('close')">
          Done
        </button>
      </div>
    </div>
  </div>
</template>
