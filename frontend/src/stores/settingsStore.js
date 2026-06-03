import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const DEFAULT_SHORTCUTS = {
  undo: 'ctrl+z',
  redo: 'ctrl+y',
  cut: 'ctrl+x',
  copy: 'ctrl+c',
  paste: 'ctrl+v',
}

export const useSettingsStore = defineStore('settings', () => {
  const shortcuts = ref({ ...DEFAULT_SHORTCUTS })

  // Load from localStorage on startup
  try {
    const saved = localStorage.getItem('gocut-shortcuts')
    if (saved) {
      shortcuts.value = { ...DEFAULT_SHORTCUTS, ...JSON.parse(saved) }
    }
  } catch (e) {
    console.error('Failed to load shortcuts', e)
  }

  // Save to localStorage whenever it changes
  watch(shortcuts, (newVal) => {
    localStorage.setItem('gocut-shortcuts', JSON.stringify(newVal))
  }, { deep: true })

  function updateShortcut(action, keyCombo) {
    if (shortcuts.value.hasOwnProperty(action)) {
      shortcuts.value[action] = keyCombo.toLowerCase()
    }
  }

  function resetToDefaults() {
    shortcuts.value = { ...DEFAULT_SHORTCUTS }
  }

  return {
    shortcuts,
    updateShortcut,
    resetToDefaults
  }
})
