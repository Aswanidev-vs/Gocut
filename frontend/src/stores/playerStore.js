import { defineStore, getActivePinia } from 'pinia'
import { ref, computed, watch } from 'vue'
import { GetPreviewFrame, PreloadFrames, GetMediaServerPort } from '../lib/wails'

// playerStore touches both projectStore and timelineStore. To avoid the
// "Cannot access '<store>' before initialization" crash when the first
// component to be mounted happens to be one of these, we resolve the
// other stores lazily through the active pinia.

function getProjectStore() {
  try {
    const pinia = getActivePinia()
    if (!pinia) return null
    return pinia._s && pinia._s.get('project') ? pinia._s.get('project') : null
  } catch (_) { return null }
}
function getTimelineStore() {
  try {
    const pinia = getActivePinia()
    if (!pinia) return null
    return pinia._s && pinia._s.get('timeline') ? pinia._s.get('timeline') : null
  } catch (_) { return null }
}

export const usePlayerStore = defineStore('player', () => {
  const isPlaying = ref(false)
  const currentTime = ref(0)
  const volume = ref(1.0)
  const loop = ref(false)
  const previewQuality = ref('full')
  const isScrubbing = ref(false)
  const previewImage = ref(null)
  const isPreviewLoading = ref(false)
  const previewError = ref(null)
  let previewRequestId = 0
  let lastCompletedPreviewTime = -1
  let queuedPreview = false
  let queuedPreviewTime = 0
  const mediaServerPort = ref(0)

  const formattedTime = computed(() => formatTimecode(currentTime.value))

  // Mirror timeline store's currentTime so the player + preview share state.
  watch(() => {
    const ts = getTimelineStore()
    return ts ? ts.currentTime : 0
  }, (t) => {
    if (Math.abs(t - currentTime.value) > 0.01) {
      currentTime.value = t
    }
  })

  function play() {
    const ts = getTimelineStore()
    if (ts && ts.duration > 0) isPlaying.value = true
  }

  function pause() {
    isPlaying.value = false
  }

  function togglePlay() {
    if (isPlaying.value) pause()
    else play()
  }

  async function seek(time) {
    const ts = getTimelineStore()
    if (ts) ts.setCurrentTime(time)
    await refreshPreview(true)
  }

  async function refreshPreview(force = false, overrideTime = null) {
    const ps = getProjectStore()
    if (!ps || !ps.project) {
      previewImage.value = null
      return
    }

    const targetTime = overrideTime ?? currentTime.value

    if (!force && isPreviewLoading.value) {
      queuedPreview = true
      queuedPreviewTime = targetTime
      return
    }

    if (!force && Math.abs(targetTime - lastCompletedPreviewTime) < 0.03) {
      return
    }

    const requestId = ++previewRequestId
    isPreviewLoading.value = true
    previewError.value = null
    try {
      const w = ps.project.resolution?.width || 1920
      const h = ps.project.resolution?.height || 1080
      const data = await GetPreviewFrame(JSON.parse(JSON.stringify(ps.project)), targetTime, w, h)
      if (requestId !== previewRequestId) return
      previewImage.value = data
      lastCompletedPreviewTime = targetTime
    } catch (e) {
      if (requestId !== previewRequestId) return
      previewError.value = e?.message || String(e)
      previewImage.value = null
    } finally {
      if (requestId === previewRequestId) {
        isPreviewLoading.value = false
      }
      if (queuedPreview) {
        const nextTime = queuedPreviewTime
        queuedPreview = false
        queuedPreviewTime = 0
        Promise.resolve().then(() => refreshPreview(true, nextTime))
      }
    }
  }

  function invalidatePreview() {
    previewRequestId++
    queuedPreview = false
    queuedPreviewTime = 0
    lastCompletedPreviewTime = -1
    previewImage.value = null
    previewError.value = null
  }

  // Drops the "already rendered this timestamp" memo without clearing the
  // visible frame, so a non-force refresh still goes through when clip
  // content changed but the playhead did not.
  function markPreviewStale() {
    lastCompletedPreviewTime = -1
  }

  async function preload() {
    const ps = getProjectStore()
    if (!ps || !ps.project) return
    try {
      await PreloadFrames(JSON.parse(JSON.stringify(ps.project)), currentTime.value, 5)
    } catch (_) { /* ignore */ }
  }

  function stepForward() {
    const ps = getProjectStore()
    const fps = ps?.project?.fps || 30
    const ts = getTimelineStore()
    if (ts) ts.setCurrentTime(currentTime.value + 1 / fps)
  }

  function stepBackward() {
    const ps = getProjectStore()
    const fps = ps?.project?.fps || 30
    const ts = getTimelineStore()
    if (ts) ts.setCurrentTime(Math.max(0, currentTime.value - 1 / fps))
  }

  function setVolume(value) {
    volume.value = Math.max(0, Math.min(2.0, value))
  }

  function toggleLoop() {
    loop.value = !loop.value
  }

  function setPreviewQuality(quality) {
    previewQuality.value = quality
  }

  function formatTimecode(seconds) {
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = Math.floor(seconds % 60)
    const ms = Math.floor((seconds % 1) * 1000)
    const pad = (n, len = 2) => String(n).padStart(len, '0')
    return pad(h) + ':' + pad(m) + ':' + pad(s) + '.' + pad(ms, 3)
  }

  async function fetchMediaServerPort() {
    try {
      const port = await GetMediaServerPort()
      mediaServerPort.value = port || 0
    } catch (_) { /* ignore */ }
  }

  function getMediaUrl(path, asAudioProxy = false) {
    if (!mediaServerPort.value || !path) return ''
    const endpoint = asAudioProxy ? 'audio' : 'media'
    return `http://127.0.0.1:${mediaServerPort.value}/${endpoint}?path=${encodeURIComponent(path)}`
  }

  return {
    isPlaying,
    currentTime,
    volume,
    loop,
    previewQuality,
    isScrubbing,
    previewImage,
    isPreviewLoading,
    previewError,
    formattedTime,
    play,
    pause,
    togglePlay,
    seek,
    stepForward,
    stepBackward,
    setVolume,
    toggleLoop,
    setPreviewQuality,
    refreshPreview,
    invalidatePreview,
    markPreviewStale,
    preload,
    formatTimecode,
    mediaServerPort,
    fetchMediaServerPort,
    getMediaUrl,
  }
})
