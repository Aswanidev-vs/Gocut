import { defineStore } from 'pinia'
import { ref } from 'vue'
import { AnalyzeMotion, ApplyStabilization } from '../lib/wails'

export const useTrackingStore = defineStore('tracking', () => {
  const isAnalyzing = ref(false)
  const progress = ref(0)
  const results = ref(null) // TrackingData
  const error = ref(null)
  const selectedAssetId = ref(null)
  const method = ref('stabilize') // 'stabilize' | 'point'
  const settings = ref({
    regionX: 0,
    regionY: 0,
    regionW: 100,
    regionH: 100,
    shaking: 1, // 0=low, 1=medium, 2=high
    accuracy: 1, // 0=fast, 1=accurate, 2=very accurate
  })

  async function analyze(assetId, startTime, duration) {
    if (!assetId) return
    isAnalyzing.value = true
    progress.value = 0
    error.value = null
    results.value = null
    selectedAssetId.value = assetId

    try {
      const trackingSettings = {
        method: method.value,
        assetId,
        startTime: startTime || 0,
        duration: duration || 5,
        regionX: settings.value.regionX,
        regionY: settings.value.regionY,
        regionW: settings.value.regionW,
        regionH: settings.value.regionH,
        shaking: settings.value.shaking,
        accuracy: settings.value.accuracy,
      }

      results.value = await AnalyzeMotion(trackingSettings)
    } catch (e) {
      error.value = e?.message || String(e)
    } finally {
      isAnalyzing.value = false
    }
  }

  async function applyStabilization(assetId, outputPath) {
    if (!results.value) return
    isAnalyzing.value = true
    progress.value = 0
    error.value = null

    try {
      await ApplyStabilization(assetId, results.value, outputPath)
    } catch (e) {
      error.value = e?.message || String(e)
    } finally {
      isAnalyzing.value = false
    }
  }

  function clear() {
    results.value = null
    error.value = null
    progress.value = 0
    selectedAssetId.value = null
  }

  return {
    isAnalyzing,
    progress,
    results,
    error,
    selectedAssetId,
    method,
    settings,
    analyze,
    applyStabilization,
    clear,
  }
})
