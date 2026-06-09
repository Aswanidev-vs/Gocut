<script setup>
import { ref, computed, onMounted, onUnmounted, watch, watchEffect, nextTick } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { usePlayerStore } from '../../stores/playerStore'
import { useTimelineStore, getInterpolatedProperty } from '../../stores/timelineStore'
import { Play, Pause, SkipBack, SkipForward, Maximize2, Minimize2, Volume2, VolumeX, Repeat, Loader2, ImageOff } from 'lucide-vue-next'
import { getCombinedTables } from '../../lib/curves'

const projectStore = useProjectStore()
const playerStore = usePlayerStore()
const timelineStore = useTimelineStore()

const isFullscreen = ref(false)
const playInterval = ref(null)
const videoRef = ref(null)
const useVideoElement = ref(false)  // true when <video> is active (during play)

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
function startTicker() {
  stopTicker()
  let last = performance.now()
  const tick = (now) => {
    const dt = (now - last) / 1000
    last = now
    if (playerStore.isPlaying) {
      const next = timelineStore.currentTime + dt
      if (next >= Math.max(timelineStore.duration, 0.5)) {
        if (playerStore.loop) {
          timelineStore.setCurrentTime(0)
        } else {
          playerStore.pause()
          return
        }
      } else {
        timelineStore.setCurrentTime(next)
      }
    }
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

    const clipTime = (clip.trimStart || 0) + (t - clip.startTime)
    const track = timelineStore.tracks.find(tr => tr.id === clip.trackId)
    const trackVol = (track?.muted) ? 0 : (track?.volume ?? 1)
    entry.audio.volume = Math.max(0, Math.min(1, playerStore.volume * trackVol * (clip.volume ?? 1)))

    if (playerStore.isPlaying) {
      if (Math.abs(entry.audio.currentTime - clipTime) > 0.3) {
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
}

function destroyAudioElements() {
  for (const [, entry] of audioElements) {
    entry.audio.pause()
    entry.audio.src = ''
  }
  audioElements.clear()
}

// ---- Video element sync ----

// When play/pause changes, control the video element
watch(() => playerStore.isPlaying, (playing) => {
  if (playing) {
    useVideoElement.value = !!videoSrc.value
    const v = videoRef.value
    if (v && videoSrc.value) {
      const clip = currentVisualClip.value
      if (clip) {
        const clipTime = (clip.trimStart || 0) + (timelineStore.currentTime - clip.startTime)
        v.currentTime = clipTime
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
    setTimeout(() => { useVideoElement.value = false }, 50)
    syncAudioElements()
  }
})

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
    nextTick(() => {
      const v = videoRef.value
      if (v) {
        const clip = currentVisualClip.value
        if (clip) {
          const clipTime = (clip.trimStart || 0) + (timelineStore.currentTime - clip.startTime)
          v.currentTime = clipTime
        }
        v.play().catch(() => {})
      }
    })
  }
})

// ---- Throttled preview refresh (replaces broken debounce) ----
// Uses throttle: fire immediately, then ignore for N ms. This ensures
// frames actually update during playback instead of the debounce that
// kept getting cleared by rAF every 16ms and never fired.
let lastRefreshTime = 0
let pendingRefreshRAF = null

watch(() => timelineStore.currentTime, (t) => {
  // Sync audio elements on every time change
  if (playerStore.isPlaying) syncAudioElements()

  // During playback with active video element, skip ffmpeg frame refresh
  if (playerStore.isPlaying && useVideoElement.value && videoSrc.value) return

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
})

watch(() => projectStore.project?.id, () => {
  setTimeout(() => playerStore.refreshPreview(false, timelineStore.currentTime).catch(() => {}), 100)
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
        class="preview-frame relative bg-black rounded-md overflow-hidden shadow-2xl shadow-black/50 ring-1 ring-border max-h-full"
        :style="{ ...aspectStyle, maxWidth: '100%', maxHeight: '100%', width: '100%' }"
      >
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
          muted
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
