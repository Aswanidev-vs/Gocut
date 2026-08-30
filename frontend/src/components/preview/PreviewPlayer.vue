<script setup>
import { ref, computed, reactive, onMounted, onUnmounted, watch, watchEffect, nextTick } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { usePlayerStore } from '../../stores/playerStore'
import { useTimelineStore, getInterpolatedProperty } from '../../stores/timelineStore'
import { useUiStore } from '../../stores/uiStore'
import { Play, Pause, SkipBack, SkipForward, Maximize2, Minimize2, Volume2, VolumeX, Repeat, Loader2, ImageOff } from 'lucide-vue-next'
import { getCombinedTables } from '../../lib/curves'

const projectStore = useProjectStore()
const playerStore = usePlayerStore()
const timelineStore = useTimelineStore()
const uiStore = useUiStore()

const isFullscreen = ref(false)
const playInterval = ref(null)
const videoRef = ref(null)
const useVideoElement = ref(false)  // true when <video> is active (during play)
const isTickerUpdating = ref(false)

// ---- Fetch media server port on mount ----
onMounted(() => {
  playerStore.fetchMediaServerPort()
  startTicker()
  setTimeout(() => playerStore.refreshPreview(false, timelineStore.currentTime).catch(() => {}), 100)
  document.addEventListener('fullscreenchange', onFullscreenChange)
})

onUnmounted(() => {
  stopTicker()
  destroyAudioElements()
  document.removeEventListener('fullscreenchange', onFullscreenChange)
})

function onFullscreenChange() {
  isFullscreen.value = !!document.fullscreenElement
}

// ---- Playback ticker: advance currentTime every animation frame ----
// Uses wall-clock delta for smooth advancement. The video element plays
// independently at native speed; we only correct drift when it exceeds
// a threshold. This avoids the feedback loop of reading video.currentTime
// every frame → triggering Vue watchers → causing stutter.
function getActivePlaybackSpeed() {
  const clip = currentVisualClip.value
  return (clip && clip.speed) ? clip.speed : 1.0
}

let lastAudioSyncTime = 0

function startTicker() {
  stopTicker()
  let last = performance.now()
  const tick = (now) => {
    try {
      const wallDt = (now - last) / 1000
      last = now
      if (playerStore.isPlaying) {
        const speed = getActivePlaybackSpeed()
        const dt = wallDt * speed
        const next = timelineStore.currentTime + dt
        if (next >= Math.max(timelineStore.duration, 0.5)) {
          if (playerStore.loop) {
            isTickerUpdating.value = true
            timelineStore.setCurrentTime(0)
            isTickerUpdating.value = false
          } else {
            playerStore.pause()
          }
        } else {
          isTickerUpdating.value = true
          timelineStore.setCurrentTime(next)
          isTickerUpdating.value = false
        }
        // Throttled audio sync inside ticker (every ~100ms)
        if (now - lastAudioSyncTime > 100) {
          lastAudioSyncTime = now
          syncAudioElements()
        }
      }
    } catch (err) {
      console.error('Ticker error:', err)
    }
    // Always schedule next frame — never let the loop die
    playInterval.value = requestAnimationFrame(tick)
  }
  playInterval.value = requestAnimationFrame(tick)
}

function stopTicker() {
  if (playInterval.value) {
    cancelAnimationFrame(playInterval.value)
    playInterval.value = null
  }
}

// ---- Determine current visual clip + asset at playhead ----
// Z-order: video → image → pip (PIP sits on top of everything else).
// We pick the top-most active clip so the preview reflects what the
// viewer will see in the final composition.
const visualTrackPriority = ['video', 'image', 'pip', 'text']
const currentVisualClip = computed(() => {
  const t = timelineStore.currentTime
  for (const want of visualTrackPriority) {
    for (const c of timelineStore.clips) {
      const track = timelineStore.tracks.find(tr => tr.id === c.trackId)
      if (track?.type !== want) continue
      if (t >= c.startTime && t < c.startTime + c.duration) {
        return c
      }
    }
  }
  return null
})

const currentVisualAsset = computed(() => {
  if (!currentVisualClip.value) return null
  return projectStore.getAsset(currentVisualClip.value.assetId)
})

const isVideoPlaybackAsset = computed(() => {
  return currentVisualAsset.value?.type === 'video' || currentVisualAsset.value?.type === 'gif'
})

const isImageAsset = computed(() => {
  return currentVisualAsset.value?.type === 'image'
})

// Direct image URL: the media server's /media endpoint serves the file
// straight from disk, so the WebView can render it as a normal <img>
// without going through ffmpeg. This is what the preview should show
// for still-image clips.
const imageSrc = computed(() => {
  if (!isImageAsset.value || !currentVisualAsset.value) return null
  return playerStore.getMediaUrl(currentVisualAsset.value.path)
})

const videoSrc = computed(() => {
  if (!isVideoPlaybackAsset.value || !currentVisualAsset.value) return null
  return playerStore.getMediaUrl(currentVisualAsset.value.path)
})

// Convert a timeline position into the source position used by a media
// element. A looped clip wraps inside its trimmed source range; a non-looped
// clip is clamped to the last usable source frame instead of seeking past the
// end and freezing the preview.
function getClipSourceTime(clip, asset, timelineTime, mediaDuration = 0) {
  if (!clip) return 0
  const speed = Math.max(0.001, Number(clip.speed) || 1)
  const start = Number(clip.startTime) || 0
  const trimStart = Math.max(0, Number(clip.trimStart) || 0)
  const trimEnd = Math.max(0, Number(clip.trimEnd) || 0)
  const elapsed = Math.max(0, (Number(timelineTime) || 0) - start) * speed
  const totalDuration = Math.max(0, Number(asset?.duration) || Number(mediaDuration) || 0)
  const sourceEnd = totalDuration > trimStart
    ? Math.max(trimStart, totalDuration - trimEnd)
    : 0
  const sourceSpan = sourceEnd - trimStart

  if (clip.loop && sourceSpan > 0) {
    return trimStart + (elapsed % sourceSpan)
  }
  if (sourceSpan > 0) return Math.min(trimStart + elapsed, sourceEnd)
  return trimStart + elapsed
}

const currentAudioClips = computed(() => {
  const t = timelineStore.currentTime
  return timelineStore.clips.filter(c => {
    const track = timelineStore.tracks.find(tr => tr.id === c.trackId)
    if (track?.type !== 'audio') return false
    return t >= c.startTime && t < c.startTime + c.duration
  })
})

// Manage audio elements for audio-track clips
let audioElements = new Map() // clipId -> { audio, assetPath }

