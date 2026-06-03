<template>
  <div v-if="uiStore.isSettingsDialogOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-bg/70 backdrop-blur-sm" @click.self="close">
    <div class="w-[600px] h-[400px] bg-panel border border-border rounded-md shadow-2xl flex overflow-hidden">
      
      <!-- Sidebar -->
      <div class="w-48 border-r border-border bg-bg/50 p-4">
        <h3 class="text-sm font-semibold text-text-primary mb-4">Settings</h3>
        <div class="flex flex-col gap-1">
          <button 
            class="px-3 py-2 text-sm text-left rounded transition-colors"
            :class="activeTab === 'shortcuts' ? 'bg-accent/10 text-accent font-medium' : 'text-text-secondary hover:bg-border/50'"
            @click="activeTab = 'shortcuts'"
          >
            Keyboard Shortcuts
          </button>
        </div>
      </div>

      <!-- Main Content -->
      <div class="flex-1 flex flex-col">
        <div class="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 class="text-lg font-medium text-text-primary">
            {{ activeTab === 'shortcuts' ? 'Keyboard Shortcuts' : 'Settings' }}
          </h2>
          <button class="p-1 rounded hover:bg-border text-text-secondary transition-colors" @click="close">
            <X :size="16" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto p-6">
          <div v-if="activeTab === 'shortcuts'" class="space-y-4">
            
            <div class="flex justify-between items-center mb-6">
              <span class="text-sm text-text-secondary">Customize your editor hotkeys. Click a shortcut to change it.</span>
              <button class="text-xs text-accent hover:underline" @click="settings.resetToDefaults">Reset to Defaults</button>
            </div>

            <div v-for="(label, action) in actionLabels" :key="action" class="flex items-center justify-between py-2 border-b border-border/50">
              <span class="text-sm text-text-primary">{{ label }}</span>
              <button 
                class="px-3 py-1.5 min-w-[80px] rounded text-xs font-mono border transition-colors"
                :class="recordingAction === action ? 'bg-accent text-bg border-accent animate-pulse' : 'bg-bg border-border text-text-secondary hover:border-text-secondary'"
                @click="startRecording(action)"
              >
                {{ recordingAction === action ? 'Press keys...' : formatCombo(settings.shortcuts[action]) }}
              </button>
            </div>

          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { X } from 'lucide-vue-next'
import { useUiStore } from '../../stores/uiStore'
import { useSettingsStore } from '../../stores/settingsStore'

const uiStore = useUiStore()
const settings = useSettingsStore()

const activeTab = ref('shortcuts')
const recordingAction = ref(null)

const actionLabels = {
  undo: 'Undo',
  redo: 'Redo',
  cut: 'Cut',
  copy: 'Copy',
  paste: 'Paste',
}

function close() {
  uiStore.isSettingsDialogOpen = false
  recordingAction.value = null
}

function startRecording(action) {
  recordingAction.value = action
}

function handleKeyDown(e) {
  if (!recordingAction.value) return

  // Prevent default behavior while recording to avoid triggering actual shortcuts
  e.preventDefault()
  
  if (e.key === 'Escape') {
    recordingAction.value = null
    return
  }

  // Only record if a non-modifier key is pressed
  if (e.key !== 'Control' && e.key !== 'Shift' && e.key !== 'Alt' && e.key !== 'Meta') {
    let combo = []
    if (e.ctrlKey || e.metaKey) combo.push('ctrl')
    if (e.shiftKey) combo.push('shift')
    if (e.altKey) combo.push('alt')
    combo.push(e.key.toLowerCase())

    settings.updateShortcut(recordingAction.value, combo.join('+'))
    recordingAction.value = null
  }
}

function formatCombo(combo) {
  if (!combo) return ''
  return combo.split('+').map(k => {
    if (k === 'ctrl') return 'Ctrl'
    if (k === 'alt') return 'Alt'
    if (k === 'shift') return 'Shift'
    return k.toUpperCase()
  }).join(' + ')
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>
