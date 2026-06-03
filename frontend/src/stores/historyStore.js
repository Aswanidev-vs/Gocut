import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useTimelineStore } from './timelineStore'
import { useUiStore } from './uiStore'

export const useHistoryStore = defineStore('history', () => {
  const undoStack = ref([])
  const redoStack = ref([])
  const isExecuting = ref(false)
  const MAX_HISTORY = 50

  function pushSnapshot() {
    if (isExecuting.value) return // Don't snapshot while undoing/redoing

    const timelineStore = useTimelineStore()
    const snapshot = timelineStore.createSnapshot()
    
    // Don't push if nothing changed
    if (undoStack.value.length > 0) {
      const lastSnapshot = undoStack.value[undoStack.value.length - 1]
      if (JSON.stringify(lastSnapshot) === JSON.stringify(snapshot)) {
        return
      }
    }

    undoStack.value.push(snapshot)
    if (undoStack.value.length > MAX_HISTORY) {
      undoStack.value.shift()
    }
    
    // Clear redo stack when new action happens
    redoStack.value = []
  }

  function undo() {
    if (undoStack.value.length <= 1) return // Need at least 1 previous state to go back to (current state)
    
    const uiStore = useUiStore()
    const timelineStore = useTimelineStore()
    
    isExecuting.value = true
    
    // Pop current state and move to redo
    const currentState = undoStack.value.pop()
    redoStack.value.push(currentState)
    
    // Get previous state
    const previousState = undoStack.value[undoStack.value.length - 1]
    
    // Apply previous state
    timelineStore.restoreSnapshot(previousState)
    
    uiStore.addToast('Undo', 'info', 1000)
    
    // Allow reactivity to settle
    setTimeout(() => {
      isExecuting.value = false
    }, 50)
  }

  function redo() {
    if (redoStack.value.length === 0) return
    
    const uiStore = useUiStore()
    const timelineStore = useTimelineStore()
    
    isExecuting.value = true
    
    // Pop from redo and move to undo
    const nextState = redoStack.value.pop()
    undoStack.value.push(nextState)
    
    // Apply next state
    timelineStore.restoreSnapshot(nextState)
    
    uiStore.addToast('Redo', 'info', 1000)
    
    // Allow reactivity to settle
    setTimeout(() => {
      isExecuting.value = false
    }, 50)
  }

  function clearHistory() {
    undoStack.value = []
    redoStack.value = []
    isExecuting.value = false
  }

  return {
    undoStack,
    redoStack,
    pushSnapshot,
    undo,
    redo,
    clearHistory
  }
})