function syncAudioElements() {
  try {
    const t = timelineStore.currentTime
    const activeClipIds = new Set()

    for (const clip of currentAudioClips.value) {
      activeClipIds.add(clip.id)
      const asset = projectStore.getAsset(clip.assetId)
      if (!asset) continue

      const proxyUrl = playerStore.getMediaUrl(asset.path, true) // /audio proxy -> mp3
      if (!proxyUrl) continue

      let entry = audioElements.get(clip.id)
      if (!entry || entry.assetPath !== asset.path) {
        // Create or recreate audio element
        if (entry) { entry.audio.pause(); entry.audio.src = '' }
        const audio = new Audio(proxyUrl)
        audio.preload = 'auto'
        audio.loop = !!clip.loop && !(clip.trimStart > 0 || clip.trimEnd > 0)
        // Preserve pitch when changing playback rate to avoid chipmunk effect
        audio.preservesPitch = true
        audio.addEventListener('ended', () => {
          if (!clip.loop || !playerStore.isPlaying) return
          audio.currentTime = getClipSourceTime(clip, asset, timelineStore.currentTime, audio.duration)
          audio.play().catch(() => {})
        })
        // Fallback: if proxy fails (no audio track, ffmpeg error), try raw media
        audio.addEventListener('error', () => {
          const rawUrl = playerStore.getMediaUrl(asset.path, false)
          if (rawUrl && audio.src !== rawUrl) {
            audio.src = rawUrl
            audio.load()
          }
        }, { once: true })
        entry = { audio, assetPath: asset.path }
        audioElements.set(clip.id, entry)
      }

      const speed = clip.speed || 1.0
      // Scale audio playbackRate by both its own speed and the timeline's active playback speed
      const timelineSpeed = getActivePlaybackSpeed()
      const targetPlaybackRate = speed * timelineSpeed
      if (entry.audio.playbackRate !== targetPlaybackRate) {
        entry.audio.playbackRate = targetPlaybackRate
      }
      entry.audio.loop = !!clip.loop && !(clip.trimStart > 0 || clip.trimEnd > 0)

      const clipTime = getClipSourceTime(clip, asset, t, entry.audio.duration)
      const track = timelineStore.tracks.find(tr => tr.id === clip.trackId)
      const trackVol = (track?.muted) ? 0 : (track?.volume ?? 1)
      entry.audio.volume = Math.max(0, Math.min(1, playerStore.volume * trackVol * (clip.volume ?? 1)))

      if (playerStore.isPlaying) {
        // Tighter sync threshold: at higher speeds drift accumulates faster
        const syncThreshold = 0.15
        if (Math.abs(entry.audio.currentTime - clipTime) > syncThreshold) {
          entry.audio.currentTime = clipTime
        }
        if (entry.audio.paused) entry.audio.play().catch(() => {})
      } else {
        entry.audio.pause()
        entry.audio.currentTime = clipTime
      }
    }

    // Remove audio elements for clips no longer active
    for (const [clipId, entry] of audioElements) {
      if (!activeClipIds.has(clipId)) {
        entry.audio.pause()
        entry.audio.src = ''
        audioElements.delete(clipId)
      }
    }
  } catch (err) {
    console.error('Error syncing audio elements:', err)
  }
}

function destroyAudioElements() {
  for (const [, entry] of audioElements) {
    try {
      entry.audio.pause()
      entry.audio.src = ''
    } catch (_) {}
  }
  audioElements.clear()
}

// ---- Video element sync ----

// When play/pause changes, control the video element
watch(() => playerStore.isPlaying, (playing) => {
  try {
    if (playing) {
      useVideoElement.value = !!videoSrc.value
      const v = videoRef.value
      if (v && videoSrc.value) {
        const clip = currentVisualClip.value
        if (clip) {
          const speed = clip.speed || 1.0
          const clipTime = getClipSourceTime(clip, currentVisualAsset.value, timelineStore.currentTime, v.duration)
          v.currentTime = clipTime
          v.playbackRate = speed
          v.loop = !!clip.loop && !(clip.trimStart > 0 || clip.trimEnd > 0)
          v.preservesPitch = true
          const track = timelineStore.tracks.find(tr => tr.id === clip.trackId)
          const trackVol = (track?.muted) ? 0 : (track?.volume ?? 1)
          v.volume = Math.max(0, Math.min(1, playerStore.volume * trackVol * (clip.volume ?? 1)))
        }
        v.play().catch(() => {})
      }
      syncAudioElements()
    } else {
      const v = videoRef.value
      if (v) v.pause()
      // Switch back to ffmpeg frame after short delay so we show precise frame
      playerStore.refreshPreview(true, timelineStore.currentTime).catch(() => {})
      if (!uiStore.cropMode) setTimeout(() => { useVideoElement.value = false }, 50)
      syncAudioElements()
    }
  } catch (err) {
    console.error('Error on isPlaying change:', err)
  }
})

function syncVideoElement(force = false) {
  try {
    const v = videoRef.value
    if (!v) return
    const clip = currentVisualClip.value
    if (!clip) return
    const speed = clip.speed || 1.0
    const clipTime = getClipSourceTime(clip, currentVisualAsset.value, timelineStore.currentTime, v.duration)
    v.loop = !!clip.loop && !(clip.trimStart > 0 || clip.trimEnd > 0)
    
    // Set speed
    if (v.playbackRate !== speed) {
      v.playbackRate = speed
    }
    if (v.preservesPitch !== true) {
      v.preservesPitch = true
    }

    // Avoid aggressive seeking during active playback (seeking clears buffer and stutters)
    // We let the browser's native decoder play smoothly at playbackRate.
    const threshold = force ? 0.0 : (playerStore.isPlaying ? 0.5 : 0.05)
    if (force || Math.abs(v.currentTime - clipTime) > threshold) {
      v.currentTime = clipTime
    }
  } catch (err) {
    console.error('Error syncing video element:', err)
  }
}

// Native HTML media looping always restarts at source time 0. For a looped
// clip with a trim-in/out, restart at the trimmed in-point instead.
function onVideoEnded() {
  const v = videoRef.value
  const clip = currentVisualClip.value
  if (!v || !clip?.loop || !playerStore.isPlaying) return
  if (clip.trimStart > 0 || clip.trimEnd > 0) {
    v.currentTime = getClipSourceTime(clip, currentVisualAsset.value, timelineStore.currentTime, v.duration)
    v.play().catch(() => {})
  }
}

