import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const activePanelTab = ref('media')
  const activeInspectorTab = ref('edit')
  const activeWorkspace = ref('edit')
  const renderJobs = ref([])
  const toasts = ref([])
  const isExportDialogOpen = ref(false)
  const isNewProjectDialogOpen = ref(false)
  const isSettingsDialogOpen = ref(false)
  const snapIndicatorTime = ref(null)

  const panelTabs = [
    { id: 'media', label: 'Media', icon: 'Image' },
    { id: 'audio', label: 'Audio', icon: 'Music' },
    { id: 'text', label: 'Text', icon: 'Type' },
    { id: 'stickers', label: 'Stickers', icon: 'Smile' },
    { id: 'fx', label: 'FX', icon: 'Sparkles' },
    { id: 'transitions', label: 'Transitions', icon: 'ArrowRightLeft' },
  ]

  const inspectorTabs = [
    { id: 'transform', label: 'Transform' },
    { id: 'color', label: 'Color' },
    { id: 'audio', label: 'Audio' },
    { id: 'text', label: 'Text' },
  ]

  function addToast(message, type = 'info', duration = 4000) {
    const toast = {
      id: crypto.randomUUID(),
      message,
      type,
      duration,
      createdAt: Date.now(),
    }
    toasts.value.push(toast)
    if (toast.duration > 0) {
      setTimeout(() => {
        toasts.value = toasts.value.filter(t => t.id !== toast.id)
      }, toast.duration)
    }
  }

  function removeToast(toastId) {
    toasts.value = toasts.value.filter(t => t.id !== toastId)
  }

  function setActivePanelTab(tab) {
    activePanelTab.value = tab
  }

  function setActiveInspectorTab(tab) {
    activeInspectorTab.value = tab
  }

  function setActiveWorkspace(ws) {
    activeWorkspace.value = ws
    // When switching workspace, also switch inspector to the matching tab
    if (ws === 'color') activeInspectorTab.value = 'color'
    else if (ws === 'audio') activeInspectorTab.value = 'audio'
    else activeInspectorTab.value = 'edit'
  }

  function addRenderJob(job) {
    renderJobs.value.push({
      ...job,
      id: crypto.randomUUID(),
      progress: 0,
      status: 'queued',
    })
  }

  function updateRenderJob(jobId, updates) {
    const index = renderJobs.value.findIndex(j => j.id === jobId)
    if (index !== -1) {
      renderJobs.value[index] = { ...renderJobs.value[index], ...updates }
    }
  }

  function removeRenderJob(jobId) {
    renderJobs.value = renderJobs.value.filter(j => j.id !== jobId)
  }

  function setSnapIndicator(time) {
    snapIndicatorTime.value = time
    setTimeout(() => {
      snapIndicatorTime.value = null
    }, 200)
  }

  return {
    activePanelTab,
    activeInspectorTab,
    activeWorkspace,
    renderJobs,
    toasts,
    isExportDialogOpen,
    isNewProjectDialogOpen,
    isSettingsDialogOpen,
    snapIndicatorTime,
    panelTabs,
    inspectorTabs,
    addToast,
    removeToast,
    setActivePanelTab,
    setActiveInspectorTab,
    setActiveWorkspace,
    addRenderJob,
    updateRenderJob,
    removeRenderJob,
    setSnapIndicator,
  }
})
