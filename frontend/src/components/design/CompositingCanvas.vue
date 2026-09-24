<script setup>
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useDesignStore } from '../../stores/designStore'
import { useCompositingStore } from '../../stores/compositingStore'
import { evaluateGraph, findOutputNode } from '../../engine/graphEvaluator'
import { nodeEvaluators } from '../../engine/nodeEvaluators'

const props = defineProps({
  playheadTime: { type: Number, default: 0 },
  isPlaying: { type: Boolean, default: false },
  targetNodeId: { type: String, default: null },
  viewerLabel: { type: String, default: '' },
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

    // Fusion Viewer Target Routing
    if (props.targetNodeId && outputs.has(props.targetNodeId)) {
      const output = outputs.get(props.targetNodeId)
      renderNodeOutput(ctx2d, output, canvas.width, canvas.height, resolution)
      const targetNode = designStore.nodes.find(n => n.id === props.targetNodeId)
      statusText.value = targetNode ? `${targetNode.label || targetNode.type}` : ''
    } else {
      // Find output / mediaOut node and render its result
      const outputNode = designStore.nodes.find(n => n.type === 'output' || n.type === 'mediaOut')
      if (outputNode && outputs.has(outputNode.id)) {
        const output = outputs.get(outputNode.id)
        renderOutputToCanvas(ctx2d, output, canvas.width, canvas.height, resolution)
        statusText.value = ''
      } else {
        // Render all source nodes as a preview
        renderSourcesPreview(ctx2d, outputs, canvas.width, canvas.height, resolution)
        statusText.value = 'Connect to an Output/MediaOut node to finalize'
      }
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

    // ===== Fusion toolset previews =====
    // These node types were evaluated but never drawn, so the viewer showed a
    // grey "Unknown" box for the tools Fusion is built around.

    case 'background': {
      const bgType = output.bgType || 'solid'
      ctx2d.save()
      if (bgType === 'horizontal' || bgType === 'vertical') {
        const grad = bgType === 'horizontal'
          ? ctx2d.createLinearGradient(0, 0, canvasW, 0)
          : ctx2d.createLinearGradient(0, 0, 0, canvasH)
        grad.addColorStop(0, output.color || '#000000')
        grad.addColorStop(1, output.color2 || '#1e293b')
        ctx2d.fillStyle = grad
      } else if (bgType === 'radial') {
        const grad = ctx2d.createRadialGradient(
          canvasW / 2, canvasH / 2, 0,
          canvasW / 2, canvasH / 2, Math.max(canvasW, canvasH) / 2
        )
        grad.addColorStop(0, output.color || '#000000')
        grad.addColorStop(1, output.color2 || '#1e293b')
        ctx2d.fillStyle = grad
      } else {
        ctx2d.fillStyle = output.color || '#000000'
      }
      ctx2d.globalAlpha = output.alpha ?? 1
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      ctx2d.restore()
      break
    }

    case 'fastNoise': {
      const tile = Math.max(4, Math.round(output.scale || 15))
      const noiseCanvas = document.createElement('canvas')
      noiseCanvas.width = tile
      noiseCanvas.height = tile
      const nCtx = noiseCanvas.getContext('2d')
      const imgData = nCtx.createImageData(tile, tile)
      const detail = Math.max(1, Math.round(output.detail || 4))
      const contrast = output.contrast ?? 1.5
      const brightness = output.brightness || 0
      const seed = output.seethe || 0.5
      const c1 = hexToRgb(output.color1 || '#000000')
      const c2 = hexToRgb(output.color2 || '#FFFFFF')
      for (let y = 0; y < tile; y++) {
        for (let x = 0; x < tile; x++) {
          // Octave stack of value hashes: a cheap fractal field that still
          // reads as noise when the tile is repeated across the frame.
          let v = 0, amp = 1, norm = 0, step = 1
          for (let o = 0; o < detail; o++) {
            v += amp * hash2D(Math.floor(x / step), Math.floor(y / step), seed + o)
            norm += amp
            amp *= 0.5
            step *= 2
          }
          v = norm > 0 ? v / norm : 0
          v = Math.min(1, Math.max(0, (v - 0.5) * contrast + 0.5 + brightness))
          const i = (y * tile + x) * 4
          imgData.data[i] = c1.r + (c2.r - c1.r) * v
          imgData.data[i + 1] = c1.g + (c2.g - c1.g) * v
          imgData.data[i + 2] = c1.b + (c2.b - c1.b) * v
          imgData.data[i + 3] = 255
        }
      }
      nCtx.putImageData(imgData, 0, 0)
      const pattern = ctx2d.createPattern(noiseCanvas, 'repeat')
      ctx2d.fillStyle = pattern || '#000000'
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      break
    }

    case 'colorCorrector': {
      if (output.input) {
        ctx2d.save()
        const filters = []
        const lift = output.lift || 0
        const gain = output.gain ?? 1
        if (lift || gain !== 1) filters.push(`brightness(${Math.max(0, 1 + lift) * gain})`)
        if (output.gamma && output.gamma !== 1) filters.push(`contrast(${1 / output.gamma})`)
        if (output.brightness) filters.push(`brightness(${Math.max(0, 1 + output.brightness)})`)
        if (output.saturation && output.saturation !== 1) filters.push(`saturate(${output.saturation})`)
        if (output.tint) filters.push(`hue-rotate(${output.tint * 30}deg)`)
        if (filters.length) ctx2d.filter = filters.join(' ')
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
        ctx2d.filter = 'none'
        ctx2d.restore()
      }
      break
    }

    case 'glow': {
      if (output.input) {
        const radius = Math.max(1, (output.radius || 10) * Math.min(scaleX, scaleY))
        const passes = Math.max(1, Math.round(output.intensity || 1))
        ctx2d.save()
        ctx2d.globalCompositeOperation = 'lighter'
        for (let pass = 0; pass < passes; pass++) {
          ctx2d.save()
          ctx2d.filter = `blur(${radius}px) drop-shadow(0 0 ${radius}px ${output.color || '#00D4FF'})`
          renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
          ctx2d.restore()
        }
        ctx2d.restore()
        // Crisp original drawn over its own halo.
        renderNodeOutput(ctx2d, output.input, canvasW, canvasH, resolution)
      }
      break
    }

    case 'shadow': {
      if (output.input) {
        const layer = renderToOffscreen(output.input, canvasW, canvasH, resolution)
        const blur = Math.max(0, (output.blur || 8) * Math.min(scaleX, scaleY))
        ctx2d.save()
        ctx2d.globalAlpha = 0.55
        if (blur) ctx2d.filter = `blur(${blur}px)`
        ctx2d.drawImage(layer, (output.offsetX || 0) * scaleX, (output.offsetY || 4) * scaleY)
        ctx2d.restore()
        ctx2d.drawImage(layer, 0, 0)
      }
      break
    }

    case 'chromaKey': {
      if (output.input) {
        const layer = renderToOffscreen(output.input, canvasW, canvasH, resolution)
        const lCtx = layer.getContext('2d')
        const imgData = lCtx.getImageData(0, 0, layer.width, layer.height)
        const key = hexToRgb(output.keyColor || '#00FF00')
        // Normalised RGB distance, so similarity/smoothness feel the same as
        // they do in Fusion's Chroma Keyer.
        const maxDist = Math.sqrt(3) * 255
        const similarity = Math.max(0, output.similarity ?? 0.4) * maxDist
        const smoothness = Math.max(1, (output.smoothness ?? 0.1) * maxDist)
        const data = imgData.data
        for (let i = 0; i < data.length; i += 4) {
          if (data[i + 3] === 0) continue
          const dist = Math.sqrt(
            (data[i] - key.r) ** 2 + (data[i + 1] - key.g) ** 2 + (data[i + 2] - key.b) ** 2
          )
          if (dist < similarity) data[i + 3] = 0
          else if (dist < similarity + smoothness) {
            data[i + 3] = Math.round(255 * (dist - similarity) / smoothness)
          }
        }
        lCtx.putImageData(imgData, 0, 0)
        ctx2d.drawImage(layer, 0, 0)
      }
      break
    }

    case 'maskPolygon': {
      // The evaluator carries no polygon points yet, so the preview shows the
      // mask plate it produces (feather, border, invert) rather than geometry.
      ctx2d.save()
      ctx2d.fillStyle = '#FFFFFF'
      if (output.softEdge) ctx2d.filter = `blur(${output.softEdge}px)`
      const inset = Math.min(canvasW, canvasH) * 0.08
      ctx2d.beginPath()
      if (output.invert || output.paintMode === 'subtract') {
        ctx2d.rect(0, 0, canvasW, canvasH)
        ctx2d.roundRect(inset, inset, canvasW - inset * 2, canvasH - inset * 2, 8)
        ctx2d.fill('evenodd')
      } else {
        ctx2d.roundRect(inset, inset, canvasW - inset * 2, canvasH - inset * 2, 8)
        ctx2d.fill()
      }
      if (output.borderWidth) {
        ctx2d.filter = 'none'
        ctx2d.strokeStyle = '#FFFFFF'
        ctx2d.lineWidth = output.borderWidth
        ctx2d.stroke()
      }
      ctx2d.restore()
      break
    }

    case 'media':
      // Media frames live in the ffmpeg frame cache, which the design viewer
      // cannot reach, so show an honest slate instead of a grey error box.
      ctx2d.fillStyle = '#0E0E11'
      ctx2d.fillRect(0, 0, canvasW, canvasH)
      ctx2d.strokeStyle = '#232329'
      ctx2d.lineWidth = 1
      ctx2d.strokeRect(0.5, 0.5, canvasW - 1, canvasH - 1)
      ctx2d.fillStyle = '#8A8A94'
      ctx2d.font = `${Math.max(10, Math.round(canvasH * 0.028))}px "JetBrains Mono", monospace`
      ctx2d.textAlign = 'center'
      ctx2d.textBaseline = 'middle'
      ctx2d.fillText(
        output.assetId ? `MediaIn · ${String(output.assetId).slice(0, 8)}` : 'MediaIn · pick a source',
        canvasW / 2, canvasH / 2
      )
      ctx2d.textAlign = 'left'
      ctx2d.textBaseline = 'alphabetic'
      break

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

// Effects that need to read their input's pixels cannot do it through the
// recursive draw-into-one-context design, so they render the input to an
// offscreen buffer first.
function renderToOffscreen(output, canvasW, canvasH, resolution) {
  const off = document.createElement('canvas')
  off.width = canvasW
  off.height = canvasH
  renderNodeOutput(off.getContext('2d'), output, canvasW, canvasH, resolution)
  return off
}

function hash2D(x, y, seed) {
  const n = Math.sin(x * 127.1 + y * 311.7 + seed * 74.7) * 43758.5453
  return n - Math.floor(n)
}

function hexToRgb(hex) {
  let h = String(hex || '').trim().replace(/^#/, '')
  if (h.length === 3) h = h.split('').map(c => c + c).join('')
  if (!/^[0-9a-f]{6}$/i.test(h)) return { r: 255, g: 255, b: 255 }
  return {
    r: parseInt(h.slice(0, 2), 16),
    g: parseInt(h.slice(2, 4), 16),
    b: parseInt(h.slice(4, 6), 16),
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
  <div class="relative w-full h-full bg-cr-void flex flex-col select-none overflow-hidden">
    <!-- Viewer header: hairline rule, mono numerals, no chrome above the image -->
    <div class="h-6 bg-cr-panel border-b border-cr-line flex items-center justify-between px-2 text-[9px] z-10">
      <div class="flex items-center gap-2 min-w-0">
        <span class="px-1.5 py-px font-jetbrains-mono uppercase tracking-[0.14em] text-[9px] text-ink-dim border border-cr-line">
          {{ viewerLabel || 'Viewer' }}
        </span>
        <span v-if="statusText" class="text-ink truncate">{{ statusText }}</span>
      </div>
      <div class="flex items-center gap-2 font-jetbrains-mono tabular-nums text-ink-faint">
        <span>{{ designStore.composition.width }}×{{ designStore.composition.height }}</span>
        <span v-if="compositingStore.renderTime > 0" class="text-ink-dim">· {{ compositingStore.renderTime.toFixed(1) }}ms</span>
      </div>
    </div>

    <!-- Canvas area -->
    <div class="flex-1 relative flex items-center justify-center overflow-hidden">
      <canvas
        ref="canvasRef"
        :width="designStore.composition.width"
        :height="designStore.composition.height"
        class="max-w-full max-h-full object-contain border border-cr-line"
      />
    </div>
  </div>
</template>
