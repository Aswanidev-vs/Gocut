<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useProjectStore } from './stores/projectStore'
import { useUiStore } from './stores/uiStore'
import { useTimelineStore } from './stores/timelineStore'
import { useHistoryStore } from './stores/historyStore'
import { OpenFilePicker } from './lib/wails'
import { Film, FolderOpen, FilePlus, Sparkles, Github, Scissors, Palette, Headphones, X } from 'lucide-vue-next'
import { useHotkeys } from './composables/useHotkeys'
import { useDesignHistoryStore } from './stores/designHistoryStore'

import TopBar from './components/layout/TopBar.vue'
import LeftPanel from './components/layout/LeftPanel.vue'
import RightPanel from './components/layout/RightPanel.vue'
import TimelinePanel from './components/timeline/TimelinePanel.vue'
import PreviewPlayer from './components/preview/PreviewPlayer.vue'
import NewProjectDialog from './components/common/NewProjectDialog.vue'
import ExportDialog from './components/export/ExportDialog.vue'
import ToastContainer from './components/common/ToastContainer.vue'
import DesignWorkspace from './components/design/DesignWorkspace.vue'

const projectStore = useProjectStore()
const uiStore = useUiStore()
const timelineStore = useTimelineStore()
const historyStore = useHistoryStore()
const designHistoryStore = useDesignHistoryStore()

useHotkeys()

const isLoaded = ref(false)

onMounted(() => {
  projectStore.fetchRecentProjects().catch(() => {})
  setTimeout(() => { isLoaded.value = true }, 80)

  // Global keyboard shortcuts.
  window.addEventListener('keydown', onKeyDown)
  window.addEventListener('beforeunload', onBeforeUnload)
  window.addEventListener('shortcut:export', onShortcutExport)
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  window.removeEventListener('beforeunload', onBeforeUnload)
  window.removeEventListener('shortcut:export', onShortcutExport)
})

watch(() => projectStore.project?.id, () => {
  historyStore.clearHistory()
  designHistoryStore.clearHistory()
  if (projectStore.hasProject) {
    historyStore.pushSnapshot()
  }
}, { immediate: true })

function onBeforeUnload(e) {
  if (projectStore.hasProject && projectStore.isDirty) {
    projectStore.flushAutosave().catch(() => {})
    e.preventDefault()
    e.returnValue = ''
  }
}

function onKeyDown(e) {
  if (!projectStore.hasProject) return
  if (['INPUT', 'TEXTAREA', 'SELECT'].includes(e.target.tagName)) return
  // Space — toggle playback
  if (e.code === 'Space') {
    e.preventDefault()
    import('./stores/playerStore').then(({ usePlayerStore }) => {
      usePlayerStore().togglePlay()
    })
  }
  // Delete / Backspace — delete selected
  if (e.code === 'Delete' || e.code === 'Backspace') {
    e.preventDefault()
    timelineStore.removeSelected()
  }
  // S — toggle snap
  if (e.key === 's' || e.key === 'S') {
    timelineStore.snapEnabled = !timelineStore.snapEnabled
    uiStore.addToast(`Snap ${timelineStore.snapEnabled ? 'on' : 'off'}`, 'info', 1200)
  }
  // C — razor tool, V / Escape — back to select tool
  if ((e.key === 'c' || e.key === 'C') && !e.ctrlKey && !e.metaKey) {
    e.preventDefault()
    timelineStore.setActiveTool('razor')
    uiStore.addToast('Razor tool', 'info', 1200)
  }
  if ((e.key === 'v' || e.key === 'V' || e.key === 'Escape') && !e.ctrlKey && !e.metaKey) {
    if (timelineStore.activeTool !== 'select') {
      timelineStore.setActiveTool('select')
      uiStore.addToast('Select tool', 'info', 1200)
    }
  }
  // Ctrl+S — save
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    projectStore.saveProject().then(() => {
      uiStore.addToast('Saved', 'success', 1200)
    }).catch(err => uiStore.addToast('Save failed: ' + err, 'error'))
  }
  // Ctrl+E — export dialog
  if ((e.ctrlKey || e.metaKey) && e.key === 'e') {
    e.preventDefault()
    uiStore.isExportDialogOpen = true
  }
}

// Listen for shortcut:export event from useHotkeys
function onShortcutExport() {
  uiStore.isExportDialogOpen = true
}

function openNewProjectDialog() {
  uiStore.isNewProjectDialogOpen = true
}

