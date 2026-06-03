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
    }
  }

  onMounted(() => {
    window.addEventListener('keydown', handleKeyDown)
  })

  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeyDown)
  })
}