// Sync volume changes
watch(() => playerStore.volume, () => {
  const vid = videoRef.value
  if (vid && currentVisualClip.value) {
    const clip = currentVisualClip.value
    const track = timelineStore.tracks.find(tr => tr.id === clip.trackId)
    const trackVol = (track?.muted) ? 0 : (track?.volume ?? 1)
    vid.volume = Math.max(0, Math.min(1, playerStore.volume * trackVol * (clip.volume ?? 1)))
  }
  syncAudioElements()
})

// When current video clip changes during playback (e.g. clip boundary), reload video
watch(videoSrc, (newSrc, oldSrc) => {
  if (playerStore.isPlaying && newSrc && newSrc !== oldSrc) {
    useVideoElement.value = true
    nextTick(() => {
      const v = videoRef.value
      if (v) {
        const clip = currentVisualClip.value
        if (clip) {
          const speed = clip.speed || 1.0
          const clipTime = getClipSourceTime(clip, currentVisualAsset.value, timelineStore.currentTime, v.duration)
          v.currentTime = clipTime
          v.playbackRate = speed
          v.loop = !!clip.loop && !(clip.trimStart > 0 || clip.trimEnd > 0)
        }
        v.play().catch(() => {})
      }
    })
  } else if (!newSrc) {
    useVideoElement.value = false
  }
})

// ---- Throttled preview refresh (replaces broken debounce) ----
// Uses throttle: fire immediately, then ignore for N ms. This ensures
// frames actually update during playback instead of the debounce that
// kept getting cleared by rAF every 16ms and never fired.
let lastRefreshTime = 0
let pendingRefreshRAF = null

watch(() => timelineStore.currentTime, (t) => {
  try {
    // If the ticker is updating, we skip audio sync (handled by ticker itself)
    // but we still need to sync the video element or refresh the preview.
    if (!isTickerUpdating.value) {
      // User-initiated seek: sync audio immediately
      syncAudioElements()
    }

    if (useVideoElement.value && videoSrc.value) {
      // Keep video in sync with playhead.
      // During ticker update, allow native playback to flow (drift check).
      // During manual seek (!isTickerUpdating), force-seek immediately.
      syncVideoElement(!isTickerUpdating.value)
      return
    }

    // Refresh backend/ffmpeg preview frame (e.g., for images, gaps, overlays)
    const now = performance.now()
    const interval = playerStore.isPlaying ? 100 : 50  // ms between refreshes

    if (now - lastRefreshTime >= interval) {
      lastRefreshTime = now
      playerStore.refreshPreview(false, t).catch(() => {})
    } else if (!pendingRefreshRAF) {
      // Schedule one more refresh after the throttle window
      const remaining = interval - (now - lastRefreshTime)
      pendingRefreshRAF = setTimeout(() => {
        pendingRefreshRAF = null
        lastRefreshTime = performance.now()
        playerStore.refreshPreview(false, timelineStore.currentTime).catch(() => {})
      }, remaining)
    }
  } catch (err) {
    console.error('Error in currentTime watch:', err)
  }
}, { flush: 'sync' })

// Sync immediately when clip properties (like speed, volume, etc.) change.
// A deep watch on a ref hands back the same array for in-place edits, so
// removals are detected against an id set we track ourselves.
let knownClipIds = new Set(timelineStore.clips.map(c => c.id))

watch(() => timelineStore.clips, (newClips) => {
  try {
    const ids = new Set(newClips.map(c => c.id))
    let removed = false
    for (const id of knownClipIds) {
      if (!ids.has(id)) { removed = true; break }
    }
    knownClipIds = ids

    syncAudioElements()
    syncVideoElement()

    if (removed) {
      // A removed clip must invalidate the old FFmpeg frame immediately;
      // otherwise the previous frame can remain visible while a stale
      // request finishes in the background.
      playerStore.invalidatePreview()
      playerStore.refreshPreview(true, timelineStore.currentTime).catch(() => {})
    } else {
      // Drag/trim mutates clips on every mousemove. Forcing here would bypass
      // the store in-flight guard and pile up backend renders, so let the
      // store coalesce to one active plus one queued request.
      playerStore.markPreviewStale()
      playerStore.refreshPreview(false, timelineStore.currentTime).catch(() => {})
    }
  } catch (err) {
    console.error('Error syncing clips on change:', err)
  }
}, { deep: true })

watch(() => projectStore.project?.id, () => {
  try {
    setTimeout(() => playerStore.refreshPreview(false, timelineStore.currentTime).catch(() => {}), 100)
  } catch (err) {
    console.error('Error on project ID change:', err)
  }
})

const aspectStyle = computed(() => {
  const res = projectStore.project?.resolution || { width: 1920, height: 1080 }
  return { aspectRatio: `${res.width} / ${res.height}` }
})

const projectSize = computed(() => {
  const r = projectStore.project?.resolution || { width: 1920, height: 1080 }
  return `${r.width} × ${r.height} · ${projectStore.project?.fps || 30}fps`
})

const previewSrc = computed(() => {
  if (!playerStore.previewImage) return null
  // The Go side returns a bare base64 string; we wrap as JPEG.
  return `data:image/jpeg;base64,${playerStore.previewImage}`
})

// Konva-style overlay: render text and sticker clips on top of the video frame.
const overlayItems = computed(() => {
  const t = timelineStore.currentTime
  const items = []
  for (const c of timelineStore.clips) {
    if (t < c.startTime || t >= c.startTime + c.duration) continue
    if (c.textProps) items.push({ kind: 'text', clip: c, props: c.textProps })
    if (c.stickerProps) items.push({ kind: 'sticker', clip: c, props: c.stickerProps })
  }
  return items
})

// ---- Text entry animations (time-driven CSS, mirrors ffmpeg export) ----
// Returns a partial style {opacity?, transform?, clipPath?} plus an optional
// `text` override (typewriter reveal). Identity ({}) once the animation
// window has elapsed so the rest of playback is unaffected.
const easeOutCubic = (x) => 1 - Math.pow(1 - x, 3)
const clamp01 = (x) => Math.min(1, Math.max(0, x))

