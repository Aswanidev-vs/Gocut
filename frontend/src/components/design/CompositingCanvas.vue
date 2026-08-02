<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { useCompositingStore } from '../../stores/compositingStore'
import { evaluateGraph, findOutputNode } from '../../engine/graphEvaluator'
import { nodeEvaluators } from '../../engine/nodeEvaluators'

const props = defineProps({
  playheadTime: { type: Number, default: 0 },
  isPlaying: { type: Boolean, default: false },
})

const designStore = useDesignStore()
const compositingStore = useCompositingStore()
const canvasRef = ref(null)
const statusText = ref('')

let animFrame = null

// Render the graph to canvas
async function renderFrame() {
  const canvas = canvasRef.value
  if (!canvas) return

  const ctx2d = canvas.getContext('2d')
  if (!ctx2d) return

  const start = performance.now()
  compositingStore.setRendering(true)

  try {
    // Clear canvas
    ctx2d.clearRect(0, 0, canvas.width, canvas.height)

    // Fill with composition background
    ctx2d.fillStyle = designStore.composition.background || '#000000'
    ctx2d.fillRect(0, 0, canvas.width, canvas.height)

    // If no nodes, show placeholder
    if (designStore.nodes.length === 0) {
      statusText.value = 'No nodes — add sources to see output'
      compositingStore.setRendering(false)
      return
    }

    // Evaluate the graph
    const resolution = {
      width: designStore.composition.width,
      height: designStore.composition.height,
    }

    const outputs = await evaluateGraph(
      designStore.nodes,
      designStore.connections,
      props.playheadTime,
      resolution,
      nodeEvaluators
    )

    // Find output node and render its result
    const outputNode = findOutputNode(designStore.nodes)
    if (outputNode && outputs.has(outputNode.id)) {
      const output = outputs.get(outputNode.id)
      renderOutputToCanvas(ctx2d, output, canvas.width, canvas.height, resolution)
      statusText.value = ''
    } else {
      // Render all source nodes as a preview
      renderSourcesPreview(ctx2d, outputs, canvas.width, canvas.height, resolution)
      statusText.value = 'Connect to an Output node to finalize'
    }

    compositingStore.setNodeOutputs(outputs)
    compositingStore.setRenderTime(performance.now() - start)
  } catch (e) {
    console.error('Compositing render error:', e)
    compositingStore.setError(e.message)
    statusText.value = `Error: ${e.message}`
  } finally {
    compositingStore.setRendering(false)
  }
}

function renderOutputToCanvas(ctx2d, output, canvasW, canvasH, resolution) {
  if (!output || !output.input) return

  const input = output.input
  renderNodeOutput(ctx2d, input, canvasW, canvasH, resolution)
}

function renderSourcesPreview(ctx2d, outputs, canvasW, canvasH, resolution) {
  // Render the last evaluated source/effect node as a preview
  for (const [nodeId, output] of outputs) {
    if (output && output.type !== 'output') {
      renderNodeOutput(ctx2d, output, canvasW, canvasH, resolution)
      break
    }
  }
}

