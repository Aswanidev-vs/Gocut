import { defineStore, getActivePinia } from 'pinia'
import { ref, computed } from 'vue'
import {
  NewProject,
  SaveProject,
  LoadProject,
  GetRecentProjects,
  ImportMedia,
  ExtractThumbnail,
  ExtractWaveform,
  CheckFFmpegInstalled,
  OpenFilePicker,
  OpenDirectoryPicker,
  SaveFilePicker,
  DeleteProject,
  ClearRecentProjects,
} from '../lib/wails'

// IMPORTANT: useTimelineStore is NOT imported here. Doing so creates a
// circular dependency (timelineStore imports useProjectStore) that
// throws "getActivePinia was called with no active Pinia" during the
// first setProject call. We resolve it lazily through the pinia registry.

const generateId = () => crypto.randomUUID()

function getTimelineStoreSafely() {
  try {
    const pinia = getActivePinia()
    if (!pinia) return null
    return pinia._s && pinia._s.get('timeline') ? pinia._s.get('timeline') : null
  } catch (_) {
    return null
  }
}

function hydrateTimelineAsync(p) {
  // Defer to a microtask so we never call into another store while still
  // inside a store's setup function.
  Promise.resolve().then(() => {
    try {
      const ts = getTimelineStoreSafely()
      if (!ts) return
      if (p?.timeline?.tracks?.length) {
        ts.loadFromProject(p)
      } else {
        ts.clearAll()
      }
    } catch (_) { /* ignore */ }
  })
}

