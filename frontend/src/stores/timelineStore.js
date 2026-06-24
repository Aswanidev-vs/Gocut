import { defineStore, getActivePinia } from 'pinia'
import { ref, computed, watch } from 'vue'

// IMPORTANT: We do NOT import useProjectStore at the top of this file.
// Importing it would create a circular dependency: projectStore also
// touches the timeline store on setProject, and the first time
// useTimelineStore() is called from a component the project store has
// not been registered yet, so the cross-store call throws
// "Cannot access 'project' before initialization".
//
// Instead we resolve the project store lazily through the active pinia.

const generateId = () => crypto.randomUUID()

function getProjectStore() {
  try {
    const pinia = getActivePinia()
    if (!pinia) return null
    return pinia._s && pinia._s.get('project') ? pinia._s.get('project') : null
  } catch (_) {
    return null
  }
}

function getHistoryStore() {
  try {
    const pinia = getActivePinia()
    if (!pinia) return null
    return pinia._s && pinia._s.get('history') ? pinia._s.get('history') : null
  } catch (_) {
    return null
  }
}

function safeMarkDirty() {
  try {
    const ps = getProjectStore()
    if (ps && typeof ps.markDirty === 'function') ps.markDirty()
    
    const hs = getHistoryStore()
    if (hs && typeof hs.pushSnapshot === 'function') hs.pushSnapshot()
  } catch (_) { /* ignore */ }
}

export const TRACK_TYPES = {
  // Order here defines the natural z-stacking: video (background)
  // → image (still background / image-only segment) → pip (foreground
  // overlay, drawn on top of the main composition) → text / sticker /
  // fx (decoration on the very top).
  video:   { label: 'V',  name: 'Video',   icon: 'Video',     color: '#00D4FF' },
  audio:   { label: 'A',  name: 'Audio',   icon: 'Music',     color: '#8B5CF6' },
  image:   { label: 'I',  name: 'Image',   icon: 'Image',     color: '#22C55E' },
  pip:     { label: 'P',  name: 'PIP',     icon: 'PictureInPicture', color: '#F472B6' },
  text:    { label: 'T',  name: 'Text',    icon: 'Type',      color: '#F59E0B' },
  sticker: { label: 'S',  name: 'Sticker', icon: 'Smile',     color: '#EC4899' },
  fx:      { label: 'FX', name: 'Effects', icon: 'Sparkles',  color: '#10B981' },
}

const DEFAULT_TRANSFORM = () => ({
  x: 0, y: 0, scaleX: 1, scaleY: 1, rotation: 0,
  flipH: false, flipV: false, cropX: 0, cropY: 0, cropW: 0, cropH: 0,
})

const DEFAULT_COLOR = () => ({
  brightness: 0, contrast: 0, saturation: 0, hue: 0,
  sharpness: 0, vignette: 0, grain: 0, blur: 0,
  tint: 0, temp: 0, highlights: 0, shadows: 0,
  liftR: 0, liftG: 0, liftB: 0,
  gammaR: 0, gammaG: 0, gammaB: 0,
  gainR: 0, gainG: 0, gainB: 0,
  curves: '',
  chromaKeyColor: '',
  chromaKeySimilarity: 0.01,
  chromaKeyBlend: 0.0,
})

export function getInterpolatedProperty(clip, property, time, defaultValue) {
  if (!clip || !clip.keyframes) return defaultValue
  const kfs = clip.keyframes.filter(k => k.property === property).sort((a, b) => a.time - b.time)
  if (kfs.length === 0) return defaultValue
  
  if (time <= kfs[0].time) return parseFloat(kfs[0].value)
  if (time >= kfs[kfs.length - 1].time) return parseFloat(kfs[kfs.length - 1].value)
  
  for (let i = 0; i < kfs.length - 1; i++) {
    const k1 = kfs[i]
    const k2 = kfs[i + 1]
    if (time >= k1.time && time < k2.time) {
      if (k1.time === k2.time) return parseFloat(k2.value)
      return parseFloat(k1.value) + (parseFloat(k2.value) - parseFloat(k1.value)) * (time - k1.time) / (k2.time - k1.time)
    }
  }
  return defaultValue
}