function renderNodeOutput(ctx2d, output, canvasW, canvasH, resolution) {
  if (!output) return

  const scaleX = canvasW / resolution.width
  const scaleY = canvasH / resolution.height

  switch (output.type) {
    case 'solid':
    case 'shape': {
      ctx2d.save()
      // Apply transform if present
      if (output.transform) {
        ctx2d.translate(canvasW / 2, canvasH / 2)
        ctx2d.rotate((output.transform.rotation || 0) * Math.PI / 180)
        ctx2d.scale(output.transform.scaleX || 1, output.transform.scaleY || 1)
        ctx2d.translate(-canvasW / 2 + (output.transform.x || 0) * scaleX, -canvasH / 2 + (output.transform.y || 0) * scaleY)
      }
      ctx2d.fillStyle = output.fill || '#FFFFFF'
      if (output.shape === 'ellipse') {
        ctx2d.beginPath()
        ctx2d.ellipse(
          canvasW / 2, canvasH / 2,
          (output.width || 200) * scaleX / 2,
          (output.height || 200) * scaleY / 2,
          0, 0, Math.PI * 2
        )
        ctx2d.fill()
      } else {
        const w = (output.width || 200) * scaleX
        const h = (output.height || 200) * scaleY
        const x = (canvasW - w) / 2
        const y = (canvasH - h) / 2
        if (output.cornerRadius) {
          ctx2d.beginPath()
          ctx2d.roundRect(x, y, w, h, output.cornerRadius * scaleX)
          ctx2d.fill()
        } else {
          ctx2d.fillRect(x, y, w, h)
        }
      }
      ctx2d.restore()
      break
    }

    case 'gradient': {
      const angle = (output.angle || 90) * Math.PI / 180
      const cx = canvasW / 2
      const cy = canvasH / 2
      const len = Math.max(canvasW, canvasH) / 2
      const x1 = cx - Math.cos(angle) * len
      const y1 = cy - Math.sin(angle) * len
      const x2 = cx + Math.cos(angle) * len
      const y2 = cy + Math.sin(angle) * len
      const grad = ctx2d.createLinearGradient(x1, y1, x2, y2)
      grad.addColorStop(0, output.color1 || '#00D4FF')
      grad.addColorStop(1, output.color2 || '#EC4899')
      ctx2d.fillStyle = grad
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      break
    }

    case 'text': {
      ctx2d.save()
      const fontSize = (output.fontSize || 48) * Math.min(scaleX, scaleY)
      const fontStyle = `${output.italic ? 'italic ' : ''}${output.bold ? 'bold ' : ''}${fontSize}px ${output.fontFamily || 'sans-serif'}`
      ctx2d.font = fontStyle
      ctx2d.fillStyle = output.color || '#FFFFFF'
      ctx2d.textAlign = 'center'
      ctx2d.textBaseline = 'middle'
      ctx2d.fillText(output.text || 'Text', canvasW / 2, canvasH / 2)
      ctx2d.restore()
      break
    }

    case 'transform': {
      if (output.input) {
        ctx2d.save()
        ctx2d.globalAlpha = output.opacity ?? 1
        ctx2d.translate(canvasW / 2 + (output.x || 0) * scaleX, canvasH / 2 + (output.y || 0) * scaleY)
        ctx2d.rotate((output.rotation || 0) * Math.PI / 180)
        ctx2d.scale(output.scaleX || 1, output.scaleY || 1)
        ctx2d.translate(-canvasW / 2, -canvasH / 2)
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.restore()
      }
      break
    }

    case 'blur': {
      if (output.input) {
        ctx2d.save()
        ctx2d.filter = `blur(${output.radius || 5}px)`
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.filter = 'none'
        ctx2d.restore()
      }
      break
    }

    case 'colorCorrect': {
      if (output.input) {
        ctx2d.save()
        const filters = []
        if (output.brightness) filters.push(`brightness(${1 + output.brightness})`)
        if (output.contrast) filters.push(`contrast(${1 + output.contrast})`)
        if (output.saturation) filters.push(`saturate(${1 + output.saturation})`)
        if (output.hue) filters.push(`hue-rotate(${output.hue}deg)`)
        if (filters.length) ctx2d.filter = filters.join(' ')
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.filter = 'none'
        ctx2d.restore()
      }
      break
    }

    case 'merge': {
      if (output.bg) renderNodeOutput(ctx2d, output.bg, canvasW, canvasH, resolution)
      if (output.fg) {
        ctx2d.save()
        // Apply blend mode via globalCompositeOperation
        const blendMap = {
          normal: 'source-over',
          multiply: 'multiply',
          screen: 'screen',
          overlay: 'overlay',
          lighten: 'lighten',
          darken: 'darken',
          'colordodge': 'color-dodge',
          'colorburn': 'color-burn',
          difference: 'difference',
          add: 'lighter',
          subtract: 'source-over', // CSS doesn't have subtract, approximate
        }
        ctx2d.globalCompositeOperation = blendMap[output.mode] || 'source-over'
        renderNodeOutput(ctx2d, output.fg, canvasW, canvasH, resolution)
        ctx2d.restore()
      }
      break
    }

    case 'output': {
      if (output.input) {
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
      }
      break
    }

    case 'noise': {
      // Simple noise visualization using canvas patterns
      const size = Math.max(1, Math.floor(output.scale || 50))
      const noiseCanvas = document.createElement('canvas')
      noiseCanvas.width = size
      noiseCanvas.height = size
      const nCtx = noiseCanvas.getContext('2d')
      const imgData = nCtx.createImageData(size, size)
      const seed = output.seed || 0
      for (let i = 0; i < imgData.data.length; i += 4) {
        const v = Math.floor(((Math.sin(seed + i * 0.01) * 43758.5453) % 1 + 1) % 1 * 255)
        imgData.data[i] = v
        imgData.data[i + 1] = v
        imgData.data[i + 2] = v
        imgData.data[i + 3] = 255
      }
      nCtx.putImageData(imgData, 0, 0)
      const pattern = ctx2d.createPattern(noiseCanvas, 'repeat')
      ctx2d.fillStyle = pattern || '#808080'
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      break
    }

    case 'mask': {
      if (output.input) {
        ctx2d.save()
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        // Apply mask using composite operations
        ctx2d.globalCompositeOperation = 'destination-in'
        ctx2d.fillStyle = 'white'
        const mx = (output.x || 0) * scaleX
        const my = (output.y || 0) * scaleY
        const mw = (output.width || 1920) * scaleX
        const mh = (output.height || 1080) * scaleY
        if (output.maskType === 'ellipse') {
          ctx2d.beginPath()
          ctx2d.ellipse(mx + mw / 2, my + mh / 2, mw / 2, mh / 2, 0, 0, Math.PI * 2)
          ctx2d.fill()
        } else {
          if (output.feather) {
            ctx2d.filter = `blur(${output.feather * scaleX}px)`
          }
          ctx2d.fillRect(mx, my, mw, mh)
          ctx2d.filter = 'none'
        }
        ctx2d.restore()
      }
      break
    }

    case 'crop': {
      if (output.input) {
        ctx2d.save()
        const cx = (output.x || 0) * scaleX
        const cy = (output.y || 0) * scaleY
        const cw = (output.width || 1920) * scaleX
        const ch = (output.height || 1080) * scaleY
        ctx2d.beginPath()
        ctx2d.rect(cx, cy, cw, ch)
        ctx2d.clip()
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.restore()
      }
      break
    }

    case 'invert': {
      if (output.input) {
        ctx2d.save()
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        // Invert using composite operation
        ctx2d.globalCompositeOperation = 'difference'
        ctx2d.fillStyle = 'white'
        ctx2d.fillRect(0, 0, canvasW, canvasH)
        ctx2d.restore()
      }
      break
    }

    case 'levels': {
      if (output.input) {
        ctx2d.save()
        // Approximate levels with brightness/contrast
        const inRange = (output.inWhite || 255) - (output.inBlack || 0)
        const outRange = (output.outWhite || 255) - (output.outBlack || 0)
        const gamma = output.gamma || 1
        const brightness = (output.outBlack || 0) / 255
        const contrast = outRange / inRange - 1
        const filters = []
        if (brightness !== 0) filters.push(`brightness(${1 + brightness})`)
        if (contrast !== 0) filters.push(`contrast(${1 + contrast})`)
        if (filters.length) ctx2d.filter = filters.join(' ')
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.filter = 'none'
        ctx2d.restore()
      }
      break
    }

    case 'temperature': {
      if (output.input) {
        ctx2d.save()
        const temp = output.temperature || 0
        const tint = output.tint || 0
        const filters = []
        if (temp !== 0) filters.push(`sepia(${Math.abs(temp) * 0.5})`)
        if (tint !== 0) filters.push(`hue-rotate(${tint * 30}deg)`)
        if (filters.length) ctx2d.filter = filters.join(' ')
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.filter = 'none'
        ctx2d.restore()
      }
      break
    }

    case 'directionalBlur': {
      if (output.input) {
        ctx2d.save()
        // Approximate directional blur with motion blur CSS
        const angle = output.angle || 0
        const dist = output.distance || 10
        ctx2d.filter = `blur(${dist * 0.5}px)`
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.filter = 'none'
        ctx2d.restore()
      }
      break
    }

    case 'solid': {
      ctx2d.fillStyle = output.fill || '#00D4FF'
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      break
    }

    default: {
      // Unknown output type — render as colored rectangle
      ctx2d.fillStyle = '#333'
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      ctx2d.fillStyle = '#666'
      ctx2d.font = '14px sans-serif'
      ctx2d.textAlign = 'center'
      ctx2d.fillText(`Unknown: ${output.type}`, canvasW / 2, canvasH / 2)
    }
  }
}