export const useProjectStore = defineStore('project', () => {
  const project = ref(null)
  const isDirty = ref(false)
  const recentProjects = ref([])
  const isLoading = ref(false)
  const error = ref(null)
  const ffmpegVersion = ref(null)
  const ffmpegStatus = ref('unknown')
  const ffmpegError = ref('')
  let autosaveTimer = null

  const hasProject = computed(() => project.value !== null)
  const projectName = computed(() => project.value?.name ?? 'Untitled')

  function markDirty() {
    if (project.value) {
      isDirty.value = true
      scheduleAutosave()
    }
  }

  function clearDirty() {
    isDirty.value = false
  }

  function buildProjectPayload() {
    if (!project.value) return null

    const payload = JSON.parse(JSON.stringify(project.value))
    const ts = getTimelineStoreSafely()
    if (ts) {
      payload.timeline = {
        tracks: ts.tracks.map(t => ({
          ...t,
          clips: ts.clips
            .filter(c => c.trackId === t.id)
            .map(c => ({ ...c })),
        })),
        duration: ts.duration,
      }
    }
    payload.assets = payload.assets || []
    return payload
  }

  async function persistProjectSnapshot({ clearDirtyOnSuccess = false } = {}) {
    const payload = buildProjectPayload()
    if (!payload) return
    await SaveProject(payload)
    if (clearDirtyOnSuccess) {
      clearDirty()
    }
  }

  function scheduleAutosave() {
    if (!project.value || !isDirty.value) return
    if (autosaveTimer) {
      clearTimeout(autosaveTimer)
    }
    autosaveTimer = setTimeout(async () => {
      autosaveTimer = null
      try {
        await persistProjectSnapshot()
      } catch (_) {
        // Best-effort autosave: explicit saves still surface errors.
      }
    }, 1200)
  }

  function setProject(p) {
    project.value = p
    isDirty.value = false
    if (autosaveTimer) {
      clearTimeout(autosaveTimer)
      autosaveTimer = null
    }
    hydrateTimelineAsync(p)
  }

  async function createProject({
    name = 'Untitled',
    aspectRatio = '16:9',
    resolution = { width: 1920, height: 1080 },
    fps = 30,
  } = {}) {
    isLoading.value = true
    error.value = null
    try {
      const p = await NewProject({
        name,
        aspectRatio,
        resolution,
        fps,
        backgroundColor: '#000000',
        autoSave: true,
        autoSaveIntervalSeconds: 60,
      })
      setProject(p)
      return p
    } catch (e) {
      error.value = e
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function saveProject() {
    if (!project.value) return
    isLoading.value = true
    try {
      if (!project.value.filePath) {
        const defaultName = (project.value.name || 'Untitled') + '.gocut'
        const path = await SaveFilePicker(defaultName, [{ name: 'Gocut Project', extensions: ['gocut'] }])
        if (!path) {
          throw new Error('Save cancelled')
        }
        project.value.filePath = path
      }
      await persistProjectSnapshot({ clearDirtyOnSuccess: true })
    } catch (e) {
      error.value = e
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function loadProject(path) {
    isLoading.value = true
    error.value = null
    try {
      const p = await LoadProject(path)
      setProject(p)
      return p
    } catch (e) {
      error.value = e
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function fetchRecentProjects() {
    try {
      const list = await GetRecentProjects()
      recentProjects.value = Array.isArray(list) ? list : []
    } catch (e) {
      error.value = e
    }
  }

  async function updateProjectName(name) {
    if (!project.value) return
    project.value.name = name
    if (project.value.filePath) {
      const lastSlash = Math.max(project.value.filePath.lastIndexOf('/'), project.value.filePath.lastIndexOf('\\'))
      if (lastSlash !== -1) {
        const dir = project.value.filePath.substring(0, lastSlash)
        project.value.filePath = `${dir}/${name}.gocut`
      } else {
        project.value.filePath = `${name}.gocut`
      }
    }
    markDirty()
  }

  function addAsset(asset) {
    if (!project.value) return
    project.value.assets = project.value.assets || []
    project.value.assets.push(asset)
    markDirty()
  }

  function removeAsset(assetId) {
    if (!project.value) return
    project.value.assets = (project.value.assets || []).filter(a => a.id !== assetId)
    markDirty()
  }

  async function importMedia(paths) {
    if (!project.value) return []
    isLoading.value = true
    try {
      const activeProject = project.value
      activeProject.assets = activeProject.assets || []
      const existingPaths = new Set(activeProject.assets.map(a => a.path))

      const toImport = paths.filter(p => !existingPaths.has(p))
      if (toImport.length === 0) return []

      const assets = await ImportMedia(toImport)
      const enriched = assets.map(asset => ({
        ...asset,
        id: asset.id || generateId(),
        thumbnail: '',
        waveform: [],
      }))

      activeProject.assets.push(...enriched)

      if (!activeProject.filePath && enriched.length > 0) {
        const firstPath = enriched[0].path
        const lastSlash = Math.max(firstPath.lastIndexOf('/'), firstPath.lastIndexOf('\\'))
        if (lastSlash !== -1) {
          const dir = firstPath.substring(0, lastSlash)
          activeProject.filePath = `${dir}/${activeProject.name || 'Untitled'}.gocut`
        }
      }

      markDirty()

      for (const asset of enriched) {
        if (asset.type === 'video' || asset.type === 'image') {
          ExtractThumbnail(asset.id, 1000).then((b64) => {
            asset.thumbnail = b64
            markDirty()
          }).catch(() => {})
        }
        if (asset.type === 'video' || asset.type === 'audio') {
          ExtractWaveform(asset.id).then((wf) => {
            asset.waveform = wf
            markDirty()
          }).catch(() => {})
        }
      }

      return enriched
    } catch (e) {
      error.value = e
      throw e
    } finally {
      isLoading.value = false
    }
  }

  async function checkFFmpeg() {
    try {
      const info = await CheckFFmpegInstalled()
      ffmpegVersion.value = info
      ffmpegStatus.value = 'ready'
      ffmpegError.value = ''
      return info
    } catch (e) {
      ffmpegVersion.value = null
      ffmpegStatus.value = 'missing'
      ffmpegError.value = e?.message || String(e)
      error.value = e
      return null
    }
  }

  function getAsset(assetId) {
    return project.value?.assets?.find(a => a.id === assetId) || null
  }

  function updateLocalProject(updater) {
    if (!project.value) return
    updater(project.value)
    markDirty()
  }

  async function selectSaveDirectory() {
    if (!project.value) return
    try {
      const dir = await OpenDirectoryPicker()
      if (dir) {
        project.value.customSaveDirectory = dir
        project.value.filePath = `${dir}/${project.value.name || 'Untitled'}.gocut`
        markDirty()
      }
    } catch (e) {
      error.value = e
      throw e
    }
  }

  function closeProject() {
    if (autosaveTimer) {
      clearTimeout(autosaveTimer)
      autosaveTimer = null
    }
    project.value = null
    isDirty.value = false
    const ts = getTimelineStoreSafely()
    if (ts) ts.clearAll()
  }

  function flushAutosave() {
    if (autosaveTimer) {
      clearTimeout(autosaveTimer)
      autosaveTimer = null
    }
    return persistProjectSnapshot()
  }

  async function deleteRecentProject(id) {
    try {
      await DeleteProject(id)
      await fetchRecentProjects()
    } catch (e) {
      error.value = e
      throw e
    }
  }

  async function clearRecentProjects() {
    try {
      await ClearRecentProjects()
      recentProjects.value = []
    } catch (e) {
      error.value = e
      throw e
    }
  }

  return {
    project,
    isDirty,
    recentProjects,
    isLoading,
    error,
    ffmpegVersion,
    ffmpegStatus,
    ffmpegError,
    hasProject,
    projectName,
    createProject,
    saveProject,
    loadProject,
    fetchRecentProjects,
    updateProjectName,
    addAsset,
    removeAsset,
    importMedia,
    checkFFmpeg,
    getAsset,
    updateLocalProject,
    closeProject,
    selectSaveDirectory,
    markDirty,
    clearDirty,
    setProject,
    flushAutosave,
    deleteRecentProject,
    clearRecentProjects,
  }
})