async function openProject(selectedPath) {
  try {
    let projectPath = selectedPath
    if (!projectPath) {
      const paths = await OpenFilePicker([{ name: 'Gocut Project', extensions: ['gocut', 'json'] }])
      if (!Array.isArray(paths) || paths.length === 0) return
      projectPath = paths[0]
    }
    await projectStore.loadProject(projectPath)
    uiStore.addToast('Project loaded', 'success', 1500)
  } catch (error) {
    uiStore.addToast('Failed to open: ' + (error?.message || error), 'error')
  }
}

async function deleteRecent(id) {
  try {
    await projectStore.deleteRecentProject(id)
    uiStore.addToast('Removed from recents', 'success', 1500)
  } catch (err) {
    uiStore.addToast('Failed to remove: ' + (err?.message || err), 'error')
  }
}

async function clearRecent() {
  try {
    await projectStore.clearRecentProjects()
    uiStore.addToast('Recent projects cleared', 'success', 1500)
  } catch (err) {
    uiStore.addToast('Failed to clear: ' + (err?.message || err), 'error')
  }
}

const stats = computed(() => ({
  assets: projectStore.project?.assets?.length || 0,
  clips: timelineStore.clips.length,
  tracks: timelineStore.tracks.length,
}))

const leftWidth = ref(260)
const rightWidth = ref(300)
const bottomHeight = ref(256)

let activeDrag = null
function startDrag(e, panel) {
  activeDrag = panel
  document.addEventListener('mousemove', onDrag)
  document.addEventListener('mouseup', stopDrag)
  // Prevent text selection during drag
  e.preventDefault()
}

function onDrag(e) {
  if (!activeDrag) return
  if (activeDrag === 'left') {
    const isDesign = uiStore.activeWorkspace === 'design'
    const maxX = isDesign ? window.innerWidth - 400 : window.innerWidth - rightWidth.value - 200
    leftWidth.value = Math.max(150, Math.min(e.clientX, maxX))
  } else if (activeDrag === 'right') {
    rightWidth.value = Math.max(150, Math.min(window.innerWidth - e.clientX, window.innerWidth - leftWidth.value - 200))
  } else if (activeDrag === 'bottom') {
    // topBar is usually 48px, but we can just use bounding rects if needed.
    // e.clientY is absolute. The timeline is at the bottom.
    // bottom height = window.innerHeight - e.clientY
    bottomHeight.value = Math.max(100, Math.min(window.innerHeight - e.clientY, window.innerHeight - 150))
  }
}

function stopDrag() {
  activeDrag = null
  document.removeEventListener('mousemove', onDrag)
  document.removeEventListener('mouseup', stopDrag)
}
</script>

