import { onMounted, onUnmounted } from 'vue'
import { useSettingsStore } from '../stores/settingsStore'
import { useHistoryStore } from '../stores/historyStore'
import { useTimelineStore } from '../stores/timelineStore'

function getComboString(e) {
  let combo = []
  if (e.ctrlKey || e.metaKey) combo.push('ctrl')
  if (e.shiftKey) combo.push('shift')
  if (e.altKey) combo.push('alt')

  if (e.key !== 'Control' && e.key !== 'Shift' && e.key !== 'Alt' && e.key !== 'Meta') {
    combo.push(e.key.toLowerCase())
  }

  return combo.join('+')
}

// Shuttle state: track the current shuttle speed
let shuttleInterval = null
let shuttleSpeed = 0

export function useHotkeys() {
  const handleKeyDown = (e) => {
    // Ignore hotkeys when typing in inputs
    const tag = e.target.tagName.toLowerCase()
    if (tag === 'input' || tag === 'textarea' || e.target.isContentEditable) {
      return
    }

    const combo = getComboString(e)
    if (!combo) return

    const settings = useSettingsStore()
    const history = useHistoryStore()
    const timeline = useTimelineStore()

    switch (combo) {
      case settings.shortcuts.undo:
        e.preventDefault()
        history.undo()
        break
      case settings.shortcuts.redo:
        e.preventDefault()
        history.redo()
        break
      case settings.shortcuts.cut:
        e.preventDefault()
        timeline.cutSelected()
        break
      case settings.shortcuts.copy:
        e.preventDefault()
        timeline.copySelected()
        break
      case settings.shortcuts.paste:
        e.preventDefault()
        timeline.pasteSelected()
        break
      case settings.shortcuts.duplicate:
        e.preventDefault()
        duplicateSelectedClips(timeline)
        break
      case settings.shortcuts.export:
        e.preventDefault()
        openExportDialog()
        break
      // J/K/L shuttle controls
      case settings.shortcuts.shuttleReverse:
        e.preventDefault()
        startShuttle(-1)
        break
      case settings.shortcuts.shuttleStop:
        e.preventDefault()
        stopShuttle()
        break
      case settings.shortcuts.shuttleForward:
        e.preventDefault()
        startShuttle(1)
        break
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyDown)
    stopShuttle()
  })
}

function duplicateSelectedClips(timeline) {
  const selected = timeline.selectedClips
  if (selected.length === 0) return

  const newClips = selected.map(clip => ({
    ...clip,
    id: undefined,
    startTime: clip.startTime + clip.duration + 0.1,
  }))

  newClips.forEach(clip => timeline.addClip(clip))
}

function openExportDialog() {
  window.dispatchEvent(new CustomEvent('shortcut:export'))
}

function startShuttle(direction) {
  stopShuttle()

  const playerStore = getPlayerStore()
  if (!playerStore) return

  // Set initial shuttle speed
  shuttleSpeed = direction

  // Apply shuttle immediately
  applyShuttle(playerStore, shuttleSpeed)

  // Continue shuttling
  shuttleInterval = setInterval(() => {
    applyShuttle(playerStore, shuttleSpeed)
  }, 100)
}

function applyShuttle(playerStore, speed) {
  if (!playerStore) return

  // Step forward/backward based on direction
  const step = speed * (1 / 30) // 30fps step
  const newTime = Math.max(0, playerStore.currentTime + step)
  playerStore.seek(newTime)
}

function stopShuttle() {
  if (shuttleInterval) {
    clearInterval(shuttleInterval)
    shuttleInterval = null
  }
  shuttleSpeed = 0
}

function getPlayerStore() {
  try {
    const pinia = window.__pinia
    if (!pinia) return null
    return pinia._s && pinia._s.get('player') ? pinia._s.get('player') : null
  } catch (_) {
    return null
  }
}
