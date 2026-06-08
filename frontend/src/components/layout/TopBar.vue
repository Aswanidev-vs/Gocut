<script setup>
import { ref, computed, onMounted } from 'vue'
import { useProjectStore } from '../../stores/projectStore'
import { useUiStore } from '../../stores/uiStore'
import { Undo2, Redo2, Download, Save, FolderOpen, FilePlus, Settings as SettingsIcon, Loader2, CheckCircle2, AlertCircle } from 'lucide-vue-next'
import SettingsDialog from '../settings/SettingsDialog.vue'

const projectStore = useProjectStore()
const uiStore = useUiStore()

const isEditingName = ref(false)
const nameInput = ref(projectStore.projectName)

function startRename() {
  nameInput.value = projectStore.projectName
  isEditingName.value = true
}
function commitRename() {
  if (nameInput.value.trim()) {
    projectStore.updateProjectName(nameInput.value.trim())
  }
  isEditingName.value = false
}
function cancelRename() {
  isEditingName.value = false
}

const ffmpegLabel = computed(() => {
  if (projectStore.ffmpegStatus === 'ready') {
    return 'ffmpeg ready'
  }
  if (projectStore.ffmpegError) {
    if (projectStore.ffmpegError.includes('ffprobe not found')) return 'ffprobe missing'
    if (projectStore.ffmpegError.includes('ffmpeg not found')) return 'ffmpeg missing'
    return 'ffmpeg unavailable'
  }
  return 'checking ffmpeg'
})

const ffmpegTone = computed(() => (
  projectStore.ffmpegStatus === 'ready' ? 'ok' : 'warn'
))

const ffmpegTitle = computed(() => {
  if (projectStore.ffmpegStatus === 'ready') {
    return projectStore.ffmpegVersion || 'ffmpeg available'
  }
  if (projectStore.ffmpegError.includes('ffprobe not found')) {
    return 'ffprobe was not found in PATH. Install the full FFmpeg package or add its bin folder to PATH.'
  }
  if (projectStore.ffmpegError.includes('ffmpeg not found')) {
    return 'ffmpeg was not found in PATH. Install FFmpeg and add its bin folder to PATH.'
  }
  return projectStore.ffmpegError || 'Checking FFmpeg availability'
})

onMounted(() => {
  projectStore.checkFFmpeg().catch(() => {})
})

async function onSave() {
  try {
    await projectStore.saveProject()
    uiStore.addToast('Project saved', 'success', 2000)
  } catch (e) {
    uiStore.addToast('Failed to save: ' + (e?.message || e), 'error')
  }

}

function openExport() {
  uiStore.isExportDialogOpen = true
}

function openSettings() {
  uiStore.isSettingsDialogOpen = true
}
</script>

<template>
  <div class="h-11 bg-panel border-b border-border flex items-center px-2 gap-1 select-none flex-shrink-0">
    <!-- Logo / Home -->
    <button
      class="flex items-center gap-1.5 px-2 py-1 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="projectStore.closeProject()"
      title="Back to home"
    >
      <!-- <div class="w-5 h-5 rounded bg-accent text-bg font-bold text-[11px] flex items-center justify-center">G</div> -->
      <img src="../../assets/images/logo-universal.png" alt="Gocut Logo" class="w-5 h-5 rounded" />
    </button>

    <div class="h-5 w-px bg-border mx-1" />

    <!-- Project name -->
    <div class="flex items-center min-w-0 max-w-[280px]">
      <template v-if="isEditingName">
        <input
          v-model="nameInput"
          class="bg-bg border border-accent rounded px-2 py-1 text-sm text-text-primary outline-none w-full"
          @keyup.enter="commitRename"
          @keyup.esc="cancelRename"
          @blur="commitRename"
          autofocus
        />
      </template>
      <button
        v-else
        class="px-2 py-1 rounded text-sm text-text-primary hover:bg-border transition-colors truncate max-w-[260px]"
        @dblclick="startRename"
        @click="startRename"
        :title="projectStore.projectName"
      >
        {{ projectStore.projectName }}
        <span v-if="projectStore.isDirty" class="text-accent">•</span>
      </button>
    </div>

    <div class="h-5 w-px bg-border mx-1" />

    <!-- Undo / Redo (placeholders) -->
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      title="Undo (Ctrl+Z)"
    >
      <Undo2 :size="14" />
    </button>
    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      title="Redo (Ctrl+Y)"
    >
      <Redo2 :size="14" />
    </button>

    <div class="flex-1" />

    <!-- FFmpeg status -->
    <div
      class="flex items-center gap-1.5 px-2 py-1 rounded text-xs transition-colors"
      :class="ffmpegTone === 'ok'
        ? 'text-text-secondary'
        : 'bg-red-400/10 text-red-400 border border-red-400/20'"
      :title="ffmpegTitle"
    >
      <CheckCircle2 v-if="ffmpegTone === 'ok'" :size="12" />
      <AlertCircle v-else :size="12" />
      <span class="font-mono">{{ ffmpegLabel }}</span>
    </div>

    <div class="h-5 w-px bg-border mx-1" />

    <button
      class="flex items-center gap-1.5 px-2.5 py-1.5 rounded text-xs text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="onSave"
      title="Save (Ctrl+S)"
    >
      <Save :size="14" />
      Save
    </button>

    <button
      class="flex items-center gap-1.5 px-3 py-1.5 rounded bg-accent text-bg text-xs font-medium hover:bg-accent-hover transition-colors"
      @click="openExport"
      title="Export"
    >
      <Download :size="14" />
      Export
    </button>

    <button
      class="p-1.5 rounded text-text-secondary hover:text-text-primary hover:bg-border transition-colors"
      @click="openSettings"
      title="Settings"
    >
      <SettingsIcon :size="14" />
    </button>
    <SettingsDialog />
  </div>
</template>