function textAnimStyle(clip, tp, currentTime) {
  const anim = tp?.animation
  if (!anim || anim === 'none') return {}
  const t = currentTime - clip.startTime
  const D = (tp.animationDuration > 0 ? tp.animationDuration : 0.5)
  const u = clamp01(t / D)
  switch (anim) {
    case 'fade_in':
      return t >= D ? {} : { opacity: u }
    case 'fade_out': {
      // Animate the last D seconds before clip end: 1 -> 0.
      const rem = (clip.startTime + clip.duration - currentTime) / D
      return rem >= 1 ? {} : { opacity: clamp01(rem) }
    }
    case 'typewriter': {
      const n = Math.ceil(u * String(tp.text || '').length)
      return u >= 1 ? {} : { text: String(tp.text || '').slice(0, n) }
    }
    case 'slide_left':
    case 'slide_right':
    case 'slide_top':
    case 'slide_bottom': {
      if (t >= D) return {}
      const off = (1 - easeOutCubic(u)) * 100
      if (anim === 'slide_left') return { transform: `translateX(${off}%)` }
      if (anim === 'slide_right') return { transform: `translateX(-${off}%)` }
      if (anim === 'slide_top') return { transform: `translateY(-${off}%)` }
      return { transform: `translateY(${off}%)` }
    }
    case 'bounce': {
      if (t >= D) return {}
      const hop = Math.abs(Math.sin(u * Math.PI * 3)) * (1 - u) * 40
      return { transform: `translateY(-${hop}%)` }
    }
    case 'pop': {
      if (t >= D) return {}
      const s = u < 0.7 ? (u / 0.7) * 1.15 : 1.15 - ((u - 0.7) / 0.3) * 0.15
      return { transform: `scale(${s})`, opacity: clamp01(u * 3) }
    }
    case 'zoom_in': {
      if (t >= D) return {}
      const s = 3 - 2 * easeOutCubic(u)
      return { transform: `scale(${s})`, opacity: clamp01(u * 3) }
    }
    case 'wipe':
      return t >= D ? {} : { clipPath: `inset(0 ${(1 - u) * 100}% 0 0)` }
    default:
      return {}
  }
}

// Merge the static text style with the entry-animation style.
function textItemStyle(it) {
  const base = {
    color: it.props.color || '#FFFFFF',
    background: it.props.bgColor && it.props.bgColor !== 'transparent' ? it.props.bgColor : 'transparent',
    fontSize: (it.props.fontSize || 32) * 0.5 + 'px',
    fontWeight: it.props.bold ? 'bold' : it.props.fontWeight || 'normal',
    fontStyle: it.props.italic ? 'italic' : 'normal',
    textDecoration: it.props.underline ? 'underline' : 'none',
    padding: it.props.bgColor && it.props.bgColor !== 'transparent' ? '4px 12px' : '0',
    borderRadius: (it.props.bgBorderRadius || 0) + 'px',
    fontFamily: it.props.fontFamily || 'DM Sans',
    whiteSpace: 'pre',
    textAlign: it.props.align || 'center',
  }
  const anim = textAnimStyle(it.clip, it.props, timelineStore.currentTime)
  if (anim.transform) {
    base.transform = (base.transform ? base.transform + ' ' : '') + anim.transform
    base.transformOrigin = 'center'
  }
  const { text: _typewriterText, ...animStyle } = anim
  return { ...base, ...animStyle }
}

// Typewriter reveal: shortened text while the anim window is active.
function textItemContent(it) {
  const anim = textAnimStyle(it.clip, it.props, timelineStore.currentTime)
  return anim.text !== undefined ? anim.text : it.props.text
}

function toggleFullscreen() {
  const el = document.querySelector('.preview-frame')
  if (!el) return
  if (!document.fullscreenElement) {
    el.requestFullscreen?.().catch(() => {})
    isFullscreen.value = true
  } else {
    document.exitFullscreen?.()
    isFullscreen.value = false
  }
}

function toggleMute() {
  playerStore.setVolume(playerStore.volume > 0 ? 0 : 1.0)
}

function onScrub(e) {
  const value = parseFloat(e.target.value)
  if (!isNaN(value)) {
    timelineStore.setCurrentTime(value)
  }
}

const scrubMax = computed(() => Math.max(timelineStore.duration, 5))

const visibleClips = computed(() => {
  const t = timelineStore.currentTime
  return timelineStore.clips.filter(c => t >= c.startTime && t < c.startTime + c.duration)
})

// ---- Live CSS preview of clip color/transform/opacity ----
function getClipTransform(clip) {
  const clipTime = timelineStore.currentTime - clip.startTime;
  const def = clip.transform || {};
  return {
    x: getInterpolatedProperty(clip, 'x', clipTime, def.x || 0),
    y: getInterpolatedProperty(clip, 'y', clipTime, def.y || 0),
    scaleX: getInterpolatedProperty(clip, 'scaleX', clipTime, def.scaleX !== undefined ? def.scaleX : 1),
    scaleY: getInterpolatedProperty(clip, 'scaleY', clipTime, def.scaleY !== undefined ? def.scaleY : 1),
    rotation: getInterpolatedProperty(clip, 'rotation', clipTime, def.rotation || 0),
    opacity: getInterpolatedProperty(clip, 'opacity', clipTime, clip.opacity !== undefined ? clip.opacity : 1),
    flipH: def.flipH,
    flipV: def.flipV,
  }
}

const livePreviewStyle = computed(() => {
  const clip = currentVisualClip.value
  if (!clip) return {}

  const c = clip.color || {}
  const tf = getClipTransform(clip)

  // CSS filter chain from ColorGrade
  const filters = []
  if (c.brightness) filters.push(`brightness(${1 + c.brightness / 100})`)
  if (c.contrast) filters.push(`contrast(${1 + c.contrast / 100})`)
  if (c.saturation) filters.push(`saturate(${1 + c.saturation / 100})`)
  if (c.hue) filters.push(`hue-rotate(${c.hue}deg)`)
  if (c.blur) filters.push(`blur(${c.blur}px)`)
  if (c.sharpness) filters.push(`contrast(${1 + c.sharpness / 200})`)
  if (c.curves) filters.push(`url(#preview-curves-filter)`)

  // CSS transform from Transform
  const transforms = []
  if (tf.x || tf.y) transforms.push(`translate(${tf.x}px, ${tf.y}px)`)
  if (tf.scaleX !== 1) transforms.push(`scaleX(${tf.scaleX})`)
  if (tf.scaleY !== 1) transforms.push(`scaleY(${tf.scaleY})`)
  if (tf.rotation) transforms.push(`rotate(${tf.rotation}deg)`)
  if (tf.flipH) transforms.push('scaleX(-1)')
  if (tf.flipV) transforms.push('scaleY(-1)')

  const currentTime = timelineStore.currentTime
  const clipTime = currentTime - clip.startTime
  const trans = clip.transition
  let transOpacity = 1
  let transFilter = ''
  let transTransform = ''
  let transClipPath = ''

  if (trans && trans.type !== 'none' && trans.duration > 0 && clipTime < trans.duration && clipTime >= 0) {
    const p = clipTime / trans.duration // 0.0 to 1.0

    switch (trans.type) {
      case 'fade':
      case 'dissolve':
        transOpacity = p
        break
      case 'wipeleft':
        transClipPath = `inset(0 0 0 ${100 - p * 100}%)`
        break
      case 'wiperight':
        transClipPath = `inset(0 ${100 - p * 100}% 0 0)`
        break
      case 'slideleft':
        transTransform = `translateX(${100 - p * 100}%)`
        break
      case 'slideright':
        transTransform = `translateX(-${100 - p * 100}%)`
        break
      case 'zoomin':
        transTransform = `scale(${0.5 + p * 0.5})`
        transOpacity = p
        break
      case 'hflip':
        transTransform = `rotateY(${90 - p * 90}deg)`
        break
      case 'circleopen':
        transClipPath = `circle(${p * 100}% at center)`
        break
      case 'pixelize':
      case 'blur':
        transFilter = `blur(${(1 - p) * 20}px)`
        transOpacity = p
        break
    }
  }

  const style = {}
  if (filters.length || transFilter) {
    style.filter = (filters.join(' ') + ' ' + transFilter).trim()
  }
  if (transforms.length || transTransform) {
    style.transform = (transforms.join(' ') + ' ' + transTransform).trim()
  }
  const finalOpacity = tf.opacity * transOpacity
  if (finalOpacity < 1) style.opacity = finalOpacity
  if (transClipPath) style.clipPath = transClipPath

  return style
})