<template>
  <div class="flex flex-col h-screen w-screen bg-bg text-text-primary font-dm-sans overflow-hidden select-none">
    <!-- Loading -->
    <div v-if="!isLoaded" class="flex-1 flex items-center justify-center">
      <div class="flex items-center gap-2 text-text-secondary text-sm">
        <div class="w-4 h-4 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        Loading…
      </div>
    </div>

    <!-- Home screen (no project) -->
    <div v-else-if="!projectStore.hasProject" class="flex-1 flex flex-col">
      <!-- <div class="h-12 px-4 flex items-center border-b border-border bg-panel">
        <div class="flex items-center gap-2">
          <img src="./assets/images/logo-universal.png" alt="Gocut Logo" class="w-6 h-6 rounded" />
          <span class="text-sm font-semibold">Gocut</span>
          <span class="text-[10px] text-text-secondary px-1.5 py-0.5 rounded border border-border font-mono">v0.1.0</span>
        </div>
        <div class="flex-1" />
       
      </div> -->
       <!-- <a href="https://github.com/Aswanidev-vs/Gocut" target="_blank" rel="noopener">
          <Github :size="14" />
        </a> -->
        
      <div class="flex-1 flex flex-col items-center justify-center px-6 text-center">
        <div class="relative mb-6">
          <img src="./assets/images/logo-universal.png" alt="Gocut Logo" class="w-20 h-20 rounded-2xl shadow-2xl shadow-accent/20 object-contain bg-accent/10" />
        </div>
        <h1 class="text-3xl font-bold text-text-primary mb-1 tracking-tight">Gocut</h1>
        <p class="text-sm text-text-secondary mb-1">Open-source, offline video editor.</p>
        <p class="text-xs text-text-secondary mb-8 max-w-md">A privacy-respecting, fully local CapCut alternative for Windows, macOS, and Linux. Powered by FFmpeg.</p>

        <div class="flex items-center gap-2">
          <button
            class="flex items-center gap-2 px-4 py-2 rounded bg-accent text-bg text-sm font-medium hover:bg-accent-hover transition-colors shadow-lg shadow-accent/20"
            @click="openNewProjectDialog"
          >
            <FilePlus :size="16" /> New Project
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2 rounded border border-border text-sm text-text-primary hover:bg-border transition-colors"
            @click="openProject()"
          >
            <FolderOpen :size="16" /> Open Project
          </button>
        </div>

        <div v-if="projectStore.recentProjects.length" class="mt-10 w-full max-w-2xl">
          <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2 text-[11px] text-text-secondary uppercase tracking-wider">
              <Film :size="12" />
              Recent
            </div>
            <button
              class="text-[10px] text-text-secondary hover:text-red-400 transition-colors uppercase tracking-wider font-semibold"
              @click="clearRecent"
            >
              Clear All
            </button>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
            <div
              v-for="rp in projectStore.recentProjects.slice(0, 6)"
              :key="rp.id || rp.path"
              class="group relative flex items-center gap-3 px-3 py-3 rounded bg-panel border border-border hover:border-accent/50 hover:bg-panel/80 transition-colors text-left cursor-pointer"
              @click="openProject(rp.path)"
            >
              <div class="w-10 h-10 rounded bg-gradient-to-br from-accent/20 to-accent/5 flex items-center justify-center flex-shrink-0">
                <Film :size="16" class="text-accent" />
              </div>
              <div class="flex-1 min-w-0 pr-4">
                <div class="text-sm text-text-primary truncate">{{ rp.name }}</div>
                <div class="text-[10px] text-text-secondary truncate" :title="rp.path">{{ rp.path }}</div>
              </div>
              <button
                class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded hover:bg-red-500/20 text-text-secondary hover:text-red-400 opacity-0 group-hover:opacity-100 transition-all z-10"
                @click.stop="deleteRecent(rp.id || rp.path)"
                title="Remove from recents"
              >
                <X :size="12" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Editor view -->
    <div v-else class="flex-1 flex flex-col overflow-hidden">
      <TopBar />
      <div class="flex-1 flex overflow-hidden">
        <LeftPanel :style="{ width: leftWidth + 'px' }" />
        
        <div class="w-1 cursor-col-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'left')" />

        <div class="flex-1 flex flex-col overflow-hidden bg-bg">
          <template v-if="uiStore.activeWorkspace === 'design'">
            <DesignWorkspace class="flex-1" />
          </template>
          <template v-else>
            <PreviewPlayer class="flex-1" />

            <div class="h-1 cursor-row-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'bottom')" />

            <TimelinePanel :style="{ height: bottomHeight + 'px' }" />
          </template>
        </div>
        
        <template v-if="uiStore.activeWorkspace !== 'design'">
          <div class="w-1 cursor-col-resize hover:bg-accent/50 active:bg-accent/80 z-20 transition-colors" @mousedown="(e) => startDrag(e, 'right')" />

          <RightPanel :style="{ width: rightWidth + 'px' }" />
        </template>
      </div>

      <!-- DaVinci Resolve-style Workspace Bar -->
      <div class="h-8 bg-panel border-t border-border flex items-center justify-center gap-1 flex-shrink-0 px-4">
        <button
          v-for="ws in [
            { id: 'edit', label: 'Edit', icon: 'Scissors' },
            { id: 'design', label: 'Design', icon: 'Sparkles' },
            { id: 'color', label: 'Color', icon: 'Palette' },
            { id: 'audio', label: 'Audio', icon: 'Headphones' },
          ]"
          :key="ws.id"
          class="flex items-center gap-1.5 px-3 py-1 rounded text-[11px] font-medium transition-all"
          :class="uiStore.activeWorkspace === ws.id
            ? 'bg-accent/15 text-accent border border-accent/30'
            : 'text-text-secondary hover:text-text-primary hover:bg-border/60 border border-transparent'"
          @click="uiStore.setActiveWorkspace(ws.id)"
        >
          <component :is="ws.icon === 'Scissors' ? Scissors : ws.icon === 'Palette' ? Palette : ws.icon === 'Sparkles' ? Sparkles : Headphones" :size="12" />
          {{ ws.label }}
        </button>
      </div>
    </div>

    <NewProjectDialog />
    <ExportDialog :is-open="uiStore.isExportDialogOpen" @close="uiStore.isExportDialogOpen = false" />
    <ToastContainer />
  </div>
</template>
