import { defineStore } from 'pinia'
import { ref, watch } from 'vue'
import { useDesignStore } from './designStore'

export const useCompositingStore = defineStore('compositing', () => {
  const isRendering = ref(false)
  const lastError = ref(null)
  const renderTime = ref(0) // ms per frame
  const nodeOutputs = ref(new Map())

  let renderRequested = false
  let debounceTimer = null

  function requestRender() {
    if (renderRequested) return
    renderRequested = true

    clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      renderRequested = false
      // The CompositingCanvas component watches this and triggers re-render
    }, 16) // ~60fps debounce
  }

  function setRendering(val) {
    isRendering.value = val
  }

  function setError(err) {
    lastError.value = err
  }

  function setRenderTime(ms) {
    renderTime.value = ms
  }

  function setNodeOutputs(outputs) {
    nodeOutputs.value = outputs
  }

  return {
    isRendering,
    lastError,
    renderTime,
    nodeOutputs,
    requestRender,
    setRendering,
    setError,
    setRenderTime,
    setNodeOutputs,
  }
})