const curvesSvgTables = computed(() => {
  const clip = currentVisualClip.value
  if (!clip || !clip.color || !clip.color.curves) return null
  return getCombinedTables(clip.color.curves)
})

function exitFullscreen() {
  document.exitFullscreen?.().catch(() => {})
}

// ---- Interactive crop overlay ----
// Lets the user drag a crop rectangle directly on the preview. The rectangle
// is stored on clip.transform.crop* in source pixels (consumed by the export
// pipeline); here we map it to the on-screen fitted media box and back.
const previewFrameRef = ref(null)
const frameSize = reactive({ w: 0, h: 0 })
let frameObserver = null

function measureFrame() {
  const el = previewFrameRef.value
  if (!el) return
  const r = el.getBoundingClientRect()
  frameSize.w = r.width
  frameSize.h = r.height
}

onMounted(() => {
  measureFrame()
  if (previewFrameRef.value && typeof ResizeObserver !== 'undefined') {
    frameObserver = new ResizeObserver(() => measureFrame())
    frameObserver.observe(previewFrameRef.value)
  }
  window.addEventListener('resize', measureFrame)
})
onUnmounted(() => {
  frameObserver?.disconnect()
  window.removeEventListener('resize', measureFrame)
})
watch(() => uiStore.cropMode, (enabled, wasEnabled) => {
  nextTick(() => {
    measureFrame()
    // Cropping uses the direct video element so the handles stay aligned with
    // the media. Return to the normal still-frame preview when editing ends.
    if (!enabled && wasEnabled && !playerStore.isPlaying && videoSrc.value) {
      setTimeout(() => { useVideoElement.value = false }, 50)
    }
  })
})

function onCropKeydown(e) {
  if (e.key === 'Escape' && uiStore.cropMode) {
    uiStore.setCropMode(false)
  }
}
window.addEventListener('keydown', onCropKeydown)
onUnmounted(() => window.removeEventListener('keydown', onCropKeydown))

const cropClip = computed(() => {
  const sel = timelineStore.selectedClips[0]
  if (!sel) return null
  const t = timelineStore.tracks.find(tr => tr.id === sel.trackId)
  if (!t) return null
  if (t.type !== 'video' && t.type !== 'image' && t.type !== 'pip') return null
  if (currentVisualClip.value?.id !== sel.id) return null
  return sel
})
const cropAsset = computed(() => (cropClip.value ? projectStore.getAsset(cropClip.value.assetId) : null))
const cropActive = computed(() =>
  uiStore.cropMode && !!cropClip.value && !!cropAsset.value &&
  cropAsset.value.width > 0 && cropAsset.value.height > 0
)
const cropIsSet = computed(() => {
  const tf = cropClip.value?.transform || {}
  return Number(tf.cropW) > 0 && Number(tf.cropH) > 0
})
const cropHint = computed(() => {
  if (!uiStore.cropMode) return null
  const sel = timelineStore.selectedClips[0]
  if (!sel) return 'Select an image or video clip to crop'
  const t = timelineStore.tracks.find(tr => tr.id === sel.trackId)
  if (!t || (t.type !== 'video' && t.type !== 'image' && t.type !== 'pip')) return 'Crop works on image or video clips'
  if (currentVisualClip.value?.id !== sel.id) return 'Scrub the playhead onto the clip to crop it'
  return 'Drag corners to crop · Esc to finish'
})

// Source crop rect (px); defaults to the full media when none is set.
const cropSourceRect = computed(() => {
  const clip = cropClip.value
  const a = cropAsset.value
  if (!clip || !a) return { x: 0, y: 0, w: 0, h: 0 }
  const tf = clip.transform || {}
  let x = tf.cropX || 0, y = tf.cropY || 0, w = tf.cropW || 0, h = tf.cropH || 0
  if (w <= 0 || h <= 0) { x = 0; y = 0; w = a.width; h = a.height }
  w = Math.min(a.width, Math.max(1, w))
  h = Math.min(a.height, Math.max(1, h))
  x = Math.max(0, Math.min(a.width - w, x))
  y = Math.max(0, Math.min(a.height - h, y))
  return { x, y, w, h }
})

function clipTransformParams() {
  const tf = cropClip.value?.transform || {}
  return {
    tx: tf.x || 0, ty: tf.y || 0,
    sx: tf.scaleX !== undefined ? tf.scaleX : 1,
    sy: tf.scaleY !== undefined ? tf.scaleY : 1,
    flipH: !!tf.flipH, flipV: !!tf.flipV,
    rot: (tf.rotation || 0) * Math.PI / 180,
  }
}

// Map a source-pixel point to on-screen frame pixels, mirroring the
// object-contain fit + the clip transform applied by livePreviewStyle.
function sourceToFrame(p) {
  const a = cropAsset.value, FW = frameSize.w, FH = frameSize.h
  if (!a || !FW || !FH) return { x: 0, y: 0 }
  const k = Math.min(FW / a.width, FH / a.height)
  const ox = (FW - a.width * k) / 2, oy = (FH - a.height * k) / 2
  let mx = ox + p.x * k, my = oy + p.y * k
  const C = { x: FW / 2, y: FH / 2 }
  let cx = mx - C.x, cy = my - C.y
  const P = clipTransformParams()
  if (P.flipH) cx = -cx
  if (P.flipV) cy = -cy
  const cos = Math.cos(P.rot), sin = Math.sin(P.rot)
  const rx = cx * cos - cy * sin, ry = cx * sin + cy * cos
  cx = rx * P.sx; cy = ry * P.sy
  cx += P.tx; cy += P.ty
  return { x: C.x + cx, y: C.y + cy }
}