/**
 * The timeline store is the single source of truth for tracks/clips during
 * the current editing session. Persistence happens through projectStore which
 * embeds the timeline inside the project JSON on save.
 */
export const useTimelineStore = defineStore('timeline', () => {
  const tracks = ref([])
  const clips = ref([])
  const selectedClipIds = ref([])
  const clipboard = ref([])

  const zoom = ref(50)        // px per second
  const scrollX = ref(0)
  const snapEnabled = ref(true)
  const currentTime = ref(0)
  const rippleDelete = ref(false)

  const duration = computed(() => {
    let maxEnd = 0
    for (const c of clips.value) {
      const end = c.startTime + c.duration
      if (end > maxEnd) maxEnd = end
    }
    return maxEnd
  })

  const selectedClips = computed(() =>
    clips.value.filter(c => selectedClipIds.value.includes(c.id))
  )

  // ---- track operations ----
  function addTrack(type = 'video', opts = {}) {
    const track = {
      id: generateId(),
      type,
      muted: false,
      locked: false,
      volume: 1.0,
      label: opts.label || '',
    }
    tracks.value.push(track)
    return track
  }

  function removeTrack(trackId) {
    clips.value = clips.value.filter(c => c.trackId !== trackId)
    tracks.value = tracks.value.filter(t => t.id !== trackId)
  }

  function getTrackByType(type) {
    return tracks.value.find(t => t.type === type)
  }

  // ---- clip operations ----
  function addClip({
    assetId = '',
    trackId = null,
    trackType = null,
    startTime = 0,
    duration = 3,
    trimStart = 0,
    trimEnd = 0,
    textProps = null,
    stickerProps = null,
  } = {}) {
    if (!trackId && trackType) {
      let t = getTrackByType(trackType)
      if (!t) t = addTrack(trackType)
      trackId = t.id
    }
    if (!trackId) return null
    const clip = {
      id: generateId(),
      assetId,
      trackId,
      startTime: Math.max(0, startTime),
      duration: Math.max(0.1, duration),
      trimStart: Math.max(0, trimStart),
      trimEnd: Math.max(0, trimEnd),
      speed: 1.0,
      reversed: false,
      volume: 1.0,
      opacity: 1.0,
      normalize: false,
      noiseReduction: false,
      transform: DEFAULT_TRANSFORM(),
      color: DEFAULT_COLOR(),
      keyframes: [],
      transition: null,
      textProps,
      stickerProps,
    }
    clips.value.push(clip)
    safeMarkDirty()
    return clip
  }

  function removeClip(clipId) {
    clips.value = clips.value.filter(c => c.id !== clipId)
    selectedClipIds.value = selectedClipIds.value.filter(id => id !== clipId)
    safeMarkDirty()
  }

  function removeSelected() {
    if (rippleDelete.value) {
      const removed = selectedClips.value
      const removedByTrack = {}
      for (const c of removed) {
        removedByTrack[c.trackId] = removedByTrack[c.trackId] || []
        removedByTrack[c.trackId].push(c)
      }
      for (const [tid, list] of Object.entries(removedByTrack)) {
        const totalRemoved = list.reduce((s, c) => s + c.duration, 0)
        for (const c of clips.value) {
          if (c.trackId !== tid) continue
          if (removed.find(r => r.id === c.id)) continue
          if (c.startTime >= Math.min(...list.map(r => r.startTime))) {
            c.startTime = Math.max(0, c.startTime - totalRemoved)
          }
        }
      }
    }
    clips.value = clips.value.filter(c => !selectedClipIds.value.includes(c.id))
    selectedClipIds.value = []
    safeMarkDirty()
  }

  function updateClip(clipId, updates) {
    const clip = clips.value.find(c => c.id === clipId)
    if (clip) {
      Object.assign(clip, updates)
      safeMarkDirty()
    }
  }

  function updateClipTransform(clipId, transform) {
    const clip = clips.value.find(c => c.id === clipId)
    if (clip) {
      clip.transform = { ...clip.transform, ...transform }
      safeMarkDirty()
    }
  }

  function updateClipColor(clipId, color) {
    const clip = clips.value.find(c => c.id === clipId)
    if (clip) {
      clip.color = { ...clip.color, ...color }
      safeMarkDirty()
    }
  }

  function moveClip(clipId, newStartTime) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip) return
    const t = snapToClips(Math.max(0, newStartTime), clipId)
    clip.startTime = t
    safeMarkDirty()
  }

  function trimClip(clipId, edge, newValue) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip) return
    if (edge === 'left') {
      const v = Math.max(0, Math.min(newValue, clip.startTime + clip.duration - 0.1))
      const delta = v - clip.startTime
      clip.startTime = v
      clip.trimStart = Math.max(0, clip.trimStart + delta)
      clip.duration = Math.max(0.1, clip.duration - delta)
    } else {
      const v = Math.max(clip.startTime + 0.1, newValue)
      clip.duration = v - clip.startTime
    }
    safeMarkDirty()
  }

  function splitClipAt(clipId, time) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip) return
    if (time <= clip.startTime || time >= clip.startTime + clip.duration) return
    const left = { ...clip, duration: time - clip.startTime }
    const right = {
      ...clip,
      id: generateId(),
      startTime: time,
      duration: clip.startTime + clip.duration - time,
      trimStart: clip.trimStart + (time - clip.startTime),
    }
    const idx = clips.value.findIndex(c => c.id === clipId)
    clips.value.splice(idx, 1, left, right)
    safeMarkDirty()
  }

  function addKeyframe(clipId, property, value, time) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip) return
    clip.keyframes = clip.keyframes || []
    clip.keyframes = clip.keyframes.filter(k => !(k.property === property && Math.abs(k.time - time) < 0.001))
    clip.keyframes.push({ id: generateId(), property, value, time, easing: 'linear' })
    safeMarkDirty()
  }

  function removeKeyframeAtTime(clipId, property, time) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip || !clip.keyframes) return
    clip.keyframes = clip.keyframes.filter(k => !(k.property === property && Math.abs(k.time - time) < 0.001))
    safeMarkDirty()
  }

  function removeKeyframe(clipId, keyframeId) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip || !clip.keyframes) return
    clip.keyframes = clip.keyframes.filter(k => k.id !== keyframeId)
    safeMarkDirty()
  }

  function updateKeyframe(clipId, keyframeId, updates) {
    const clip = clips.value.find(c => c.id === clipId)
    if (!clip || !clip.keyframes) return
    const kf = clip.keyframes.find(k => k.id === keyframeId)
    if (kf) {
      Object.assign(kf, updates)
      safeMarkDirty()
    }
  }

  function snapToClips(t, exceptId = null) {
    if (!snapEnabled.value) return t
    const threshold = 8 / Math.max(1, zoom.value)
    let best = t
    let bestDelta = threshold
    const check = (candidate) => {
      if (Math.abs(candidate - t) < bestDelta) {
        bestDelta = Math.abs(candidate - t)
        best = candidate
      }
    }
    if (Math.abs(0 - t) < bestDelta) best = 0
    for (const c of clips.value) {
      if (c.id === exceptId) continue
      check(c.startTime)
      check(c.startTime + c.duration)
    }
    return best
  }

  function selectClip(clipId, additive = false) {
    if (additive) {
      if (selectedClipIds.value.includes(clipId)) {
        selectedClipIds.value = selectedClipIds.value.filter(id => id !== clipId)
      } else {
        selectedClipIds.value.push(clipId)
      }
    } else {
      selectedClipIds.value = [clipId]
    }
  }

  function clearSelection() {
    selectedClipIds.value = []
  }

  function setZoom(value) {
    zoom.value = Math.max(5, Math.min(400, value))
  }

  function zoomBy(factor) {
    setZoom(zoom.value * factor)
  }

  function setCurrentTime(time) {
    if (typeof time !== 'number' || isNaN(time)) return
    currentTime.value = Math.max(0, time)
  }

  function loadFromProject(p) {
    tracks.value = (p?.timeline?.tracks || []).map(t => ({ ...t }))
    clips.value = (p?.timeline?.tracks || []).flatMap(t =>
      (t.clips || []).map(c => ({ ...c }))
    )
    selectedClipIds.value = []
    currentTime.value = 0
    scrollX.value = 0
  }



  // ---- history ----
  function createSnapshot() {
    return {
      tracks: JSON.parse(JSON.stringify(tracks.value)),
      clips: JSON.parse(JSON.stringify(clips.value))
    }
  }

  function restoreSnapshot(snapshot) {
    if (!snapshot) return
    tracks.value = JSON.parse(JSON.stringify(snapshot.tracks || []))
    clips.value = JSON.parse(JSON.stringify(snapshot.clips || []))
    
    // Clear selection if selected clips no longer exist
    const clipIds = new Set(clips.value.map(c => c.id))
    selectedClipIds.value = selectedClipIds.value.filter(id => clipIds.has(id))
  }

  // ---- clipboard ----
  function copySelected() {
    if (selectedClips.value.length === 0) return
    clipboard.value = JSON.parse(JSON.stringify(selectedClips.value))
  }

  function cutSelected() {
    if (selectedClips.value.length === 0) return
    copySelected()
    removeSelected()
  }

  function pasteSelected() {
    if (clipboard.value.length === 0) return
    
    // Find earliest start time in clipboard to maintain relative positioning
    const minTime = Math.min(...clipboard.value.map(c => c.startTime))
    const offset = currentTime.value - minTime

    const newClips = clipboard.value.map(c => {
      const newClip = {
        ...c,
        id: generateId(),
        startTime: Math.max(0, c.startTime + offset)
      }
      return newClip
    })

    clips.value.push(...newClips)
    selectedClipIds.value = newClips.map(c => c.id)
    safeMarkDirty()
  }

  function clearAll() {
    tracks.value = []
    clips.value = []
    selectedClipIds.value = []
    currentTime.value = 0
    zoom.value = 50
    scrollX.value = 0
  }

  watch([tracks, clips], () => {
    const ps = getProjectStore()
    if (!ps || !ps.project) return
    ps.project.timeline = ps.project.timeline || { tracks: [], duration: 0 }
    ps.project.timeline.tracks = tracks.value.map(t => ({
      ...t,
      clips: clips.value
        .filter(c => c.trackId === t.id)
        .map(c => ({ ...c })),
    }))
    ps.project.timeline.duration = duration.value
    if (typeof ps.markDirty === 'function') ps.markDirty()
  }, { deep: true })

  return {
    tracks,
    clips,
    selectedClipIds,
    zoom,
    scrollX,
    snapEnabled,
    rippleDelete,
    currentTime,
    duration,
    selectedClips,
    addTrack,
    removeTrack,
    getTrackByType,
    addClip,
    removeClip,
    removeSelected,
    updateClip,
    updateClipTransform,
    updateClipColor,
    moveClip,
    trimClip,
    splitClipAt,
    addKeyframe,
    removeKeyframeAtTime,
    removeKeyframe,
    updateKeyframe,
    snapToClips,
    selectClip,
    clearSelection,
    setZoom,
    zoomBy,
    setCurrentTime,
    loadFromProject,
    createSnapshot,
    restoreSnapshot,
    copySelected,
    cutSelected,
    pasteSelected,
    clearAll,
    TRACK_TYPES,
  }
})