// Watch for changes and re-render
watch(
  () => [designStore.nodes, designStore.connections, props.playheadTime, designStore.composition],
  () => { requestAnimationFrame(renderFrame) },
  { deep: true }
)

// Handle playhead changes during playback
watch(() => props.playheadTime, () => {
  if (props.isPlaying) {
    requestAnimationFrame(renderFrame)
  }
})

onMounted(() => {
  nextTick(renderFrame)
})

onUnmounted(() => {
  if (animFrame) cancelAnimationFrame(animFrame)
})

defineExpose({ renderFrame })
</script>

<template>
  <div class="relative w-full h-full bg-black flex flex-col">
    <canvas
      ref="canvasRef"
      :width="designStore.composition.width"
      :height="designStore.composition.height"
      class="w-full h-full object-contain"
    />
    <!-- Status overlay -->
    <div v-if="statusText" class="absolute bottom-2 left-2 right-2 flex items-center justify-center">
      <div class="px-2 py-1 rounded bg-black/60 text-[10px] text-text-secondary font-mono">
        {{ statusText }}
      </div>
    </div>
    <!-- Render info -->
    <div class="absolute top-2 right-2 text-[9px] text-text-secondary/50 font-mono">
      {{ designStore.composition.width }}×{{ designStore.composition.height }}
      <span v-if="compositingStore.renderTime > 0"> · {{ compositingStore.renderTime.toFixed(1) }}ms</span>
    </div>
  </div>
</template>