function frameToSource(p) {
  const a = cropAsset.value, FW = frameSize.w, FH = frameSize.h
  if (!a || !FW || !FH) return { x: 0, y: 0 }
  const k = Math.min(FW / a.width, FH / a.height)
  const ox = (FW - a.width * k) / 2, oy = (FH - a.height * k) / 2
  const C = { x: FW / 2, y: FH / 2 }
  let cx = p.x - C.x, cy = p.y - C.y
  const P = clipTransformParams()
  cx -= P.tx; cy -= P.ty
  cx /= Math.abs(P.sx) > 0.001 ? P.sx : 1
  cy /= Math.abs(P.sy) > 0.001 ? P.sy : 1
  const cos = Math.cos(-P.rot), sin = Math.sin(-P.rot)
  const rx = cx * cos - cy * sin, ry = cx * sin + cy * cos
  cx = rx; cy = ry
  if (P.flipH) cx = -cx
  if (P.flipV) cy = -cy
  return { x: (C.x + cx - ox) / k, y: (C.y + cy - oy) / k }
}

// Crop rect on screen (axis-aligned bounding box of the transformed source rect).
const cropRectPx = computed(() => {
  if ((!cropActive.value && !cropIsSet.value) || !cropClip.value || !cropAsset.value) return null
  const r = cropSourceRect.value
  const pts = [
    sourceToFrame({ x: r.x, y: r.y }),
    sourceToFrame({ x: r.x + r.w, y: r.y }),
    sourceToFrame({ x: r.x, y: r.y + r.h }),
    sourceToFrame({ x: r.x + r.w, y: r.y + r.h }),
  ]
  const xs = pts.map(p => p.x), ys = pts.map(p => p.y)
  return { x: Math.min(...xs), y: Math.min(...ys), w: Math.max(...xs) - Math.min(...xs), h: Math.max(...ys) - Math.min(...ys) }
})

const cropRenderGeometry = computed(() => {
  const r = cropRectPx.value
  if (!r || !cropIsSet.value || uiStore.cropMode || !frameSize.w || !frameSize.h || r.w <= 0 || r.h <= 0) return null

  // The source crop is currently displayed at the fitted-media scale. Fit
  // that selected rectangle back into the project frame, just as the render
  // pipeline does after its source-space crop filter runs.
  const fit = Math.min(frameSize.w / r.w, frameSize.h / r.h)
  const centerX = frameSize.w / 2
  const centerY = frameSize.h / 2
  const cropCenterX = r.x + r.w / 2
  const cropCenterY = r.y + r.h / 2
  const viewport = {
    x: centerX - r.w * fit / 2,
    y: centerY - r.h * fit / 2,
    w: r.w * fit,
    h: r.h * fit,
  }

  return {
    fit,
    viewport,
    offsetX: centerX - (centerX + (cropCenterX - centerX) * fit),
    offsetY: centerY - (centerY + (cropCenterY - centerY) * fit),
  }
})

// Once crop editing is finished, keep the cropped media centred and fitted in
// the project frame, matching the render pipeline's centred overlay behavior.
// During editing the full image remains visible under the dimmed mask so the
// user can position the crop against surrounding context.
const cropViewportStyle = computed(() => {
  const geometry = cropRenderGeometry.value
  if (!geometry) return {}
  const r = geometry.viewport
  const right = frameSize.w - r.x - r.w
  const bottom = frameSize.h - r.y - r.h
  return {
    clipPath: `inset(${Math.max(0, r.y)}px ${Math.max(0, right)}px ${Math.max(0, bottom)}px ${Math.max(0, r.x)}px)`,
  }
})

const cropMediaOffsetStyle = computed(() => {
  const geometry = cropRenderGeometry.value
  if (!geometry) return {}
  return { transform: `translate(${geometry.offsetX}px, ${geometry.offsetY}px) scale(${geometry.fit})` }
})

const cropHandles = ['nw', 'ne', 'sw', 'se']
function handleStyle(h) {
  const m = { nw: ['0%', '0%'], ne: ['100%', '0%'], sw: ['0%', '100%'], se: ['100%', '100%'] }[h]
  return { left: m[0], top: m[1], cursor: (h === 'nw' || h === 'se') ? 'nwse-resize' : 'nesw-resize' }
}

// ---- Drag interaction ----
let drag = null
function clampSrc(p) {
  const a = cropAsset.value
  if (!a) return { x: 0, y: 0 }
  return { x: Math.max(0, Math.min(a.width, p.x)), y: Math.max(0, Math.min(a.height, p.y)) }
}
function beginCropDrag(mode, e, sourceStart = null) {
  if (!cropActive.value) return
  e.preventDefault(); e.stopPropagation()
  measureFrame()
  const r = cropSourceRect.value
  const fr = previewFrameRef.value.getBoundingClientRect()
  const start = sourceStart || frameToSource({ x: e.clientX - fr.left, y: e.clientY - fr.top })
  drag = {
    mode,
    fr,
    corners: {
      nw: { x: r.x, y: r.y }, ne: { x: r.x + r.w, y: r.y },
      sw: { x: r.x, y: r.y + r.h }, se: { x: r.x + r.w, y: r.y + r.h },
    },
    start: { x: e.clientX, y: e.clientY },
    startSource: clampSrc(start),
    moved: false,
  }
  window.addEventListener('pointermove', onCropPointerMove)
  window.addEventListener('pointerup', onCropPointerUp)
  window.addEventListener('pointercancel', onCropPointerUp)
}

function onCropPointerDown(mode, e) {
  beginCropDrag(mode, e)
}

function onCropCanvasPointerDown(e) {
  // The selection and handles stop propagation. A press on the remaining
  // preview starts a brand-new free-form crop rectangle from that point.
  if (e.target !== e.currentTarget) return
  const fr = previewFrameRef.value?.getBoundingClientRect()
  if (!fr) return
  const start = clampSrc(frameToSource({ x: e.clientX - fr.left, y: e.clientY - fr.top }))
  beginCropDrag('create', e, start)
}

