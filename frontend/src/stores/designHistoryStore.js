import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useDesignStore } from './designStore'
import { useUiStore } from './uiStore'

export const useDesignHistoryStore = defineStore('designHistory', () => {
  const undoStack = ref([])
  const redoStack = ref([])
  const isExecuting = ref(false)
  const MAX_HISTORY = 50

  function pushSnapshot() {
    if (isExecuting.value) return

    const designStore = useDesignStore()
    const snapshot = designStore.serialize()

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

    redoStack.value = []
  }

  function undo() {
    if (undoStack.value.length <= 1) return

    const uiStore = useUiStore()
    const designStore = useDesignStore()

    isExecuting.value = true

    const currentState = undoStack.value.pop()
    redoStack.value.push(currentState)

    const previousState = undoStack.value[undoStack.value.length - 1]
    designStore.deserialize(previousState)

    uiStore.addToast('Undo', 'info', 1000)

    setTimeout(() => {
      isExecuting.value = false
    }, 50)
  }

  function redo() {
    if (redoStack.value.length === 0) return

    const uiStore = useUiStore()
    const designStore = useDesignStore()

    isExecuting.value = true

    const nextState = redoStack.value.pop()
    undoStack.value.push(nextState)

    designStore.deserialize(nextState)

    uiStore.addToast('Redo', 'info', 1000)

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
    clearHistory,
  }
})