function onPreviewPointerDown(e) {
  if (e.target.closest?.('button, input')) return
  if (!cropClip.value || uiStore.cropMode) return
  e.preventDefault()
  uiStore.setCropMode(true)
}
function onCropPointerMove(e) {
  if (!drag) return
  const cur = { x: e.clientX - drag.fr.left, y: e.clientY - drag.fr.top }
  const a = cropAsset.value
  if (!a) return
  const c = drag.corners
  let nw, se
  if (drag.mode === 'move') {
    const startSrc = frameToSource({ x: drag.start.x - drag.fr.left, y: drag.start.y - drag.fr.top })
    const curSrc = frameToSource(cur)
    const dx = curSrc.x - startSrc.x, dy = curSrc.y - startSrc.y
    nw = { x: c.nw.x + dx, y: c.nw.y + dy }
    se = { x: c.se.x + dx, y: c.se.y + dy }
    const w = se.x - nw.x, h = se.y - nw.y
    nw.x = Math.max(0, Math.min(a.width - w, nw.x))
    nw.y = Math.max(0, Math.min(a.height - h, nw.y))
    se = { x: nw.x + w, y: nw.y + h }
  } else if (drag.mode === 'create') {
    const cp = clampSrc(frameToSource(cur))
    nw = { x: Math.min(drag.startSource.x, cp.x), y: Math.min(drag.startSource.y, cp.y) }
    se = { x: Math.max(drag.startSource.x, cp.x), y: Math.max(drag.startSource.y, cp.y) }
  } else {
    const cp = clampSrc(frameToSource(cur))
    const opp = { nw: 'se', ne: 'sw', sw: 'ne', se: 'nw' }[drag.mode]
    const fixed = c[opp]
    const minW = Math.min(2, a.width)
    const minH = Math.min(2, a.height)
    if (drag.mode === 'nw') {
      nw = {
        x: Math.max(0, Math.min(cp.x, fixed.x - minW)),
        y: Math.max(0, Math.min(cp.y, fixed.y - minH)),
      }
      se = { ...fixed }
    } else if (drag.mode === 'ne') {
      nw = {
        x: fixed.x,
        y: Math.max(0, Math.min(cp.y, fixed.y - minH)),
      }
      se = {
        x: Math.min(a.width, Math.max(cp.x, fixed.x + minW)),
        y: fixed.y,
      }
    } else if (drag.mode === 'sw') {
      nw = {
        x: Math.max(0, Math.min(cp.x, fixed.x - minW)),
        y: fixed.y,
      }
      se = {
        x: fixed.x,
        y: Math.min(a.height, Math.max(cp.y, fixed.y + minH)),
      }
    } else {
      nw = { ...fixed }
      se = {
        x: Math.min(a.width, Math.max(cp.x, fixed.x + minW)),
        y: Math.min(a.height, Math.max(cp.y, fixed.y + minH)),
      }
    }
  }
  drag.moved = true
  timelineStore.updateClipTransform(cropClip.value.id, {
    cropX: Math.round(nw.x), cropY: Math.round(nw.y),
    cropW: Math.max(1, Math.round(se.x - nw.x)), cropH: Math.max(1, Math.round(se.y - nw.y)),
  })
}
function onCropPointerUp() {
  drag = null
  window.removeEventListener('pointermove', onCropPointerMove)
  window.removeEventListener('pointerup', onCropPointerUp)
  window.removeEventListener('pointercancel', onCropPointerUp)
}

// Force the direct <video> element (single clip) while cropping so the
// overlay aligns with the fitted media box rather than the ffmpeg composite.
watch(cropActive, (on) => {
  if (on && isVideoPlaybackAsset.value) {
    useVideoElement.value = true
    nextTick(() => { if (videoRef.value) videoRef.value.pause() })
  }
})
</script>

<template>
  <div class="flex-1 flex flex-col bg-bg/50 min-h-0 overflow-hidden relative">
    <svg v-if="curvesSvgTables" width="0" height="0" class="absolute pointer-events-none">
      <filter id="preview-curves-filter" color-interpolation-filters="sRGB">
        <feComponentTransfer>
          <feFuncR type="table" :tableValues="curvesSvgTables.r" />
          <feFuncG type="table" :tableValues="curvesSvgTables.g" />
          <feFuncB type="table" :tableValues="curvesSvgTables.b" />
        </feComponentTransfer>
      </filter>
    </svg>

    <!-- Aspect-ratio preview frame -->
    <div class="flex-1 flex items-center justify-center p-4 min-h-0">
      <div
        ref="previewFrameRef"
        class="preview-frame relative bg-black rounded-md overflow-hidden shadow-2xl shadow-black/50 ring-1 ring-border max-h-full"
        :style="{ ...aspectStyle, maxWidth: '100%', maxHeight: '100%', width: '100%' }"
        @pointerdown="onPreviewPointerDown"
      >
        <!-- The crop viewport clips the visible media after editing. During
             editing it remains transparent so the dimmed crop mask can show
             the surrounding image for accurate framing. -->
        <div class="absolute inset-0" :style="cropViewportStyle">
          <div class="absolute inset-0" :style="cropMediaOffsetStyle">
            <!-- Direct <img> for still image clips: served straight from
                 the media server so it does not require an ffmpeg roundtrip. -->
            <img
              v-show="isImageAsset && imageSrc && !useVideoElement"
              :src="imageSrc"
              class="absolute inset-0 w-full h-full object-contain bg-black"
              :style="livePreviewStyle"
              alt="image preview"
            />

            <!-- Live <video> element for playback (video only, audio proxy handles sound) -->
            <video
              v-show="useVideoElement && videoSrc"
              ref="videoRef"
              :src="videoSrc"
              class="absolute inset-0 w-full h-full object-contain bg-black"
              :style="livePreviewStyle"
              preload="auto"
              playsinline
              @ended="onVideoEnded"
              @error="() => { /* fallback to ffmpeg frame on error */ useVideoElement = false }"
            />

            <!-- Static frame from ffmpeg (shown when paused or no video element) -->
            <!-- For non-image assets, this is the ffmpeg-extracted frame. -->
            <!-- For image assets, the direct <img> above is preferred. -->
            <img
              v-show="!useVideoElement && previewSrc"
              :src="previewSrc"
              class="absolute inset-0 w-full h-full object-contain bg-black"
              :style="livePreviewStyle"
              alt="preview"
            />
          </div>
        </div>

        <!-- Interactive crop overlay (image & video clips) -->
        <div v-if="cropActive" class="absolute inset-0 z-30" style="touch-action: none;" @pointerdown="onCropCanvasPointerDown">
          <div
            v-if="cropRectPx"
            class="absolute pointer-events-none"
            :style="{ left: cropRectPx.x + 'px', top: cropRectPx.y + 'px', width: cropRectPx.w + 'px', height: cropRectPx.h + 'px', boxShadow: '0 0 0 9999px rgba(0,0,0,0.55)' }"
          ></div>
          <div
            v-if="cropRectPx"
            data-crop-selection
            class="absolute cursor-move border-2 border-accent"
            :style="{ left: cropRectPx.x + 'px', top: cropRectPx.y + 'px', width: cropRectPx.w + 'px', height: cropRectPx.h + 'px' }"
            @pointerdown="onCropPointerDown('move', $event)"
          >
            <div
              v-for="h in cropHandles"
              :key="h"
              class="absolute w-3 h-3 bg-accent border border-white rounded-sm -translate-x-1/2 -translate-y-1/2"
              :style="handleStyle(h)"
              @pointerdown.stop="onCropPointerDown(h, $event)"
            ></div>
          </div>
        </div>
        <div
          v-if="cropHint"
          class="absolute bottom-3 left-1/2 -translate-x-1/2 z-30 px-3 py-1.5 rounded-full bg-black/70 text-white text-[11px] flex items-center gap-2"
        >
          <span>{{ cropHint }}</span>
          <button
            v-if="cropActive"
            class="px-1.5 py-0.5 rounded bg-white/10 hover:bg-white/20 text-[10px]"
            @pointerdown.stop
            @click.stop="uiStore.setCropMode(false)"
          >Done</button>
        </div>

        <div
          v-show="!useVideoElement && !previewSrc && !imageSrc"
          class="absolute inset-0 flex flex-col items-center justify-center text-text-secondary text-sm gap-2"
        >
          <ImageOff :size="32" class="opacity-40" />
          <div v-if="playerStore.isPreviewLoading" class="flex items-center gap-2 text-xs">
            <Loader2 :size="14" class="animate-spin" />
            Loading frame…
          </div>
          <div v-else-if="playerStore.previewError" class="text-xs text-red-400 max-w-md text-center px-4">
            {{ playerStore.previewError }}
          </div>
          <div v-else class="text-xs">
            {{ (projectStore.project?.assets || []).length === 0 ? 'Import media to see preview' : 'No clip at playhead' }}
          </div>
        </div>

        <!-- Text overlay (no Konva, just CSS) -->
        <div class="absolute inset-0 pointer-events-none">
          <div
            v-for="(it, idx) in overlayItems"
            :key="idx"
            class="absolute"
            :style="{
              left: '50%',
              top: '50%',
              opacity: getClipTransform(it.clip).opacity,
              transform: `translate(calc(-50% + ${getClipTransform(it.clip).x}px), calc(-50% + ${getClipTransform(it.clip).y}px)) ` +
                         `scale(${getClipTransform(it.clip).scaleX}, ${getClipTransform(it.clip).scaleY}) ` +
                         `rotate(${getClipTransform(it.clip).rotation}deg) ` +
                         `scaleX(${getClipTransform(it.clip).flipH ? -1 : 1}) ` +
                         `scaleY(${getClipTransform(it.clip).flipV ? -1 : 1})`,
            }"
          >
            <div
              v-if="it.kind === 'text'"
              :style="{
                color: it.props.color || '#FFFFFF',
                background: it.props.bgColor && it.props.bgColor !== 'transparent' ? it.props.bgColor : 'transparent',
                fontSize: (it.props.fontSize || 32) * 0.5 + 'px',
                fontWeight: it.props.bold ? 'bold' : it.props.fontWeight || 'normal',
                fontStyle: it.props.italic ? 'italic' : 'normal',
                textDecoration: it.props.underline ? 'underline' : 'none',
                padding: it.props.bgColor && it.props.bgColor !== 'transparent' ? '4px 12px' : '0',
                borderRadius: (it.props.bgBorderRadius || 0) + 'px',
                fontFamily: it.props.fontFamily || 'DM Sans',
                whiteSpace: 'pre',
                textAlign: it.props.align || 'center',
              }"
            >
              {{ it.props.text }}
            </div>
            <div
              v-else-if="it.kind === 'sticker'"
              :style="{
                opacity: it.props.opacity ?? 1,
                transform: `scale(${it.props.width || 0.2}) rotate(${it.props.rotation || 0}deg)`,
                transformOrigin: 'center',
              }"
              v-html="it.props.svg"
            />
          </div>
        </div>

        <!-- Exit fullscreen button -->
        <button
          v-if="isFullscreen"
          class="absolute top-4 right-4 z-50 p-2.5 rounded-lg bg-black/60 hover:bg-black/80 text-white backdrop-blur-sm transition-all opacity-40 hover:opacity-100 focus:opacity-100"
          @click="exitFullscreen"
          title="Exit Fullscreen (Esc)"
        >
          <Minimize2 :size="20" />
        </button>
      </div>
    </div>

    <!-- Transport controls -->
    <div class="px-3 py-2 flex items-center gap-2 border-t border-border bg-panel flex-shrink-0">
      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="playerStore.stepBackward()"
        title="Previous frame (←)"
      >
        <SkipBack :size="14" />
      </button>

      <button
        class="p-2 rounded-full bg-accent text-bg hover:bg-accent-hover transition-colors"
        @click="playerStore.togglePlay()"
        :title="playerStore.isPlaying ? 'Pause (Space)' : 'Play (Space)'"
      >
        <Pause v-if="playerStore.isPlaying" :size="14" />
        <Play v-else :size="14" />
      </button>

      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="playerStore.stepForward()"
        title="Next frame (→)"
      >
        <SkipForward :size="14" />
      </button>

      <div class="font-mono text-xs text-text-secondary w-32 text-center">
        {{ playerStore.formattedTime }} / {{ timelineStore.duration ? playerStore.formatTimecode(timelineStore.duration) : '00:00:00.000' }}
      </div>

      <div class="flex-1 px-2">
        <input
          type="range"
          :min="0"
          :max="scrubMax"
          step="0.01"
          :value="timelineStore.currentTime"
          @input="onScrub"
          class="w-full h-1 bg-border rounded appearance-none cursor-pointer accent-accent"
        />
      </div>

      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="toggleMute"
        :title="playerStore.volume > 0 ? 'Mute' : 'Unmute'"
      >
        <Volume2 v-if="playerStore.volume > 0" :size="14" />
        <VolumeX v-else :size="14" />
      </button>

      <button
        class="p-1.5 rounded transition-colors"
        :class="playerStore.loop ? 'text-accent' : 'text-text-secondary hover:text-text-primary hover:bg-border'"
        @click="playerStore.toggleLoop()"
        title="Loop"
      >
        <Repeat :size="14" />
      </button>

      <button
        class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
        @click="toggleFullscreen"
        title="Fullscreen"
      >
        <Maximize2 :size="14" />
      </button>

      <div class="text-[10px] text-text-secondary font-mono hidden md:block">{{ projectSize }}</div>
    </div>
  </div>
</template>
